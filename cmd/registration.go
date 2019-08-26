////////////////////////////////////////////////////////////////////////////////
// Copyright © 2018 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

// Handles creating client registration callbacks for hooking into comms library

package cmd

import (
	"crypto"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/comms/registration"
	"gitlab.com/elixxir/comms/utils"
	"gitlab.com/elixxir/crypto/signature/rsa"
	"gitlab.com/elixxir/crypto/tls"
	"gitlab.com/elixxir/registration/database"
	"io/ioutil"
	"strconv"
	"strings"
)

type RegistrationImpl struct {
	Comms             *registration.RegistrationComms
	permissioningCert *x509.Certificate
	permissioningKey  *rsa.PrivateKey
	ndfOutputPath     string
	completedNodes    chan struct{}
	NumNodesInNet     int
}

type Params struct {
	Address       string
	CertPath      string
	KeyPath       string
	NdfOutputPath string
	NumNodesInNet int
}

type connectionID string

func (c connectionID) String() string {
	return (string)(c)
}

// Configure and start the Permissioning Server
func StartRegistration(params Params) *RegistrationImpl {
	jww.DEBUG.Printf("Starting registration\n")
	regImpl := &RegistrationImpl{}

	var cert, key []byte
	var err error

	if !noTLS {
		// Read in TLS keys from files
		cert, err = ioutil.ReadFile(utils.GetFullPath(params.CertPath))
		if err != nil {
			jww.ERROR.Printf("failed to read certificate at %+v: %+v", params.CertPath, err)
		}

		// Set globals for permissioning server
		regImpl.permissioningCert, err = tls.LoadCertificate(string(cert))
		if err != nil {
			jww.ERROR.Printf("Failed to parse permissioning server cert: %+v. "+
				"Permissioning cert is %+v",
				err, regImpl.permissioningCert)
		}
		jww.DEBUG.Printf("permissioningCert: %+v\n", regImpl.permissioningCert)
		jww.DEBUG.Printf("permissioning public key: %+v\n", regImpl.permissioningCert.PublicKey)
		jww.DEBUG.Printf("permissioning private key: %+v\n", regImpl.permissioningKey)
	}
	regImpl.NumNodesInNet = len(RegistrationCodes)
	key, err = ioutil.ReadFile(utils.GetFullPath(params.KeyPath))
	if err != nil {
		jww.ERROR.Printf("failed to read key at %+v: %+v", params.KeyPath, err)
	}
	regImpl.permissioningKey, err = rsa.LoadPrivateKeyFromPem(key)
	if err != nil {
		jww.ERROR.Printf("Failed to parse permissioning server key: %+v. "+
			"PermissioningKey is %+v",
			err, regImpl.permissioningKey)
	}

	regImpl.ndfOutputPath = params.NdfOutputPath

	// Start the communication server
	regImpl.Comms = registration.StartRegistrationServer(params.Address,
		regImpl, cert, key)

	//TODO: change the buffer length to that set in params..also set in params :)
	regImpl.completedNodes = make(chan struct{}, regImpl.NumNodesInNet)
	return regImpl
}

// Handle registration attempt by a Client
func (m *RegistrationImpl) RegisterUser(registrationCode, pubKey string) (signature []byte, err error) {
	jww.INFO.Printf("Verifying for registration code %+v",
		registrationCode)
	// Check database to verify given registration code
	err = database.PermissioningDb.UseCode(registrationCode)
	if err != nil {
		// Invalid registration code, return an error
		errMsg := errors.New(fmt.Sprintf(
			"Error validating registration code: %+v", err))
		jww.ERROR.Printf("%+v", errMsg)
		return make([]byte, 0), errMsg
	}

	sha := crypto.SHA256

	// Use hardcoded keypair to sign Client-provided public key
	//Create a hash, hash the pubKey and then truncate it
	h := sha256.New()
	h.Write([]byte(pubKey))
	data := h.Sum(nil)
	sig, err := rsa.Sign(rand.Reader, m.permissioningKey, sha, data, nil)
	if err != nil {
		errMsg := errors.New(fmt.Sprintf(
			"unable to sign client public key: %+v", err))
		jww.ERROR.Printf("%+v", errMsg)
		return make([]byte, 0),
			err
	}

	jww.INFO.Printf("Verification complete for registration code %+v",
		registrationCode)
	// Return signed public key to Client with empty error field
	return sig, nil
}

// Handle client version check
// Example valid version strings:
// 0.1.0
// 1.3.0-ff81cdae
// Major and minor versions should both be numbers, and patch versions can be anything, but they must be present
func (m *RegistrationImpl) CheckClientVersion(versionString string) (isOK bool, err error) {
	version, err := parseClientVersion(versionString)
	if err != nil {
		return false, err
	}

	desiredVersionLock.RLock()
	defer desiredVersionLock.RUnlock()
	// Compare major version: must be equal to be deemed compatible
	if version.major != desiredVersion.major {
		return false, nil
	}
	// Compare minor version: version must be numerically greater than or equal desired version to be deemed compatible
	if version.minor < desiredVersion.minor {
		return false, nil
	}
	// Patch versions aren't supposed to affect compatibility, so they're ignored for the check

	return true, nil
}

func parseClientVersion(versionString string) (*clientVersion, error) {
	versions := strings.SplitN(versionString, ".", 3)
	if len(versions) != 3 {
		return nil, errors.New("Client version string must contain a major, minor, and patch version separated by \".\"")
	}
	major, err := strconv.Atoi(versions[0])
	if err != nil {
		return nil, errors.New("Major client version couldn't be parsed as integer")
	}
	minor, err := strconv.Atoi(versions[0])
	if err != nil {
		return nil, errors.New("Minor client version couldn't be parsed as integer")
	}
	return &clientVersion{
		major: major,
		minor: minor,
		patch: versions[2],
	}, nil
}

func setDesiredVersion(version *clientVersion) {
	desiredVersionLock.Lock()
	desiredVersion = version
	desiredVersionLock.Unlock()
}

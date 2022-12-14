# permissioning

Library containing the Permissioning Server for adding new clients and nodes to 
cMix

## Example Configuration File

```yaml
# ==================================
# Permissioning Server Configuration
# ==================================

# Log message level
logLevel: 1
# Path to log file
logPath: "registration.log"
# Path to the node topology permissioning info
ndfOutputPath: "ndf.json"
# Minimum number of nodes to begin running rounds. this differs from the number of members 
# in a team because some scheduling algorithms may require multiple teams worth of nodes at minimum
minimumNodes: 3

# Path to the file containing the round ID
roundIdPath: "roundId.txt"

# Path to the file containing the update ID
updateIdPath: "updateId.txt"

# UDB ID
udbID: 1

# Public address, used in NDF it gives to client
publicAdress: "0.0.0.0:11420"

# The listening port of this  server
port: 11420

# The minimum version required of gateways to connect
minGatewayVersion: "0.0.0"

# The minimum version required of servers to connect
minServerVersion:  "0.0.0"

# Disable pinging of Gateway public IP address.
disableGatewayPing: false

# Database connection information
dbUsername: "cmix"
dbPassword: ""
dbName: "cmix_server"
dbAddress: ""

# Path to JSON file with list of Node registration codes (in order of network 
# placement)
RegCodesFilePath: "regCodes.json"

# List of client codes to be added to the database (for testing)
clientRegCodes:
  - "AAAA"
  - "BBBB"
  - "CCCC"
    

# Client version (will allow all versions with major version 0)
clientVersion: "0.0.0"

# The duration between polling the disabled Node list for updates. Optional.
# Defaults to 1m.
disabledNodesPollDuration: 1m

# Path to the text file with a list of IDs of disabled Nodes. If no path is,
# supplied, then the disabled Node list polling never starts.
disabledNodesPath: "disabledNodes.txt"

# === REQUIRED FOR ENABLING TLS ===
# Path to the permissioning server private key file
keyPath: ""
# Path to the permissioning server certificate file
certPath: ""

# Time interval (in seconds) between committing Node statistics to storage
nodeMetricInterval: 180

# Time interval (in minutes) in which the database is 
# checked for banned nodes
BanTrackerInterval: "3"

# E2E/CMIX Primes
groups:
  cmix:
    prime: "${cmix_prime}"
    generator: "${cmix_generator}"
  e2e:
    prime: "${e2e_prime}"
    generator: "${e2e_generator}"

# Selection of scheduling algorithm to use. Options are:
#   simple - Schedules multiple teams to maximize performance, does not randomly re-arrange teams, if only a single
#            only scheduling a single team, will use numerical ordering data for AlphaNet
#   secure - Schedules new teams randomly, has appropriate buffers to ensure
# unpredictability, designed for BetaNet
schedulingAlgorithm: "single"

# Path to file with config for scheduling algorithm within the user directory 
schedulingConfigPath: "Scheduling_Simple_NonRandom.json"

# Time that the registration server waits before timing out while killing the round scheduling thread
schedulingKillTimeout: 10s
# Time the registration waits for rounds to close out and stop (optional)
closeTimeout: 60s
```

### SchedulingConfig template:
```json
{
  "TeamSize": 4,
  "BatchSize": 32,
  "RandomOrdering": false,
  "SemiOptimalOrdering": false,
  "MinimumDelay": 60,
  "RealtimeDelay": 3000,
  "Threshold":     10,
  "NodeCleanUpInterval": 3,  
  "Secure": 		     true,
  "RoundTimeout": 60
}
```

### RegCodes Template
```json
[{"RegCode": "qpol", "Order": "0"},
{"RegCode": "yiiq", "Order": "1"},
{"RegCode": "vydz", "Order": "2"},
{"RegCode": "gwxs", "Order": "3"},
{"RegCode": "nahv", "Order": "4"},
{"RegCode": "plmd", "Order": "5"}]
```

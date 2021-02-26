module gitlab.com/elixxir/registration

go 1.13

require (
	github.com/armon/consul-api v0.0.0-20180202201655-eb2c6b5be1b6 // indirect
	github.com/denisenkom/go-mssqldb v0.0.0-20200428022330-06a60b6afbbc // indirect
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-sql-driver/mysql v1.5.0 // indirect
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3
	github.com/jinzhu/gorm v1.9.12
	github.com/jinzhu/now v1.1.1 // indirect
	github.com/lib/pq v1.5.2 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/smartystreets/assertions v1.1.0 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/jwalterweatherman v1.1.0
	github.com/spf13/viper v1.7.1
	github.com/ugorji/go v1.1.4 // indirect
	github.com/xordataexchange/crypt v0.0.3-0.20170626215501-b2862e3d0a77 // indirect
	gitlab.com/elixxir/client v1.2.1-0.20210222224029-4300043d7ce8
	gitlab.com/elixxir/comms v0.0.4-0.20210226175832-0cf63a696bf9
	gitlab.com/elixxir/crypto v0.0.7-0.20210226175801-f490fc89ffdd
	gitlab.com/elixxir/primitives v0.0.3-0.20210226175744-d424cb7261fd
	gitlab.com/xx_network/comms v0.0.4-0.20210226175738-04b6c562dd2f
	gitlab.com/xx_network/crypto v0.0.5-0.20210226175725-80576a407b2d
	gitlab.com/xx_network/primitives v0.0.4-0.20210226175628-2b2742ebb772
)

replace google.golang.org/grpc => github.com/grpc/grpc-go v1.27.1

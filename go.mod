module github.com/irmf/reflector.go

replace github.com/btcsuite/btcd => github.com/lbryio/lbrycrd.go v0.0.0-20200203050410-e1076f12bf19

require (
	github.com/armon/go-metrics v0.0.0-20190430140413-ec5e00d3c878 // indirect
	github.com/aws/aws-sdk-go v1.27.0
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/davecgh/go-spew v1.1.1
	github.com/friendsofgo/errors v0.9.2 // indirect
	github.com/go-errors/errors v1.1.1 // indirect
	github.com/go-ini/ini v1.62.0 // indirect
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/golang/protobuf v1.4.3
	github.com/google/gops v0.3.7
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/hashicorp/golang-lru v0.5.4
	github.com/hashicorp/memberlist v0.1.4 // indirect
	github.com/hashicorp/serf v0.8.2
	github.com/inconshreveable/go-update v0.0.0-20160112193335-8152e7eb6ccf
	github.com/irmf/chainquery v1.9.1-0.20210213022256-c00cc8714fb6
	github.com/johntdyer/slackrus v0.0.0-20180518184837-f7aae3243a07
	github.com/karrick/godirwalk v1.16.1
	github.com/lbryio/lbry.go/v2 v2.6.1-0.20200901175808-73382bb02128
	github.com/lbryio/types v0.0.0-20201019032447-f0b4476ef386
	github.com/lucas-clemente/quic-go v0.18.1
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/phayes/freeport v0.0.0-20171002185219-e27662a4a9d6
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.4.0
	github.com/volatiletech/inflect v0.0.1 // indirect
	github.com/volatiletech/null v8.0.0+incompatible
	github.com/volatiletech/sqlboiler v3.7.1+incompatible // indirect
	go.uber.org/atomic v1.5.1
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777 // indirect
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	golang.org/x/text v0.3.5 // indirect
	golang.org/x/tools v0.0.0-20200825202427-b303f430e36d // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
)

go 1.15

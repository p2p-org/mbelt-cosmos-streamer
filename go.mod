module github.com/p2p-org/mbelt-cosmos-streamer

go 1.15

require (
	github.com/cosmos/cosmos-sdk v0.41.4
	github.com/cosmos/gaia/v4 v4.0.4
	github.com/jinzhu/configor v1.2.1
	github.com/lib/pq v1.9.0
	github.com/prometheus/common v0.15.0
	github.com/segmentio/kafka-go v0.3.7
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	github.com/tendermint/tendermint v0.34.7
	google.golang.org/grpc v1.35.0
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

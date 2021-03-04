package config

type Config struct {
	LogLevel    string `env:"LOG_LEVEL" default:"info"`
	ChainID     string `env:"CHAIN_ID" default:"cosmoshub-3"`
	KafkaPrefix string `env:"KAFKA_PREFIX" required:"true"`
	Node        struct {
		Host          string `required:"true" env:"NODE_HOST"`
		WebSocketPort int    `default:"26657" env:"NODE_WS_PORT"`
		LCDPort       int    `default:"1317" env:"NODE_LCD_PORT"`
		RPCPort       int    `default:"26657" env:"NODE_RPC_PORT"`
		GRPCPort      int    `default:"9090" env:"NODE_RPC_PORT"`
	}
	KafkaHost string `required:"true" env:"KAFKA_HOST"`
	PgUrl     string `required:"true" env:"PG_URL"`

	Watcher struct {
		Worker      int   `env:"WATCHER_WORKER" default:"-1"`
		StartHeight int64 `env:"WATCHER_START_HEIGHT" default:"-1"`
	}

	Counts struct {
		CheckBlocksConsistencyPerQuery int64 `env:"COUNT_BLOCKS_CONSISTENCY" default:"500"`
	}
}

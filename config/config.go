package config

type Config struct {
	LogLevel string `env:"LOG_LEVEL"`
	ChainID  string `env:"CHAIN_ID" default:"cosmoshub-3"`
	Node     struct {
		Host          string `required:"true" env:"NODE_HOST"`
		WebSocketPort int    `default:"26657" env:"NODE_WS_PORT"`
		LCDPort       int    `default:"1317" env:"NODE_LCD_PORT"`
		RPCPort       int    `default:"26657" env:"NODE_RPC_PORT"`
	}
	KafkaHosts []string `required:"true" env:"KAFKA_HOSTS"`
	PgUrl      string   `required:"true" env:"PG_URL"`

	Watcher struct {
		Worker      int   `env:"WATCHER_WORKER" default:"-1"`
		StartHeight int64 `env:"WATCHER_START_HEIGHT" default:"-1"`
	}
}

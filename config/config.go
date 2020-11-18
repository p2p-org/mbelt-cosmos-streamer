package config

type Config struct {
	Node struct {
		Host          string `required:"true" env:"NODE_HOST"`
		WebSocketPort int    `default:"26657" env:"NODE_WS_PORT"`
		LCDPort       int    `default:"1317" env:"NODE_LCD_PORT"`
		RPCPort       int    `default:"26657" env:"NODE_RPC_PORT"`
	}
	KafkaHost string `required:"true" env:"KAFKA_HOST"`
	PgUrl     string `required:"true" env:"PG_URL"`
}

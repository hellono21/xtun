package config

type Config struct {
	Mode	string			`toml:"mode" json:"mode"`
	Script	string			`toml:"script" json:"script"`
	Logging		LoggingConfig		`toml:"logging" json:"logging"`
	Server		Server		`toml:"server" json:"server"`
	Client		Client		`toml:"client" json:"client"`
}

type Server struct {
	Bind	string			`toml:"bind" json:"bind"`
}

type Client struct {
	Remote	string			`toml:"remote" json:"remote"`
}

type LoggingConfig struct {
	Level	string			`toml:"level" json:"level"`
	Output	string			`toml:"output" json:"output"`
}

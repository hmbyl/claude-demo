package conf

import "time"

type Bootstrap struct {
	Server *Server `yaml:"server"`
	Data   *Data   `yaml:"data"`
	Auth   *Auth   `yaml:"auth"`
}

type Server struct {
	HTTP *ServerHTTP `yaml:"http"`
	GRPC *ServerGRPC `yaml:"grpc"`
}

type ServerHTTP struct {
	Addr    string        `yaml:"addr"`
	Timeout time.Duration `yaml:"timeout"`
}

type ServerGRPC struct {
	Addr    string        `yaml:"addr"`
	Timeout time.Duration `yaml:"timeout"`
}

type Data struct {
	Database *Database `yaml:"database"`
	Redis    *Redis    `yaml:"redis"`
}

type Database struct {
	Driver string `yaml:"driver"`
	Source string `yaml:"source"`
}

type Redis struct {
	Addr         string        `yaml:"addr"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type Auth struct {
	JWTSecret string `yaml:"jwt_secret"`
}

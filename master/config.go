package master

import "flag"

type Config struct {
	APIPort int
	ReadTimeOut int
	WriteTimeOut int
}

func parseConfig() *Config{
	config := &Config{}
	flag.IntVar(&config.APIPort, "port", 8080, "api server port")
	flag.IntVar(&config.ReadTimeOut, "ro", 5000, "read time out")
	flag.IntVar(&config.WriteTimeOut, "wo", 5000, "write time out")
	return config
}
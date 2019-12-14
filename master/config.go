package master

import "flag"

var GConfig *Config

type Config struct {
	APIPort      int
	ReadTimeOut  int
	WriteTimeOut int
}

func init() {
	GConfig = &Config{}
	flag.IntVar(&GConfig.APIPort, "port", 8080, "api server port")
	flag.IntVar(&GConfig.ReadTimeOut, "ro", 5000, "read time out")
	flag.IntVar(&GConfig.WriteTimeOut, "wo", 5000, "write time out")
}

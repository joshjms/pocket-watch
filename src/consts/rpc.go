package consts

import (
	"os"
	"strconv"
)

type RPC struct {
	Port int
}

func initializeRpcConfig() (*RPC, error) {
	var config RPC

	port, err := strconv.Atoi(os.Getenv("RPC_PORT"))
	if err != nil {
		return nil, err
	}

	config.Port = port

	return &config, nil
}

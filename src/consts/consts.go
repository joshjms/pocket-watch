package consts

type Consts struct {
	RPCConfig *RPC
}

var config *Consts

func Init() error {

	rpcconfig, err := initializeRpcConfig()
	if err != nil {
		return err
	}

	config = &Consts{
		RPCConfig: rpcconfig,
	}

	return nil
}

func GetConsts() *Consts {
	return config
}

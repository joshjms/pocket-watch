package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	api "github.com/joshjms/pocket-watch/src"
	"github.com/joshjms/pocket-watch/src/consts"
	"github.com/joshjms/pocket-watch/src/isolate"
	"github.com/joshjms/pocket-watch/src/rpc"
)

func main() {
	var err error

	err = godotenv.Load()
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to load environment variables: %v", err)
		log.Fatal(errorMsg)
	}

	err = consts.Init()
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to load environment variables: %v", err)
		log.Fatal(errorMsg)
	}

	isolate.InitQueueManager()

	if os.Getenv("RPC_ENABLE") == "TRUE" {
		err = rpc.StartServer()
		if err != nil {
			log.Fatal(err)
		}

		log.Print(fmt.Sprintf("Running gRPC server on port %d", consts.GetConsts().RPCConfig.Port))

	}

	api.StartAPI()

}

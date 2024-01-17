package main

import (
	"fmt"
	"log"
	"os"

	api "github.com/joshjms/pocket-watch/src"
	"github.com/joshjms/pocket-watch/src/consts"
	"github.com/joshjms/pocket-watch/src/isolate"
	"github.com/joshjms/pocket-watch/src/rpc"
)

func main() {
	var err error

	err = consts.Init()
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to load environment variable: %v", err)
		log.Fatal(errorMsg)
	}

	isolate.InitQueueManager()
	api.StartAPI()

	if os.Getenv("RPC_ENABLE") == "TRUE" {
		err = rpc.StartServer()
		if err != nil {
			panic(err)
		}
	}

}

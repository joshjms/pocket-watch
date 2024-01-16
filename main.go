package main

import (
	api "github.com/joshjms/pocket-watch/src"
	"github.com/joshjms/pocket-watch/src/isolate"
)

func main() {
	isolate.InitQueueManager()
	api.Start()
}

package api

import "github.com/joshjms/pocket-watch/src/routes"

func StartAPI() {
	e := routes.AddRoutes()
	e.Logger.Fatal(e.Start(":8080"))
}

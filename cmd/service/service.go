package main

import (
	"fmt"
	"net/http"

	"helm.sh/helm/v3/pkg/http/api"
	"helm.sh/helm/v3/pkg/http/api/list"
	"helm.sh/helm/v3/pkg/http/api/logger"
	"helm.sh/helm/v3/pkg/http/api/ping"
	"helm.sh/helm/v3/pkg/servercontext"
)

func main() {
	app := servercontext.NewApp()
	startServer(app)
}

func startServer(appconfig *servercontext.Application) {
	router := http.NewServeMux()

	//TODO: use gorilla mux and add middleware to write content type and other headers
	app := servercontext.App()
	logger.Setup("debug")
	service := api.NewService(app.Config, app.ActionConfig)
	router.Handle("/ping", ping.Handler())
	router.Handle("/list", list.Handler())
	router.Handle("/install", api.Install(service))
	router.Handle("/upgrade", api.Upgrade(service))

	err := http.ListenAndServe(fmt.Sprintf(":%d", 8080), router)
	if err != nil {
		fmt.Println("error starting server", err)
	}
}

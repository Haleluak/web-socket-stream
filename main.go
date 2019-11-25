package main

import (
	"github.com/micro/cli"
	"github.com/micro/go-micro/config"
	"github.com/micro/go-micro/util/log"
	"github.com/micro/go-micro/web"
	cf "gitlab.com/brazncorp/brazn-websocket/config"
	"gitlab.com/brazncorp/brazn-websocket/handler"
	"gitlab.com/brazncorp/brazn-websocket/util"
)

func init() {
	err := cf.LoadConfigFile("config.json")
	if err != nil {
		log.Fatal(err)
	}

}

func main() {
	service := web.NewService()

	if err := service.Init(
		web.Action(func(context *cli.Context) {
		handler.Init(service.Options().Service.Options().Broker)
	}),	web.BeforeStart(handler.Subsubcriber)); err != nil {
		log.Fatal(err)
	}

	service.HandleFunc("/stream", handler.Stream)

	log.Info("[Init] Config: ", util.Beautify(config.Map()))
	if err := service.Run(); err != nil {
		log.Fatal("[Init] Error: ", util.Beautify(err))
	}
}

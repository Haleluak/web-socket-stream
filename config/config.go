package config

import (
	"github.com/micro/go-micro/config"
	"github.com/micro/go-micro/config/source/file"
	"github.com/micro/go-micro/util/log"
	"gitlab.com/brazncorp/brazn-websocket/util"
)

func LoadConfigFile(fileName string) error {
	err := config.Load(
		//env.NewSource(),
		file.NewSource(
			file.WithPath(fileName),
		),
	)

	if err != nil {
		log.Info("[Config] Failed to connect file. Error: ", err.Error())
		return err
	}

	watcher, err := config.Watch()
	if err != nil {
		log.Info("[Config] start watching files error，%s", err)
		return err
	}

	refreshLogLevel()

	go func() {
		for {
			_, err := watcher.Next()
			if err != nil {
				log.Info("[Config] watch files error，%s", err)
				return
			}
			refreshLogLevel()
			log.Info("[Config] Changed， %s", util.Beautify(config.Map()))
		}
	}()

	return nil
}

func refreshLogLevel() {
	l := int(log.GetLevel())
	l = config.Get("log_level").Int(l)
	log.SetLevel(log.Level(l))
}

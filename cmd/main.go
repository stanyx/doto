package main

import (
	"flag"

	"github.com/stanyx/doto/internal/doto"

	"github.com/stanyx/nanolog"
	log "github.com/stanyx/nanolog"
)

func main() {

	var telegramToken string
	var maxTasksPerWorker int

	flag.StringVar(&telegramToken, "token", "", "telegram bot token")
	flag.IntVar(&maxTasksPerWorker, "tpw", 100, "max tasks per worker")
	flag.Parse()

	log.Init(log.Options{
		Level: nanolog.DebugLevel,
	})

	if telegramToken == "" {
		log.Fatal().Println("telegram token argument is missing")
		return
	}

	app := doto.New(&doto.Config{
		TelegramBotToken:  telegramToken,
		MaxTasksPerWorker: maxTasksPerWorker,
	})

	log.Debug().Println("starting the application")
	if err := app.Start(); err != nil {
		log.Error().Println("application stopped with error: ", err)
		return
	}
	log.Debug().Println("application was stopped")
}

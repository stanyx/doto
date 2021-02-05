package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/stanyx/doto/internal/doto"
	log "github.com/stanyx/nanolog"
)

func main() {

	log.Init(log.Options{
		Level: log.DebugLevel,
	})

	var apiUrl string
	var putCommand string
	var chatID string
	var date string
	var message string
	var rmCommand string

	flag.StringVar(&apiUrl, "url", "http://localhost:5000", "api url")
	flag.StringVar(&putCommand, "put", "1", "create new event for notification")
	flag.StringVar(&chatID, "chat", "", "telegram chat id")
	flag.StringVar(&date, "date", "", "date when to send message")
	flag.StringVar(&message, "msg", "", "message to send")
	flag.StringVar(&rmCommand, "rm", "", "remove existing event")
	flag.Parse()

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	var request *http.Request
	if putCommand != "" {

		e := doto.EventForm{
			ChatID:      chatID,
			Description: message,
		}
		timeToNotify, err := time.Parse(time.RFC3339, date)
		if err != nil {
			log.Fatal().Println(err)
			return
		}
		e.Timestamp = timeToNotify

		body := bytes.NewBuffer([]byte{})

		if err := json.NewEncoder(body).Encode(&e); err != nil {
			log.Fatal().Println(err)
			return
		}

		r, err := http.NewRequest("POST", apiUrl+"/event", body)
		if err != nil {
			log.Fatal().Println(err)
			return
		}

		log.Debug().Printf("send command to (%s): %s\n", chatID, body.String())

		request = r
	} else {
		log.Fatal().Println("unknown command")
	}

	resp, err := client.Do(request)
	if err != nil {
		log.Fatal().Println(err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatal().Println("response code is not ok, ", resp.StatusCode)
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal().Println(err)
	}

	log.Debug().Println("response for request", string(body))

	log.Info().Println("command executed successfully")

}

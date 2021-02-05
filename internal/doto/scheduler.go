package doto

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	log "github.com/stanyx/nanolog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Scheduler struct {
	app                *App
	bot                *tgbotapi.BotAPI
	maxTaskPerInstance int
}

func NewScheduler(app *App, telegramToken string, maxTaskPerInstance int) (*Scheduler, error) {

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	bot, err := tgbotapi.NewBotAPIWithClient(telegramToken, client)
	if err != nil {
		return nil, err
	}

	sch := &Scheduler{
		app:                app,
		bot:                bot,
		maxTaskPerInstance: maxTaskPerInstance,
	}

	return sch, nil
}

func (s *Scheduler) Start() {

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGQUIT)

	for {

		select {
		case <-sigCh:
			return
		default:
		}

		queryTime := time.Now()
		entries := s.app.GetEntries()
		log.Debug().Printf("get entries time: %v, count: %d\n", time.Since(queryTime), len(entries))

		keysToDelete := make(map[string]bool)
		if len(entries) > 0 {
			workerCount := len(entries)/s.maxTaskPerInstance + 1

			taskChan := make(chan *Event)

			var wg sync.WaitGroup
			wg.Add(workerCount)

			for i := 0; i < workerCount; i++ {
				go s.Worker(taskChan, &wg)
			}

			for _, e := range entries {
				taskChan <- e
				keysToDelete[s.app.GetEventKey(e)] = true
			}

			close(taskChan)

			wg.Wait()
		}

		s.app.cache.Range(func(k interface{}, _ interface{}) bool {
			key := k.(string)
			if keysToDelete[key] {
				s.app.cache.Delete(k)
			}
			return true
		})

		time.Sleep(time.Second * 10)
	}
}

func (s *Scheduler) Worker(taskCh chan *Event, wg *sync.WaitGroup) {
	for event := range taskCh {

		if event == nil {
			break
		}

		chatInt, _ := strconv.Atoi(event.ChatID)

		msg := tgbotapi.NewMessage(int64(chatInt), fmt.Sprintf(`
			<b>%s</b>
			<i>%s</i>
		`, event.Title, event.Description))
		msg.ParseMode = "html"

		if _, err := s.bot.Send(msg); err != nil {
			log.Error().Printf("error send message (%d): %s\n", int64(chatInt), err)
		}
	}

	wg.Done()
}

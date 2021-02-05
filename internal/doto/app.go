package doto

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	log "github.com/stanyx/nanolog"
)

type App struct {
	cache             sync.Map
	serverPort        int
	server            *http.Server
	telegramToken     string
	maxTasksPerWorker int
}

func New(cfg *Config) *App {
	return &App{
		cache:             sync.Map{},
		serverPort:        5000,
		telegramToken:     cfg.TelegramBotToken,
		maxTasksPerWorker: cfg.MaxTasksPerWorker,
	}
}

func (a *App) GetEntries() []*Event {
	var es []*Event
	timeNow := time.Now()
	a.cache.Range(func(_ interface{}, e interface{}) bool {
		event := e.(*Event)
		if event.Timestamp.Before(timeNow) {
			es = append(es, event)
		}
		return true
	})
	return es
}

func (a *App) GetEventKey(e *Event) string {
	return fmt.Sprintf("%s:%d", e.ChatID, e.Timestamp.Unix())
}

func (a *App) CreateEvent(e *Event) {
	key := a.GetEventKey(e)
	a.cache.Store(key, e)
	log.Debug().Printf("add new task: %+v\n", e)
}

func (a *App) DeleteEvent(e *Event) {
	key := a.GetEventKey(e)
	a.cache.Delete(key)
	log.Debug().Printf("delete task: %+v\n", e)
}

type EventForm struct {
	ChatID      string    `json:"chat_id"`
	Timestamp   time.Time `json:"timestamp"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
}

func (a *App) Start() error {

	mx := http.NewServeMux()

	mx.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {

		var event EventForm
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte("error"))
			return
		}

		if r.Method == "POST" {
			a.CreateEvent(&Event{
				ChatID:      event.ChatID,
				Timestamp:   event.Timestamp,
				Title:       event.Title,
				Description: event.Description,
			})
		} else if r.Method == "DELETE" {
			a.DeleteEvent(&Event{
				ChatID:    event.ChatID,
				Timestamp: event.Timestamp,
			})
		}

		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.serverPort),
		Handler: mx,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Error().Printf("server closed with error: %s\n", err)
			return
		}
	}()

	a.server = server

	scheduler, err := NewScheduler(a, a.telegramToken, a.maxTasksPerWorker)
	if err != nil {
		return err
	}

	scheduler.Start()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

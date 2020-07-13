package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/theverything/communique/internal/hub"
	"github.com/theverything/communique/internal/notify"
)

type Config struct {
	Port int
}

type notification struct {
	Topic   string          `json:"topic"`
	Payload json.RawMessage `json:"payload"`
}

type handler struct {
	dispatcher hub.Hub
}

func (h *handler) notify(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported.", http.StatusInternalServerError)
		return
	}

	t := r.URL.Query().Get("topic")
	if t == "" {

		http.Error(w, "Missing `topics` query param.", http.StatusBadRequest)
		return
	}

	client := notify.New()

	h.dispatcher.Join(t, client)
	defer func() {
		log.Println("unregistering client")
		h.dispatcher.Leave(t, client)
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	closeNotify := w.(http.CloseNotifier).CloseNotify()

	ping := time.NewTicker(time.Minute)
	defer ping.Stop()

	for {
		select {
		case payload := <-client.C:
			fmt.Fprintf(w, "data: %s\n\n", string(payload))
		case <-ping.C:
			fmt.Fprintf(w, "event: ping\ndata: {\"time\":\"%s\"}\n\n", time.Now().Format(time.RFC3339Nano))
		case <-closeNotify:
			return
		}

		flusher.Flush()
	}
}

func (h *handler) dispatch(w http.ResponseWriter, r *http.Request) {
	var body notification

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Println(err)
		http.Error(w, "Bad request body.", http.StatusInternalServerError)
		return
	}

	go h.dispatcher.Dispatch(body.Topic, body.Payload)

	w.WriteHeader(http.StatusOK)
}

func (h *handler) serveIndex(w http.ResponseWriter, r *http.Request) {
	p, err := filepath.Abs("public/index.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Can not build path to index.html", http.StatusInternalServerError)
		return
	}

	html, err := os.Open(p)
	if err != nil {
		log.Println(err)
		http.Error(w, "Can not open index.html", http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, "index", time.Now(), html)
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(fmt.Sprintf("%s %s", r.Method, r.URL.Path))

	if r.URL.Path == "/api/notify" {
		h.notify(w, r)
		return
	} else if r.URL.Path == "/api/dispatch" {
		h.dispatch(w, r)
		return
	}

	h.serveIndex(w, r)
}

// New -
func New(config Config, dispatcher hub.Hub) *http.Server {
	return &http.Server{
		Handler: &handler{dispatcher: dispatcher},
		Addr:    fmt.Sprintf(":%d", config.Port),
	}
}

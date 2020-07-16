package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/lithammer/shortuuid/v3"
	"github.com/theverything/communique/internal/hub"
	"github.com/theverything/communique/internal/notify"
)

// Config -
type Config struct {
	Port        int
	DisableCORS bool
}

type notification struct {
	Topic   string          `json:"topic"`
	Payload json.RawMessage `json:"payload"`
}

type handler struct {
	dispatcher hub.Hub
}

var (
	messageHeader     = []byte(`{"type":"message","payload":`)
	messageTrailer    = []byte(`}`)
	messageWrapperLen = len(messageHeader) + len(messageTrailer)
)

func createMessage(payload []byte) []byte {
	msg := make([]byte, 0, len(payload)+messageWrapperLen)

	msg = append(msg, messageHeader...)
	msg = append(msg, payload...)
	msg = append(msg, messageTrailer...)

	return msg
}

func setHeaders(next http.Handler, disableCORS bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := "https://chat.jeffh.dev"
		methods := "POST, GET, OPTIONS"
		headers := "Content-Type"

		if disableCORS {
			origin = "*"
			methods = "*"
			headers = "*"
		}

		w.Header().Set("Accept", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", methods)
		w.Header().Set("Access-Control-Allow-Headers", headers)

		if r.Method == http.MethodOptions {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *handler) notify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported.", http.StatusInternalServerError)
		return
	}

	t := r.URL.Query().Get("topic")
	if t == "" {
		http.Error(w, "Missing `topic` query param.", http.StatusBadRequest)
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

	closeNotify := w.(http.CloseNotifier).CloseNotify()

	ping := time.NewTicker(time.Minute)
	defer ping.Stop()

	for {
		select {
		case payload := <-client.C:
			id := shortuuid.New()
			fmt.Fprintf(w, "id: %s\ndata: %s\n\n", id, string(payload))
		case <-ping.C:
			fmt.Fprintf(w, "event: ping\ndata: {\"time\":\"%s\"}\n\n", time.Now().Format(time.RFC3339Nano))
		case <-closeNotify:
			return
		}

		flusher.Flush()
	}
}

func (h *handler) dispatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	var body notification

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Println(err)
		http.Error(w, "Bad request body.", http.StatusInternalServerError)
		return
	}

	go h.dispatcher.Dispatch(body.Topic, createMessage(body.Payload))

	w.WriteHeader(http.StatusOK)
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

	http.NotFound(w, r)
}

// New -
func New(config Config, dispatcher hub.Hub) *http.Server {
	return &http.Server{
		Handler: setHeaders(
			&handler{
				dispatcher: dispatcher,
			},
			config.DisableCORS,
		),
		Addr: fmt.Sprintf(":%d", config.Port),
	}
}

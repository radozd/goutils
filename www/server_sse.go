package www

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type SseClient struct {
	ch         chan string
	remoteAddr string
	id         uint64

	lastMessageId   uint64
	lastMessageSent time.Time
	overflow        bool
}

func (c *SseClient) String() string {
	return fmt.Sprintf("id=%d, ip=%s, lastMessageId=%d, lastMessageSent=%s, chan=%d, overflow=%v",
		c.id, c.remoteAddr, c.lastMessageId, c.lastMessageSent.Format(time.RFC3339), len(c.ch), c.overflow)
}

type SseMessage struct {
	id    uint64
	data  string
	event string
	retry int
}

func (m *SseMessage) String() string {
	var buffer bytes.Buffer

	if m.retry > 0 {
		buffer.WriteString(fmt.Sprintf("retry: %d\n", m.retry))
	}
	if len(m.event) > 0 {
		buffer.WriteString(fmt.Sprintf("event: %s\n", m.event))
	}
	if len(m.data) > 0 {
		buffer.WriteString(fmt.Sprintf("data: %s\n", strings.Replace(m.data, "\n", "\ndata: ", -1)))
	}
	if m.id > 0 {
		buffer.WriteString(fmt.Sprintf("id: %d\n", m.id))
	}
	buffer.WriteString("\n")

	return buffer.String()
}

type SseBroker struct {
	addClient chan *SseClient
	delClient chan *SseClient
	messages  chan string

	mu            sync.Mutex
	clients       map[uint64]*SseClient
	nextClientId  uint64
	nextMessageId uint64

	debugLog bool
}

func (b *SseBroker) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	str := fmt.Sprintf("SSE: clients=%d, nextId=%d\n", len(b.clients), b.nextMessageId)
	for _, cli := range b.clients {
		str += "  " + cli.String() + "\n"
	}
	return str
}

func StartSseBroker(debugLog bool) *SseBroker {
	b := SseBroker{
		clients:   make(map[uint64]*SseClient),
		addClient: make(chan *SseClient),
		delClient: make(chan *SseClient),
		messages:  make(chan string),
		debugLog:  debugLog,
	}

	go b.SseThread()

	return &b
}

func (b *SseBroker) Stop() {
	for _, cli := range b.clients {
		b.delClient <- cli
	}
}

func (b *SseBroker) MakeMessage(str string) *SseMessage {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.nextMessageId++
	msg := &SseMessage{
		id:    b.nextMessageId,
		data:  str,
		event: "",
		retry: 1000,
	}
	if b.debugLog {
		log.Print("SSE: " + msg.String())
	}
	return msg
}

// SendMessage broadcast simple string
func (b *SseBroker) SendMessage(msg string) {
	b.messages <- msg
}

func (b *SseBroker) SendMap(msg map[string]interface{}) {
	buffer, _ := json.Marshal(msg)
	b.messages <- string(buffer)
}

func (b *SseBroker) ClientCount() int {
	return len(b.clients)
}

func (b *SseBroker) AddClient(cli *SseClient) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.nextClientId++
	cli.id = b.nextClientId
	if b.debugLog {
		log.Println("SSE: connect", cli.String())
	}
	b.clients[cli.id] = cli
}

func (b *SseBroker) DelClient(cli *SseClient) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.debugLog {
		log.Println("SSE: disconnect", cli.String())
	}
	delete(b.clients, cli.id)
	close(cli.ch)
}

// SseThread handles the addition & removal of clients, as well as the broadcasting
// of messages out to clients that are currently attached
func (b *SseBroker) SseThread() {
	for {
		select {
		case cli := <-b.addClient:
			b.AddClient(cli)

		case cli := <-b.delClient:
			b.DelClient(cli)

		case msg := <-b.messages:
			if b.debugLog {
				log.Println("SSE: new broadcast message")
			}
			for _, cli := range b.clients {
				select {
				case cli.ch <- msg:
					cli.lastMessageId = b.nextMessageId
				default:
					cli.overflow = true
					log.Println("SSE: channel overflow", cli.String())
				}
			}
		}
	}
}

func ReadUserIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-Ip")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}

// ServeHTTP handles and HTTP request at the "/events/" URL
func (b *SseBroker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if b.debugLog {
		log.Print("SSE: SseBroker.ServeHTTP started")
	}

	defer func() {
		if b.debugLog {
			log.Print("SSE: SseBroker.ServeHTTP exited")
		}
	}()

	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE: Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	cli := &SseClient{
		ch:         make(chan string, 10),
		remoteAddr: ReadUserIP(r),
	}
	b.addClient <- cli

	greetings, _ := json.Marshal(map[string]interface{}{})
	fmt.Fprint(w, b.MakeMessage(string(greetings)).String())
	f.Flush()
	cli.lastMessageSent = time.Now()

	ctx := r.Context()
	for {
		select {
		case str, open := <-cli.ch:
			if !open { // Stop() сам делает delClient, явно вызывать не надо
				if b.debugLog {
					log.Print("SSE: SseBroker.ServeHTTP !open")
				}
				return
			}
			msg := b.MakeMessage(str)
			fmt.Fprint(w, msg.String())
			f.Flush()
			cli.lastMessageSent = time.Now()

		case <-ctx.Done(): // клиент отключился по своей инициативе
			if b.debugLog {
				log.Print("SSE: SseBroker.ServeHTTP ctx.Done")
			}
			b.delClient <- cli
			return
		}
	}
}

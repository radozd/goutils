package www

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/radozd/goutils/logger"
)

// LocalServer - Serve будет запускать локальный браузер и закрываться, когда браузер закроется
var LocalServer bool = false

// DefaultBrowser если пусто, запускается браузер по умолчанию
var DefaultBrowser string

func PanicHandler(w http.ResponseWriter) {
	if r := recover(); r != nil {
		var err error
		switch x := r.(type) {
		case string:
			err = errors.New(x)
		case error:
			err = x
		default:
			err = errors.New("unknown panic")
		}
		log.Print(err)
		log.Print(string(debug.Stack()[:]))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ReadInputJSON(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	return data
}

// SendJSONBuffer сжимает, проставляет заголовки и отправляет json
func SendJSONBuffer(w http.ResponseWriter, buffer []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	writer, _ := gzip.NewWriterLevel(w, gzip.BestCompression)
	defer writer.Close()
	writer.Write(buffer)
}

func runUntilSomeoneIsConnected(serverSSE *SseBroker, programInterrupt chan os.Signal) {
	counter := 0
	for counter < 6 {
		time.Sleep(time.Second)

		if serverSSE.ClientCount() == 0 {
			counter++
		} else {
			counter = 0
		}
	}
	log.Println("SSE: no active clients. Exiting.")
	programInterrupt <- os.Interrupt
}

// Serve запускает преднастроенный сервер и ждет завершения
func Serve(ctx context.Context, port int, page string, serverSSE *SseBroker, programInterrupt chan os.Signal) error {
	sport := ":" + strconv.Itoa(port)
	log.Println("WWW: running at " + sport)
	server := &http.Server{Addr: sport, Handler: nil, ReadTimeout: 5 * time.Minute, WriteTimeout: 5 * time.Minute}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("WWW: listen:%+s\n", err)
		}
	}()

	if LocalServer {
		go func() {
			<-time.After(500 * time.Millisecond)
			if err := Browse("http://localhost"+sport+page, DefaultBrowser); err != nil {
				log.Println("WWW:", err)
			}

			if serverSSE != nil {
				runUntilSomeoneIsConnected(serverSSE, programInterrupt)
			}
		}()
	}

	log.Printf("WWW: server started")
	<-ctx.Done()
	log.Printf("WWW: server stopped")

	if serverSSE != nil {
		serverSSE.Stop()
	}

	ctx_shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var err error
	if err = server.Shutdown(ctx_shutdown); err != nil {
		log.Fatalf("WWW: shutdown failed:%+s", err)
	}

	log.Printf("WWW: exited properly")

	if err == http.ErrServerClosed {
		err = nil
	}
	return err
}

func Run(serve func(context.Context, chan os.Signal) error) {
	logger := logger.NewLogger()
	defer logger.Close()
	log.Println("--------------------------------------")

	program_interrupt := make(chan os.Signal, 1)
	signal.Notify(program_interrupt, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		oscall := <-program_interrupt
		log.Printf("WWW: system call:%+v", oscall)
		cancel()
	}()

	if err := serve(ctx, program_interrupt); err != nil {
		log.Printf("WWW: failed to serve:+%v\n", err)
	}
}

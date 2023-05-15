package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	listenFlag = flag.String("listen", ":3000", "address and port to listen")
	textFlag   = flag.String("text", "It works!", "text to put on the webpage")
	keyFlag    = flag.String("key", "This is @ $3cr3t key", "text that simulates the secret key")

	// "Global state"
	initialized = flag.Bool("initialized", false, "skip the initialization mock")

	// stdoutW and stderrW are for overriding in test.
	stdoutW = os.Stdout
	stderrW = os.Stderr
)

func main() {
	flag.Parse()

	echoText := *textFlag
	secretKey := *keyFlag

	// Validation
	args := flag.Args()
	if len(args) > 0 {
		fmt.Fprintln(stderrW, "Too many arguments!")
		os.Exit(127)
	}

	mux := http.NewServeMux()

	// "content" route
	mux.HandleFunc("/", httpLog(stdoutW, withAppHeaders(0, httpEcho(echoText))))

	// Expose half secret key
	mux.HandleFunc("/internal/key", httpLog(stdoutW, withAppHeaders(200, httpEchoKey(secretKey))))

	// Recieve full secret
	mux.HandleFunc("/internal/secret", httpLog(stdoutW, withAppHeaders(0, secretHandSkake)))

	// Health endpoint
	mux.HandleFunc("/health", withAppHeaders(200, httpHealth()))

	server := &http.Server{
		Addr:    *listenFlag,
		Handler: mux,
	}
	serverCh := make(chan struct{})
	go func() {
		log.Printf("[INFO] server is listening on %s\n", *listenFlag)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("[ERR] server exited with: %s", err)
		}
		close(serverCh)
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	// Wait for interrupt
	<-signalCh

	log.Printf("[INFO] received interrupt, shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("[ERR] failed to shutdown server: %s", err)
	}

	// If we got this far, it was an interrupt, so don't exit cleanly
	os.Exit(2)
}

func secretHandSkake(w http.ResponseWriter, r *http.Request) {
	if *initialized {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "unable to read secret")
		return
	}
	defer r.Body.Close()

	decoded, err := base64.StdEncoding.DecodeString(string(b))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "unable to read secret")
		return
	}

	fullSecret := string(decoded)
	if !strings.Contains(fullSecret, *keyFlag) || len(fullSecret) == len(*keyFlag) {
		w.WriteHeader(400)
		fmt.Fprintln(w, "invalid key")
		return
	}

	*initialized = true
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, "Accepted")
}

func httpEcho(v string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !*initialized {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintln(w, "Uninitialized")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, v)
	}
}

func httpEchoKey(v string) http.HandlerFunc {
	secret := base64.StdEncoding.EncodeToString([]byte(v))

	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, secret)
	}
}

func httpHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"status":"ok"}`)
	}
}

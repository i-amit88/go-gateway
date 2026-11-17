package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"example.com/go-gateway/config"
	"example.com/go-gateway/internal/httpapi"
	"github.com/joho/godotenv"
)

const maxRequestBodySize = 1 << 20

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found; using system environment")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	target, err := url.Parse(cfg.UpstreamURL)
	if err != nil {
		log.Fatal(err)
	}

	httpClient := &http.Client{
		Timeout: 1 * time.Second,
	}

	readyCheck := func(ctx context.Context) error {
		return checkUpstream(
			ctx,
			httpClient,
			cfg.UpstreamURL,
		)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = handleProxyError

	router := httpapi.NewRouter(
		proxy,
		cfg.ProxyTimeout,
		cfg.APIKey,
		readyCheck,
	)
	router = httpapi.WithBodyLimit(maxRequestBodySize, router)
	router = httpapi.WithRecovery(router)
	router = httpapi.WithLogging(router)
	router = httpapi.WithRequestID(router)
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Println("server starting on :" + cfg.Port)
		serverErr <- server.ListenAndServe()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	select {
	case sig := <-signals:
		log.Printf("received shutdown signal: %s", sig)

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
		}

	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}
}

func handleProxyError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf(
		"upstream proxy error: method=%s path=%s error=%v",
		r.Method,
		r.URL.Path,
		err,
	)

	status := http.StatusBadGateway
	message := "upstream unavailable"

	timedOut := errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(r.Context().Err(), context.DeadlineExceeded)

	if timedOut {
		status = http.StatusGatewayTimeout
		message = "upstream timeout"
	}

	http.Error(w, message, status)
}

func checkUpstream(
	ctx context.Context,
	client *http.Client,
	upstreamURL string,
) error {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		strings.TrimRight(upstreamURL, "/")+"/health",
		nil,
	)
	if err != nil {
		return err
	}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"upstream returned status %d",
			response.StatusCode,
		)
	}

	return nil
}

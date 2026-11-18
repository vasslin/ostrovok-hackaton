package app

import (
	"ai-service/internal/adapter/agent"
	"ai-service/internal/service"
	"ai-service/internal/transport/grpcapi"
	"ai-service/internal/transport/httpapi"
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func Run() error {
	port := env("APP_PORT", "8080")
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return errors.New("APP_PORT must be 1..65535")
	}

	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		c := http.Client{Timeout: 2 * time.Second}
		res, err := c.Get("http://127.0.0.1:" + port + "/health")
		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.StatusCode != 200 {
			return errors.New("healthcheck failed")
		}

		return nil
	}
	provider := env("LLM_PROVIDER", "ollama")
	var client *agent.Client
	if provider == "offline" {
		client = agent.NewOffline()
	} else {
		defaultURL, defaultModel := "http://localhost:11434/v1", "travel-qwen:4b"
		if provider == "groq" || provider == "openai" {
			defaultURL = "https://api.groq.com/openai/v1"
			defaultModel = "openai/gpt-oss-20b"
		} else if provider != "ollama" {
			return errors.New("LLM_PROVIDER must be offline, ollama, groq or openai")
		}
		baseURL := env("LLM_BASE_URL", defaultURL)
		parsed, err := url.Parse(baseURL)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil {
			return errors.New("invalid LLM_BASE_URL")
		}
		client = agent.New(baseURL, os.Getenv("LLM_API_KEY"), env("LLM_MODEL", defaultModel))
		if provider == "ollama" {
			client = agent.NewLocal(baseURL, env("LLM_MODEL", defaultModel))
		}
	}

	for _, file := range []string{"openapi.json", "index.html", "swagger/swagger-ui-bundle.js", "swagger/swagger-ui.css"} {
		if _, err := os.Stat(env("WEB_DIR", "web") + "/" + file); err != nil {
			return err
		}
	}
	timeout, err := time.ParseDuration(env("LLM_TIMEOUT", client.HTTP.Timeout.String()))
	if err != nil || timeout < time.Second || timeout > 10*time.Minute {
		return errors.New("LLM_TIMEOUT must be between 1s and 10m")
	}
	client.HTTP.Timeout = timeout
	svc := service.New(client)
	server := &http.Server{Addr: ":" + port, Handler: httpapi.New(svc, os.Getenv("SERVICE_API_KEY")), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: timeout + 10*time.Second, IdleTimeout: 60 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	grpcPort := env("GRPC_PORT", "9090")
	grpcNumber, err := strconv.Atoi(grpcPort)
	if err != nil || grpcNumber < 1 || grpcNumber > 65535 || grpcNumber == number {
		return errors.New("GRPC_PORT must be 1..65535 and differ from APP_PORT")
	}

	listener, err := net.Listen("tcp", ":"+grpcPort)

	if err != nil {
		return err
	}

	defer listener.Close()
	grpcServer := grpcapi.New(svc, os.Getenv("SERVICE_API_KEY"))
	defer grpcServer.Stop()
	done := make(chan error, 2)
	go func() { done <- grpcServer.Serve(listener) }()
	go func() { done <- server.ListenAndServe() }()
	defer server.Close()
	log.Printf("travel-ai HTTP=%s gRPC=:%s model=%s configured=%t", server.Addr, grpcPort, client.Model, client.Configured())
	select {
	case err := <-done:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return err
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), timeout+10*time.Second)
		defer cancel()
		grpcDone := make(chan struct{})
		go func() { grpcServer.GracefulStop(); close(grpcDone) }()
		if err := server.Shutdown(shutdown); err != nil {
			_ = server.Close()
			return err
		}

		select {
		case <-grpcDone:
		case <-shutdown.Done():
			grpcServer.Stop()
		}
		return nil
	}
}

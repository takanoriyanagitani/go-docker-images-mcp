package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	imgs2mcp "github.com/takanoriyanagitani/go-docker-images-mcp"
)

var port = flag.Int("port", 12029, "port to listen")
var dockerHost = flag.String("docker-host", "/var/run/docker.sock", "Path to the Docker Unix socket")

func main() {
	flag.Parse()

	handler, cli, err := imgs2mcp.NewServer(*dockerHost)
	if err != nil {
		log.Fatalf("failed to create server: %v\n", err)
	}
	defer func() {
		if closeErr := cli.Close(); closeErr != nil {
			log.Printf("Error closing Docker client: %v", closeErr)
		}
	}()

	address := fmt.Sprintf(":%d", *port)

	hserver := &http.Server{
		Addr:           address,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("ready to start http mcp server. listening on %s\n", address)
		if err := hserver.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to listen and serve: %v\n", err)
		}
	}()

	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := hserver.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v\n", err)
	}
	log.Println("server exited gracefully")
}

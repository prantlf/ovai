package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prantlf/ovai/internal/log"
	"github.com/prantlf/ovai/internal/routes"
	"github.com/prantlf/ovai/internal/web"
)

const version = "0.1.4"

func main() {
	if log.IsDbg {
		cwd, _ := os.Getwd()
		log.Dbg("version %s runs in %s", version, cwd)
	}

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "22434"
	}
	server := &http.Server{
		Addr: ":" + port,
	}
	sigch := make(chan os.Signal, 1)

	http.HandleFunc("/api/chat", web.WrapHandler(routes.HandleChat, []string{"POST"}))
	http.HandleFunc("/api/embeddings", web.WrapHandler(routes.HandleEmbeddings, []string{"POST"}))
	http.HandleFunc("/api/generate", web.WrapHandler(routes.HandleGenerate, []string{"POST"}))
	http.HandleFunc("/api/ping", web.WrapHandler(routes.HandlePing, []string{"GET"}))
	http.HandleFunc("/api/shutdown", func(w http.ResponseWriter, r *http.Request) {
		if routes.HandleShutdown(w, r) {
			close(sigch)
		}
	})

	go func() {
		if log.IsDbg {
			log.Dbg("listen on http://localhost:%s", port)
		} else {
			fmt.Printf("Listening on http://localhost:%s ...", port)
		}
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Ftl("listening failed: %v", err)
		}
	}()

	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	<-sigch
	log.Dbg("shut server down")
	if err := server.Shutdown(context.Background()); err != nil {
		log.Ftl("shutting down server failed: %v", err)
		server.Close()
	}
	os.Exit(0)
}

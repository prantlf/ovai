package main

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/prantlf/ovai/internal/log"
	"github.com/prantlf/ovai/internal/routes"
	"github.com/prantlf/ovai/internal/web"
)

const version = "0.11.0"

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

	http.HandleFunc("/api/chat", web.WrapHandler(routes.HandleChat, []string{"POST"}))
	http.HandleFunc("/api/embeddings", web.WrapHandler(routes.HandleEmbeddings, []string{"POST"}))
	http.HandleFunc("/api/generate", web.WrapHandler(routes.HandleGenerate, []string{"POST"}))
	http.HandleFunc("/api/ping", web.WrapHandler(routes.HandlePing, []string{"GET", "HEAD"}))
	http.HandleFunc("/api/show", web.WrapHandler(routes.HandleShow, []string{"POST"}))
	http.HandleFunc("/api/shutdown", web.WrapHandler(routes.HandleShutdown, []string{"POST"}))
	http.HandleFunc("/api/tags", web.WrapHandler(routes.HandleTags, []string{"GET"}))

	go func() {
		log.Log("listen on http://localhost:%s", port)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Ftl("listening failed: %v", err)
		}
	}()

	routes.WaitForShutdown()
	log.Log("shut server down")
	if err := server.Shutdown(context.Background()); err != nil {
		log.Log("shutting down server failed: %v", err)
		if err = server.Close(); err != nil {
			log.Log("killing server failed: %v", err)
		}
	}
}

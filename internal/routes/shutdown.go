package routes

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prantlf/ovai/internal/log"
)

var sigch = make(chan os.Signal, 1)

func WaitForShutdown() {
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	<-sigch
}

func HandleShutdown(w http.ResponseWriter, r *http.Request) int {
	log.Dbg(": shut down")
	w.WriteHeader(http.StatusNoContent)
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(sigch)
	}()
	return http.StatusNoContent
}

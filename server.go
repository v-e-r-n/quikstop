package quikstop

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ListenAndServeGracefully boots the server in the background and intercepts
// SIGINT/SIGTERM to safely flush connections before shutting down.
func ListenAndServeGracefully(srv *http.Server, shutdownTimeout time.Duration) {
	go func() {
		log.Printf("Server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server graceful shutdown failed: %v", err)
	}
	log.Println("Server stopped successfully")
}

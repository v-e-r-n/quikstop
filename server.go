package quikstop

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/v-e-r-n/quikstop/core"
)

// ListenAndServeGracefully boots the server in the background and intercepts
// SIGINT/SIGTERM to safely flush connections before shutting down.
func ListenAndServeGracefully(srv *http.Server, shutdownTimeout time.Duration) {
	go func() {
		core.Logger().Info("Server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			core.Logger().Error("Server failed to listen", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	core.Logger().Info("Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		core.Logger().Error("Server graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	core.Logger().Info("Server stopped successfully")
}

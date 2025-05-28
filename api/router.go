package api

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	v1 "github.com/zero-shubham/surveysvc/api/v1"
	"github.com/zero-shubham/surveysvc/config"
	db "github.com/zero-shubham/surveysvc/db/orm"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type Router struct {
	server *gin.Engine
	logger *zerolog.Logger
	conn   db.DBTX
}

func NewRouter(logger *zerolog.Logger, conn db.DBTX) *Router {
	return &Router{
		server: gin.Default(),
		logger: logger,
		conn:   conn,
	}
}

func (r *Router) ErrorHandler(c *gin.Context) {
	c.Next()

	for _, err := range c.Errors {
		r.logger.Err(err).Msg("error while handling error")
	}

}

func (r *Router) Start(ctx context.Context, tp trace.TracerProvider, mp metric.MeterProvider) {
	r.server.Use(
		otelgin.Middleware(
			config.ServiceName,
			otelgin.WithTracerProvider(tp),
			otelgin.WithMeterProvider(mp),
		),
	)

	v1.NewApiV1Service(r.server, r.conn, r.logger)

	go func() {
		// Create context that listens for the interrupt signal from the OS.
		ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
		defer stop()

		srv := &http.Server{
			Addr:    ":50052",
			Handler: r.server,
		}

		// Initializing the server in a goroutine so that
		// it won't block the graceful shutdown handling below
		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				r.logger.Fatal().Msgf("listen: %s\n", err)
			}
		}()

		// Listen for the interrupt signal.
		<-ctx.Done()

		// Restore default behavior on the interrupt signal and notify user of shutdown.
		stop()
		r.logger.Println("shutting down gracefully, press Ctrl+C again to force")

		// The context is used to inform the server it has 5 seconds to finish
		// the request it is currently handling
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			r.logger.Fatal().Err(err).Msg("server forced to shutdown")
		}

		r.logger.Println("Server exiting")
	}()
}

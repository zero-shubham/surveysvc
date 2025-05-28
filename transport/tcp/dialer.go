package tcp

import (
	"context"
	"net"
	"time"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// InstrumentedDialer wraps net.Dialer with instrumentation.
type InstrumentedDialer struct {
	Dialer   *kafka.Dialer
	tracer   trace.Tracer
	logger   *zerolog.Logger
	resolver *net.Resolver
}

// DialContext instruments the DialContext method of net.Dialer.
func (d *InstrumentedDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {

	ctx, span := d.tracer.Start(ctx, "kafka.dial", trace.WithAttributes(
		attribute.String("network", network),
		attribute.String("address", address),
	))
	defer span.End()

	d.logger.Info().Msg("kafka dial initiate")
	startTime := time.Now()
	conn, err := d.resolver.Dial(ctx, network, address)
	duration := time.Since(startTime)
	d.logger.Info().Msg("kafka dial complete")

	span.SetAttributes(attribute.Int64("dial.duration_ms", duration.Milliseconds()))

	if err != nil {
		span.SetAttributes(attribute.String("dial.error", err.Error()))
		return nil, err
	}

	return conn, nil
}

// NewInstrumentedDialer creates a new InstrumentedDialer with optional timeout and keep-alive settings.
func NewInstrumentedDialer(
	timeout time.Duration,
	keepAlive time.Duration,
	tracer trace.Tracer,
	logger *zerolog.Logger,
) *InstrumentedDialer {
	id := InstrumentedDialer{
		tracer: tracer,
		logger: logger,
		resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout:   timeout,
					KeepAlive: keepAlive,
				}
				return d.DialContext(ctx, network, address)
			},
		},
	}
	id.Dialer = &kafka.Dialer{
		Timeout:   timeout,
		KeepAlive: keepAlive,
		DialFunc:  id.DialContext,
	}
	return &id
}

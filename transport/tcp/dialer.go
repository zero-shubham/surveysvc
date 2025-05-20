package tcp

import (
	"context"
	"net"
	"time"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// InstrumentedDialer wraps net.Dialer with instrumentation.
type InstrumentedDialer struct {
	Dialer *kafka.Dialer
	tracer trace.Tracer
}

// DialContext instruments the DialContext method of net.Dialer.
func (d *InstrumentedDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	ctx, span := d.tracer.Start(ctx, "dial", trace.WithAttributes(
		attribute.String("network", network),
		attribute.String("address", address),
	))
	defer span.End()

	startTime := time.Now()
	conn, err := d.Dialer.DialContext(ctx, network, address)
	duration := time.Since(startTime)

	span.SetAttributes(attribute.Int64("dial.duration_ms", duration.Milliseconds()))

	if err != nil {
		span.SetAttributes(attribute.String("dial.error", err.Error()))
		return nil, err
	}

	return conn, nil
}

// NewInstrumentedDialer creates a new InstrumentedDialer with optional timeout and keep-alive settings.
func NewInstrumentedDialer(timeout time.Duration, keepAlive time.Duration, tracer trace.Tracer) *InstrumentedDialer {
	return &InstrumentedDialer{
		Dialer: &kafka.Dialer{
			Timeout:   timeout,
			KeepAlive: keepAlive,
		},
		tracer: tracer,
	}
}

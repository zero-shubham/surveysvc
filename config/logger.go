package config

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

const (
	CorrelationRequestID = "req_id"
)

type CorrelationHook struct{}

func (h CorrelationHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()

	if c, ok := ctx.(*gin.Context); ok {
		if reqID, ok := c.Get(CorrelationRequestID); ok {
			if reqStrID, ok := reqID.(string); ok {
				e.Str("request_id", reqStrID)
			}
		}
		ctx = c.Request.Context()
	}

	span := trace.SpanFromContext(ctx)

	if span.SpanContext().HasTraceID() {
		e.Str("trace_id", span.SpanContext().TraceID().String())
	}

	if span.SpanContext().HasSpanID() {
		e.Str("span_id", span.SpanContext().SpanID().String())
	}

}

func GetLogger() *zerolog.Logger {
	hookedLogger := log.Hook(CorrelationHook{})
	return &hookedLogger
}

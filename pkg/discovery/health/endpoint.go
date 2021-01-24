package health

import (
	"context"
	"time"

	"golang.org/x/time/rate"

	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/sony/gobreaker"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
)

type CheckHandler func(service string) CheckStatus

func NewCheckEndpoint(f CheckHandler, logger log.Logger, duration metrics.Histogram, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) endpoint.Endpoint {
	var checkEndpoint endpoint.Endpoint
	{
		checkEndpoint = makeEndpoint(f)
		checkEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(checkEndpoint)
		checkEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(checkEndpoint)
		checkEndpoint = opentracing.TraceServer(otTracer, "HealthCheck")(checkEndpoint)
		if zipkinTracer != nil {
			checkEndpoint = zipkin.TraceEndpoint(zipkinTracer, "HealthCheck")(checkEndpoint)
		}
		checkEndpoint = LoggingMiddleware(log.With(logger, "method", "HealthCheck"))(checkEndpoint)
		checkEndpoint = InstrumentingMiddleware(duration.With("method", "HealthCheck"))(checkEndpoint)
	}
	return checkEndpoint
}

func makeEndpoint(f CheckHandler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return CheckResponse{f(request.(CheckRequest).Service)}, nil
	}
}

type CheckStatus int

const (
	_ CheckStatus = iota
	CheckStatusServing
	CheckStatusNotServing
)

type CheckRequest struct {
	Service string
}

type CheckResponse struct {
	Status CheckStatus
}

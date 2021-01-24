package health

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/go-kit/kit/transport"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"

	"github.com/websmee/ms/pkg/discovery/health/proto"
)

type grpcServer struct {
	check grpctransport.Handler
	proto.UnimplementedHealthServer
}

func NewGRPCServer(checkEndpoint endpoint.Endpoint, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) proto.HealthServer {
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}

	if zipkinTracer != nil {
		options = append(options, zipkin.GRPCServerTrace(zipkinTracer))
	}

	return &grpcServer{
		check: grpctransport.NewServer(
			checkEndpoint,
			decodeGRPCCheckRequest,
			encodeGRPCCheckResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(otTracer, "HealthCheck", logger)))...,
		),
	}
}

func (s *grpcServer) Check(ctx context.Context, req *proto.HealthCheckRequest) (*proto.HealthCheckResponse, error) {
	_, rep, err := s.check.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*proto.HealthCheckResponse), nil
}

func decodeGRPCCheckRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	return CheckRequest{
		Service: grpcReq.(*proto.HealthCheckRequest).Service,
	}, nil
}

func encodeGRPCCheckResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(CheckResponse)
	return &proto.HealthCheckResponse{Status: proto.HealthCheckResponse_ServingStatus(resp.Status)}, nil
}

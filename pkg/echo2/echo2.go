// Package echo2 contains PROTO3 protoc-generated output and implements
// a test and demo echo services. It is intended as for reflect testing
// only.
package echo2

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// Server implements the server-side of gRPC demo Phone service.
type Server struct {
	UnimplementedEchoServer
}

// Hello is a demo echo service.
func (*Server) Hello(_ context.Context, req *HelloRequest) (*HelloResponse, error) {
	resp := fmt.Sprintf("And to you: %s", *req.Message)
	return &HelloResponse{RobotResponse: &resp}, nil
}

// HelloStream streaming RPC handler.
func (s *Server) HelloStream(req *HelloRequest, stream Echo_HelloStreamServer) error {
	for i := 0; i < 3; i++ {
		resp := fmt.Sprintf("%d. %s", i, *req.Message)
		err := stream.Send(&HelloResponse{RobotResponse: &resp})
		if err != nil {
			return errors.Wrap(err, "cannot send on HelloStream")
		}
	}
	return nil
}

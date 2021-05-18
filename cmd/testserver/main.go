package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/juliaogris/reflect/pkg/echo2"
	"github.com/juliaogris/reflect/pkg/echo3"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type config struct {
	Address string `short:"a" help:"gRPC server address, host:port" placeholder:"ADDRESS" env:"GURL_ADDRESS" default:"localhost:9090"`
}

var cfg = &config{}

func main() {
	_ = kong.Parse(cfg)

	fmt.Println("Starting testserver on", cfg.Address)
	if err := run(cfg.Address); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(addr string) error {
	s := grpc.NewServer()
	echo2Server := &echo2.Server{}
	echo2.RegisterEchoServer(s, echo2Server)
	echo3Server := &echo3.Server{}
	echo3.RegisterEchoServer(s, echo3Server)
	reflection.Register(s)

	h := &http.Server{
		Addr:    addr,
		Handler: rootHandler(s),
	}
	if err := h.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to serve gRPC service: %w", err)
	}
	return nil
}

// From: https://github.com/philips/grpc-gateway-example/issues/22#issuecomment-490733965
// Use x/net/http2/h2c so we can have http2 cleartext connections.
func rootHandler(grpcServer http.Handler) http.Handler {
	hf := func(w http.ResponseWriter, r *http.Request) {
		var label string
		switch {
		case r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc"):
			label = "grpc"
			grpcServer.ServeHTTP(w, r)
		default:
			label = "error"
			http.Error(w, r.URL.Path+": not Implemented", http.StatusNotImplemented)
		}
		fmt.Printf("%-5s: %-4s %s %s\n", label, r.Method, r.URL.Path, r.RemoteAddr)
	}
	return h2c.NewHandler(http.HandlerFunc(hf), &http2.Server{})
}

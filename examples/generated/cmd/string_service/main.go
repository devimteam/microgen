// Microgen appends missed functions.
package main

import (
	"context"
	"errors"
	"fmt"
	generated "github.com/devimteam/microgen/examples/generated"
	service "github.com/devimteam/microgen/examples/generated/service"
	transport "github.com/devimteam/microgen/examples/generated/transport"
	grpc "github.com/devimteam/microgen/examples/generated/transport/grpc"
	http "github.com/devimteam/microgen/examples/generated/transport/http"
	protobuf "github.com/devimteam/microgen/examples/protobuf"
	log "github.com/go-kit/kit/log"
	opentracinggo "github.com/opentracing/opentracing-go"
	errgroup "golang.org/x/sync/errgroup"
	grpc1 "google.golang.org/grpc"
	"io"
	"net"
	http1 "net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := log.With(InitLogger(os.Stdout), "level", "info")
	errorLogger := log.With(InitLogger(os.Stderr), "level", "error")
	logger.Log("message", "Hello, I am alive")
	defer logger.Log("message", "goodbye, good luck")

	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		return InterruptHandler(ctx)
	})

	var svc generated.StringService                      // TODO: = service.NewStringService () // Create new service.
	svc = service.LoggingMiddleware(logger)(svc)         // Setup service logging.
	svc = service.ErrorLoggingMiddleware(logger)(svc)    // Setup error logging.
	svc = service.RecoveringMiddleware(errorLogger)(svc) // Setup service recovering.

	endpoints := transport.Endpoints(svc)
	endpoints = transport.TraceServerEndpoints(endpoints, opentracinggo.NoopTracer{}) // TODO: Add tracer

	grpcAddr := ":8081" // TODO: use normal address
	// Start grpc server.
	g.Go(func() error {
		return ServeGRPC(ctx, &endpoints, grpcAddr, log.With(logger, "transport", "GRPC"))
	})

	httpAddr := ":8080" // TODO: use normal address
	// Start http server.
	g.Go(func() error {
		return ServeHTTP(ctx, &endpoints, httpAddr, log.With(logger, "transport", "HTTP"))
	})

	if err := g.Wait(); err != nil {
		logger.Log("error", err)
	}
}

// InitLogger initialize go-kit JSON logger with timestamp and caller.
func InitLogger(writer io.Writer) log.Logger {
	logger := log.NewJSONLogger(writer)
	logger = log.With(logger, "@timestamp", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)
	return logger
}

// InterruptHandler handles first SIGINT and SIGTERM and returns it as error.
func InterruptHandler(ctx context.Context) error {
	interruptHandler := make(chan os.Signal, 1)
	signal.Notify(interruptHandler, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-interruptHandler:
		return fmt.Errorf("signal received: %v", sig.String())
	case <-ctx.Done():
		return errors.New("signal listener: context canceled")
	}
}

// ServeGRPC starts new GRPC server on address and sends first error to channel.
func ServeGRPC(ctx context.Context, endpoints *transport.EndpointsSet, addr string, logger log.Logger) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	// Here you can add middlewares for grpc server.
	server := grpc.NewGRPCServer(endpoints,
		logger,
		opentracinggo.NoopTracer{}, // TODO: Add tracer
	)
	grpcServer := grpc1.NewServer()
	protobuf.RegisterStringServiceServer(grpcServer, server)
	logger.Log("listen on", addr)
	ch := make(chan error)
	go func() {
		ch <- grpcServer.Serve(listener)
	}()
	select {
	case err := <-ch:
		return fmt.Errorf("grpc server: serve: %v", err)
	case <-ctx.Done():
		grpcServer.GracefulStop()
		return errors.New("grpc server: context canceled")
	}
}

// ServeHTTP starts new HTTP server on address and sends first error to channel.
func ServeHTTP(ctx context.Context, endpoints *transport.EndpointsSet, addr string, logger log.Logger) error {
	handler := http.NewHTTPHandler(endpoints,
		logger,
		opentracinggo.NoopTracer{}, // TODO: Add tracer
	)
	httpServer := &http1.Server{
		Addr:    addr,
		Handler: handler,
	}
	logger.Log("listen on", addr)
	ch := make(chan error)
	go func() {
		ch <- httpServer.ListenAndServe()
	}()
	select {
	case err := <-ch:
		if err == http1.ErrServerClosed {
			return nil
		}
		return fmt.Errorf("http server: serve: %v", err)
	case <-ctx.Done():
		return httpServer.Shutdown(context.Background())
	}
}

package template

import (
	"path/filepath"

	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/generator/write_strategy"
	"github.com/devimteam/microgen/util"
)

type mainTemplate struct {
	Info *GenerationInfo

	logging      bool
	recovering   bool
	errorLogging bool
	grpcServer   bool
	httpServer   bool
	tracing      bool
}

func NewMainTemplate(info *GenerationInfo) Template {
	return &mainTemplate{
		Info: info,
	}
}

func (t *mainTemplate) Render() write_strategy.Renderer {
	f := NewFile("main")
	f.PackageComment(t.Info.FileHeader)
	f.PackageComment(`This file will never be overwritten.`)

	f.Line().Add(t.mainFunc())
	f.Line().Add(t.initLogger())
	f.Line().Add(t.interruptHandler())
	f.Line().Add(t.serveGrpc())
	f.Line().Add(t.serveHTTP())

	return f
}

func (t *mainTemplate) DefaultPath() string {
	return "./cmd/" + util.ToSnakeCase(t.Info.Iface.Name) + "/main.go"
}

func (t *mainTemplate) Prepare() error {
	tags := util.FetchTags(t.Info.Iface.Docs, TagMark+MicrogenMainTag)
	for _, tag := range tags {
		switch tag {
		case RecoverMiddlewareTag:
			t.recovering = true
		case LoggingMiddlewareTag:
			t.logging = true
		case HttpServerTag, HttpTag:
			t.httpServer = true
		case GrpcTag, GrpcServerTag:
			t.grpcServer = true
		case ErrorLoggingMiddlewareTag:
			t.errorLogging = true
		case TracingTag:
			t.tracing = true
		}
	}
	return nil
}

func (t *mainTemplate) ChooseStrategy() (write_strategy.Strategy, error) {
	if util.StatFile(t.Info.AbsOutPath, t.DefaultPath()) == nil {
		return write_strategy.NewNopStrategy(t.Info.AbsOutPath, t.DefaultPath()), nil
	}
	return write_strategy.NewCreateFileStrategy(t.Info.AbsOutPath, t.DefaultPath()), nil
}

func (t *mainTemplate) interruptHandler() *Statement {
	s := &Statement{}
	s.Comment(`InterruptHandler handles first SIGINT and SIGTERM and sends messages to error channel.`).Line()
	s.Func().Id("InterruptHandler").Params(Id("ch").Id("chan<- error")).Block(
		Id("interruptHandler").Op(":=").Id("make").Call(Id("chan").Qual(PackagePathOs, "Signal"), Lit(1)),
		Qual(PackagePathOsSignal, "Notify").Call(
			Id("interruptHandler"),
			Qual(PackagePathSyscall, "SIGINT"),
			Qual(PackagePathSyscall, "SIGTERM"),
		),
		Id("ch").Op("<-").Qual(PackagePathErrors, "New").Call(Parens(Op("<-").Id("interruptHandler")).Dot("String").Call()),
	)
	return s
}

func (t *mainTemplate) mainFunc() *Statement {
	return Func().Id("main").Call().BlockFunc(func(main *Group) {
		main.Id("logger").Op(":=").
			Qual(PackagePathGoKitLog, "With").Call(Id("InitLogger").Call(Qual(PackagePathOs, "Stdout")), Lit("level"), Lit("info"))
		if t.recovering {
			main.Id("errorLogger").Op(":=").
				Qual(PackagePathGoKitLog, "With").Call(Id("InitLogger").Call(Qual(PackagePathOs, "Stderr")), Lit("level"), Lit("error"))
		}
		main.Id("logger").Dot("Log").Call(Lit("message"), Lit("Hello, I am alive"))
		main.Defer().Id("logger").Dot("Log").Call(Lit("message"), Lit("goodbye, good luck"))
		main.Line()
		main.Id("errorChan").Op(":=").Make(Id("chan error"))
		main.Go().Id("InterruptHandler").Call(Id("errorChan"))
		main.Line()
		main.Id("service").Op(":=").Qual(t.Info.ServiceImportPath, constructorName(t.Info.Iface)).Call().
			Comment(`Create new service.`)
		if t.logging {
			main.Id("service").Op("=").
				Qual(filepath.Join(t.Info.ServiceImportPath, "middleware"), "ServiceLogging").Call(Id("logger")).Call(Id("service")).
				Comment(`Setup service logging.`)
		}
		if t.errorLogging {
			main.Id("service").Op("=").
				Qual(filepath.Join(t.Info.ServiceImportPath, "middleware"), "ServiceErrorLogging").Call(Id("logger")).Call(Id("service")).
				Comment(`Setup error logging.`)
		}
		if t.recovering {
			main.Id("service").Op("=").
				Qual(filepath.Join(t.Info.ServiceImportPath, "middleware"), "ServiceRecovering").Call(Id("errorLogger")).Call(Id("service")).
				Comment(`Setup service recovering.`)
		}
		main.Line().Id("endpoints").Op(":=").Qual(t.Info.ServiceImportPath, "AllEndpoints").Call(t.endpointsParams())
		if t.grpcServer {
			main.Line()
			main.Id("grpcAddr").Op(":=").Lit(":8081")
			main.Comment(`Start grpc server.`)
			main.Go().Id("ServeGRPC").Call(
				Id("endpoints"),
				Id("errorChan"),
				Id("grpcAddr"),
				Qual(PackagePathGoKitLog, "With").Call(Id("logger"), Lit("transport"), Lit("GRPC")),
			)
		}
		if t.httpServer {
			main.Line()
			main.Id("httpAddr").Op(":=").Lit(":8080")
			main.Comment(`Start http server.`)
			main.Go().Id("ServeHTTP").Call(
				Id("endpoints"),
				Id("errorChan"),
				Id("httpAddr"),
				Qual(PackagePathGoKitLog, "With").Call(Id("logger"), Lit("transport"), Lit("HTTP")),
			)
		}
		main.Line()
		main.Id("logger").Dot("Log").Call(Lit("error"), Op("<-").Id("errorChan"))
	})
}

// Renders something like this
//		func initLogger() {
//			logger = log.NewJSONLogger(os.Stdout)
//			logger = log.With(logger, "service", svchelper.NormalizeName(SERVICE_NAME))
//			logger = log.With(logger, "@timestamp", log.DefaultTimestamp)
//			logger = log.With(logger, "@message", "info")
//			logger = log.With(logger, "caller", log.DefaultCaller)
//
//			logger.Log("version", GitHash)
//			logger.Log("Build", Build)
//			logger.Log("msg", "hello")
//		}
func (t *mainTemplate) initLogger() *Statement {
	return Comment(`InitLogger initialize go-kit JSON logger with timestamp and caller.`).Line().
		Func().Id("InitLogger").Params(Id("writer").Qual(PackagePathIO, "Writer")).Params(Qual(PackagePathGoKitLog, "Logger")).BlockFunc(func(body *Group) {
		body.Id("logger").Op(":=").Qual(PackagePathGoKitLog, "NewJSONLogger").Call(Id("writer"))
		body.Id("logger").Op("=").Qual(PackagePathGoKitLog, "With").Call(Id("logger"), Lit("@timestamp"), Qual(PackagePathGoKitLog, "DefaultTimestampUTC"))
		body.Id("logger").Op("=").Qual(PackagePathGoKitLog, "With").Call(Id("logger"), Lit("caller"), Qual(PackagePathGoKitLog, "DefaultCaller"))
		body.Return(Id("logger"))
	})
}

// Renders something like this
//		func serveGRPC(endpoints *clientsvc.Endpoints, errCh chan error) {
// 			logger := log.With(logger, "transport", "grpc")
// 			listener, err := net.Listen("tcp", *bindAddr)
// 			if err != nil {
// 				errCh <- err
// 				return
// 			}
//
// 			srv := transportgrpc.NewGRPCServer(endpoints)
// 			grpcs := grpc.NewServer()
// 			pb.RegisterClientServiceServer(grpcs, srv)
//
// 			logger.Log("addr", *bindAddr)
// 			errCh <- grpcs.Serve(listener)
// 		}
func (t *mainTemplate) serveGrpc() *Statement {
	if !t.grpcServer {
		return nil
	}
	return Comment(`ServeGRPC starts new GRPC server on address and sends first error to channel.`).Line().
		Func().Id("ServeGRPC").Params(
		Id("endpoints").Op("*").Qual(t.Info.ServiceImportPath, "Endpoints"),
		Id("ch").Id("chan<- error"),
		Id("addr").Id("string"),
		Id("logger").Qual(PackagePathGoKitLog, "Logger"),
	).BlockFunc(func(body *Group) {
		body.List(Id("listener"), Err()).Op(":=").Qual(PackagePathNet, "Listen").Call(Lit("tcp"), Id("addr"))
		body.If(Err().Op("!=").Nil()).Block(
			Id("ch").Op("<-").Err(),
			Return(),
		)
		body.Comment(`Here you can add middlewares for grpc server.`)
		body.Id("server").Op(":=").Qual(filepath.Join(t.Info.ServiceImportPath, "transport/grpc"), "NewGRPCServer").Call(t.newServerParams())
		body.Id("grpcServer").Op(":=").Qual(PackagePathGoogleGRPC, "NewServer").Call()
		body.Qual(t.Info.ProtobufPackage, "Register"+util.ToUpperFirst(t.Info.Iface.Name)+"Server").Call(Id("grpcServer"), Id("server"))
		body.Id("logger").Dot("Log").Call(Lit("listen on"), Id("addr"))
		body.Id("ch").Op("<-").Id("grpcServer").Dot("Serve").Call(Id("listener"))
	})
}

func (t *mainTemplate) serveHTTP() *Statement {
	if !t.httpServer {
		return nil
	}
	return Comment(`ServeHTTP starts new HTTP server on address and sends first error to channel.`).Line().
		Func().Id("ServeHTTP").Params(
		Id("endpoints").Op("*").Qual(t.Info.ServiceImportPath, "Endpoints"),
		Id("ch").Id("chan<- error"),
		Id("addr").Id("string"),
		Id("logger").Qual(PackagePathGoKitLog, "Logger"),
	).BlockFunc(func(body *Group) {
		body.Id("handler").Op(":=").Qual(t.Info.ServiceImportPath+"/transport/http", "NewHTTPHandler").Call(t.newServerParams())
		body.Id("httpServer").Op(":=").Op("&").Qual(PackagePathHttp, "Server").Values(DictFunc(func(d Dict) {
			d[Id("Addr")] = Id("addr")
			d[Id("Handler")] = Id("handler")
		}))
		body.Id("logger").Dot("Log").Call(Lit("listen on"), Id("addr"))
		body.Id("ch").Op("<-").Id("httpServer").Dot("ListenAndServe").Call()
	})
}

func (t *mainTemplate) endpointsParams() *Statement {
	s := &Statement{}
	s.Id("service")
	if t.tracing {
		s.Op(",").Line().Nil().Op(",").Comment("TODO: Add tracer").Line()
	}
	return s
}

func (t *mainTemplate) newServerParams() *Statement {
	s := &Statement{}
	s.Id("endpoints")
	if t.tracing {
		s.Op(",").Line().Id("logger")
	}
	if t.tracing {
		s.Op(",").Line().Nil().Op(",").Comment("TODO: Add tracer").Line()
	}
	return s
}

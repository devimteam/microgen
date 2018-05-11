package template

import (
	"os"
	"path/filepath"

	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/generator/write_strategy"
	"github.com/devimteam/microgen/logger"
	"github.com/devimteam/microgen/util"
)

const (
	nameInterruptHandler = "InterruptHandler"
	nameMain             = "main"
	nameInitLogger       = "InitLogger"
	nameServeGRPC        = "ServeGRPC"
	nameServeHTTP        = "ServeHTTP"
)

type mainTemplate struct {
	Info *GenerationInfo

	logging      bool
	recovering   bool
	errorLogging bool
	grpcServer   bool
	httpServer   bool
	tracing      bool

	rendered []string
	state    WriteStrategyState
}

func NewMainTemplate(info *GenerationInfo) Template {
	return &mainTemplate{
		Info: info,
	}
}

func (t *mainTemplate) Render() write_strategy.Renderer {
	f := &Statement{}
	f.Line().Add(t.mainFunc())
	f.Line().Add(t.initLogger())
	f.Line().Add(t.interruptHandler())
	f.Line().Add(t.serveGrpc())
	f.Line().Add(t.serveHTTP())

	if t.state == AppendStrat {
		return f
	}

	file := NewFile("main")
	file.PackageComment(t.Info.FileHeader)
	file.PackageComment(`Microgen appends missed functions.`)
	file.Add(f)

	return file
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
	if err := util.StatFile(t.Info.AbsOutputFilePath, t.DefaultPath()); os.IsNotExist(err) {
		t.state = FileStrat
		return write_strategy.NewCreateFileStrategy(t.Info.AbsOutputFilePath, t.DefaultPath()), nil
	}
	file, err := util.ParseFile(filepath.Join(t.Info.AbsOutputFilePath, t.DefaultPath()))
	if err != nil {
		logger.Logger.Logln(0, "can't parse", t.DefaultPath(), ":", err)
		return write_strategy.NewNopStrategy("", ""), nil
	}
	for _, f := range file.Functions {
		t.rendered = append(t.rendered, f.Name)
	}
	t.state = AppendStrat
	return write_strategy.NewAppendToFileStrategy(t.Info.AbsOutputFilePath, t.DefaultPath()), nil
}

func (t *mainTemplate) interruptHandler() *Statement {
	if util.IsInStringSlice(nameInterruptHandler, t.rendered) {
		return nil
	}
	s := &Statement{}
	s.Comment(nameInterruptHandler + ` handles first SIGINT and SIGTERM and sends messages to error channel.`).Line()
	s.Func().Id(nameInterruptHandler).Params(Id("ch").Id("chan<- error")).Block(
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
	if util.IsInStringSlice(nameMain, t.rendered) {
		return nil
	}
	return Func().Id(nameMain).Call().BlockFunc(func(main *Group) {
		main.Id("logger").Op(":=").
			Qual(PackagePathGoKitLog, "With").Call(Id(nameInitLogger).Call(Qual(PackagePathOs, "Stdout")), Lit("level"), Lit("info"))
		if t.recovering {
			main.Id("errorLogger").Op(":=").
				Qual(PackagePathGoKitLog, "With").Call(Id(nameInitLogger).Call(Qual(PackagePathOs, "Stderr")), Lit("level"), Lit("error"))
		}
		main.Id("logger").Dot("Log").Call(Lit("message"), Lit("Hello, I am alive"))
		main.Defer().Id("logger").Dot("Log").Call(Lit("message"), Lit("goodbye, good luck"))
		main.Line()
		main.Id("errorChan").Op(":=").Make(Id("chan error"))
		main.Go().Id(nameInterruptHandler).Call(Id("errorChan"))
		main.Line()
		main.Id("service").Op(":=").Qual(t.Info.SourcePackageImport, constructorName(t.Info.Iface)).Call().
			Comment(`Create new service.`)
		if t.logging {
			main.Id("service").Op("=").
				Qual(filepath.Join(t.Info.SourcePackageImport, "middleware"), "ServiceLogging").Call(Id("logger")).Call(Id("service")).
				Comment(`Setup service logging.`)
		}
		if t.errorLogging {
			main.Id("service").Op("=").
				Qual(filepath.Join(t.Info.SourcePackageImport, "middleware"), "ServiceErrorLogging").Call(Id("logger")).Call(Id("service")).
				Comment(`Setup error logging.`)
		}
		if t.recovering {
			main.Id("service").Op("=").
				Qual(filepath.Join(t.Info.SourcePackageImport, "middleware"), "ServiceRecovering").Call(Id("errorLogger")).Call(Id("service")).
				Comment(`Setup service recovering.`)
		}
		main.Line().Id("endpoints").Op(":=").Qual(t.Info.SourcePackageImport, "AllEndpoints").Call(t.endpointsParams())
		if t.grpcServer {
			main.Line()
			main.Id("grpcAddr").Op(":=").Lit(":8081")
			main.Comment(`Start grpc server.`)
			main.Go().Id(nameServeGRPC).Call(
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
			main.Go().Id(nameServeHTTP).Call(
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
//		func InitLogger(writer io.Writer) log.Logger {
//			logger := log.NewJSONLogger(writer)
//			logger = log.With(logger, "@timestamp", log.DefaultTimestampUTC)
//			logger = log.With(logger, "caller", log.DefaultCaller)
//			return logger
//		}
func (t *mainTemplate) initLogger() *Statement {
	if util.IsInStringSlice(nameInitLogger, t.rendered) {
		return nil
	}
	return Comment(nameInitLogger + ` initialize go-kit JSON logger with timestamp and caller.`).Line().
		Func().Id(nameInitLogger).Params(Id("writer").Qual(PackagePathIO, "Writer")).Params(Qual(PackagePathGoKitLog, "Logger")).BlockFunc(func(body *Group) {
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
	if !t.grpcServer || util.IsInStringSlice(nameServeGRPC, t.rendered) {
		return nil
	}
	return Comment(nameServeGRPC+` starts new GRPC server on address and sends first error to channel.`).Line().
		Func().Id(nameServeGRPC).Params(
		Id("endpoints").Op("*").Qual(t.Info.SourcePackageImport, "Endpoints"),
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
		body.Id("server").Op(":=").Qual(filepath.Join(t.Info.SourcePackageImport, "transport/grpc"), "NewGRPCServer").Call(t.newServerParams())
		body.Id("grpcServer").Op(":=").Qual(PackagePathGoogleGRPC, "NewServer").Call()
		body.Qual(t.Info.ProtobufPackageImport, "Register"+util.ToUpperFirst(t.Info.Iface.Name)+"Server").Call(Id("grpcServer"), Id("server"))
		body.Id("logger").Dot("Log").Call(Lit("listen on"), Id("addr"))
		body.Id("ch").Op("<-").Id("grpcServer").Dot("Serve").Call(Id("listener"))
	})
}

func (t *mainTemplate) serveHTTP() *Statement {
	if !t.httpServer || util.IsInStringSlice(nameServeHTTP, t.rendered) {
		return nil
	}
	return Comment(nameServeHTTP+` starts new HTTP server on address and sends first error to channel.`).Line().
		Func().Id(nameServeHTTP).Params(
		Id("endpoints").Op("*").Qual(t.Info.SourcePackageImport, "Endpoints"),
		Id("ch").Id("chan<- error"),
		Id("addr").Id("string"),
		Id("logger").Qual(PackagePathGoKitLog, "Logger"),
	).BlockFunc(func(body *Group) {
		body.Id("handler").Op(":=").Qual(t.Info.SourcePackageImport+"/transport/http", "NewHTTPHandler").Call(t.newServerParams())
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
		s.Op(",").Line().Qual(PackagePathOpenTracingGo, "NoopTracer{}").Op(",").Comment("TODO: Add tracer").Line()
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
		s.Op(",").Line().Qual(PackagePathOpenTracingGo, "NoopTracer{}").Op(",").Comment("TODO: Add tracer").Line()
	}
	return s
}

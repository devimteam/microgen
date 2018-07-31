package template

import (
	"path/filepath"

	"context"

	. "github.com/dave/jennifer/jen"
	mstrings "github.com/devimteam/microgen/generator/strings"
	"github.com/devimteam/microgen/generator/write_strategy"
)

const (
	nameInterruptHandler = "InterruptHandler"
	nameMain             = "main"
	nameInitLogger       = "InitLogger"
	nameServeGRPC        = "ServeGRPC"
	nameServeHTTP        = "ServeHTTP"
)

const (
	_service_ = "svc"
	_logger_  = "logger"
	_ctx_     = "ctx"
)

type mainTemplate struct {
	Info     *GenerationInfo
	rendered []string
	state    WriteStrategyState
}

func NewMainTemplate(info *GenerationInfo) Template {
	return &mainTemplate{
		Info: info,
	}
}

func (t *mainTemplate) Render(ctx context.Context) write_strategy.Renderer {
	f := &Statement{}
	f.Line().Add(t.mainFunc(ctx))
	f.Line().Add(t.initLogger())
	f.Line().Add(t.interruptHandler())
	f.Line().Add(t.serveGrpc(ctx))
	f.Line().Add(t.serveHTTP(ctx))

	if t.state == AppendStrat {
		return f
	}

	file := NewFile("main")
	file.PackageComment(`Microgen appends missed functions.`)
	file.Add(f)

	return file
}

func (t *mainTemplate) DefaultPath() string {
	return filepath.Join("./", PathExecutable, mstrings.ToSnakeCase(t.Info.Iface.Name), "/main.go")
}

func (t *mainTemplate) Prepare(ctx context.Context) error {
	return nil
}

func (t *mainTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.Info.OutputFilePath, t.DefaultPath()), nil
	/*if err := statFile(t.Info.OutputFilePath, t.DefaultPath()); os.IsNotExist(err) {
		t.state = FileStrat
		return write_strategy.NewCreateFileStrategy(t.Info.OutputFilePath, t.DefaultPath()), nil
	}
	file, err := parsePackage(filepath.Join(t.Info.OutputFilePath, t.DefaultPath()))
	if err != nil {
		logger.Logger.Logln(0, "can't parse", t.DefaultPath(), ":", err)
		return write_strategy.NewNopStrategy("", ""), nil
	}
	for _, f := range file.Functions {
		t.rendered = append(t.rendered, f.Name)
	}
	t.state = AppendStrat
	return write_strategy.NewAppendToFileStrategy(t.Info.OutputFilePath, t.DefaultPath()), nil*/
}

func (t *mainTemplate) interruptHandler() *Statement {
	if mstrings.IsInStringSlice(nameInterruptHandler, t.rendered) {
		return nil
	}
	s := &Statement{}
	s.Comment(nameInterruptHandler + ` handles first SIGINT and SIGTERM and returns it as error.`).Line()
	s.Func().Id(nameInterruptHandler).Params(ctx_contextContext).Params(Error()).Block(
		Id("interruptHandler").Op(":=").Id("make").Call(Id("chan").Qual(PackagePathOs, "Signal"), Lit(1)),
		Qual(PackagePathOsSignal, "Notify").Call(
			Id("interruptHandler"),
			Qual(PackagePathSyscall, "SIGINT"),
			Qual(PackagePathSyscall, "SIGTERM"),
		),
		Select().Block(
			Case(Id("sig").Op(":= <-").Id("interruptHandler")),
			Return().Qual(PackagePathFmt, "Errorf").Call(Lit("signal received: %v"), Id("sig").Dot("String").Call()),
			Case(Op("<-").Id(_ctx_).Dot("Done").Call()),
			Return().Qual(PackagePathErrors, "New").Call(Lit("signal listener: context canceled")),
		),
	)
	return s
}

func (t *mainTemplate) mainFunc(ctx context.Context) *Statement {
	if mstrings.IsInStringSlice(nameMain, t.rendered) {
		return nil
	}
	return Func().Id(nameMain).Call().BlockFunc(func(main *Group) {
		main.Id(_logger_).Op(":=").
			Qual(PackagePathGoKitLog, "With").Call(Id(nameInitLogger).Call(Qual(PackagePathOs, "Stdout")), Lit("level"), Lit("info"))
		if Tags(ctx).Has(RecoveringMiddlewareTag) {
			main.Id("errorLogger").Op(":=").
				Qual(PackagePathGoKitLog, "With").Call(Id(nameInitLogger).Call(Qual(PackagePathOs, "Stderr")), Lit("level"), Lit("error"))
		}
		main.Id(_logger_).Dot("Log").Call(Lit("message"), Lit("Hello, I am alive"))
		main.Defer().Id(_logger_).Dot("Log").Call(Lit("message"), Lit("goodbye, good luck"))
		main.Line()
		main.List(Id("g"), Id(_ctx_)).Op(":=").Qual(PackagePathSyncErrgroup, "WithContext").Call(Qual(PackagePathContext, "Background").Call())
		main.Id("g").Dot("Go").Call(
			Func().Params().Params(Error()).Block(
				Return().Id(nameInterruptHandler).Call(Id(_ctx_)),
			),
		)
		main.Line()
		main.Var().Id(_service_).Qual(t.Info.SourcePackageImport, t.Info.Iface.Name).Comment("// TODO:").Op("=").Qual(t.Info.OutputPackageImport+"/service", constructorName(t.Info.Iface)).Call().
			Comment(`Create new service.`)
		//if Tags(ctx).Has(CachingMiddlewareTag) {
		//	main.Id(_service_).Op("=").
		//		Qual(filepath.Join(t.Info.SourcePackageImport, PathService), CachingMiddlewareName).Call(Id("errorLogger")).Call(Id(_service_)).
		//		Comment(`Setup service caching.`)
		//}
		if Tags(ctx).Has(LoggingMiddlewareTag) {
			main.Id(_service_).Op("=").
				Qual(filepath.Join(t.Info.OutputPackageImport, PathService), ServiceLoggingMiddlewareName).Call(Id(_logger_)).Call(Id(_service_)).
				Comment(`Setup service logging.`)
		}
		if Tags(ctx).Has(ErrorLoggingMiddlewareTag) {
			main.Id(_service_).Op("=").
				Qual(filepath.Join(t.Info.OutputPackageImport, PathService), ServiceErrorLoggingMiddlewareName).Call(Id(_logger_)).Call(Id(_service_)).
				Comment(`Setup error logging.`)
		}
		if Tags(ctx).Has(RecoveringMiddlewareTag) {
			main.Id(_service_).Op("=").
				Qual(filepath.Join(t.Info.OutputPackageImport, PathService), ServiceRecoveringMiddlewareName).Call(Id("errorLogger")).Call(Id(_service_)).
				Comment(`Setup service recovering.`)
		}
		main.Line().Id("endpoints").Op(":=").Qual(t.Info.OutputPackageImport+"/transport", "Endpoints").Call(t.endpointsParams(ctx))
		if Tags(ctx).HasAny(TracingMiddlewareTag) {
			main.Id("endpoints").Op("=").Qual(t.Info.OutputPackageImport+"/transport", "TraceServerEndpoints").Call(
				Id("endpoints"),
				Qual(PackagePathOpenTracingGo, "NoopTracer{}"),
			).Comment("TODO: Add tracer")
		}
		if Tags(ctx).HasAny(GrpcTag, GrpcServerTag) {
			main.Line()
			main.Id("grpcAddr").Op(":=").Lit(":8081").Comment("TODO: use normal address")
			main.Comment(`Start grpc server.`)
			main.Id("g").Dot("Go").Call(
				Func().Params().Params(Error()).Block(
					Return().Id(nameServeGRPC).Call(
						Id(_ctx_),
						Op("&").Id("endpoints"),
						Id("grpcAddr"),
						Qual(PackagePathGoKitLog, "With").Call(Id(_logger_), Lit("transport"), Lit("GRPC")),
					),
				),
			)
		}
		if Tags(ctx).HasAny(HttpTag, HttpServerTag) {
			main.Line()
			main.Id("httpAddr").Op(":=").Lit(":8080").Comment("TODO: use normal address")
			main.Comment(`Start http server.`)
			main.Id("g").Dot("Go").Call(
				Func().Params().Params(Error()).Block(
					Return().Id(nameServeHTTP).Call(
						Id(_ctx_),
						Op("&").Id("endpoints"),
						Id("httpAddr"),
						Qual(PackagePathGoKitLog, "With").Call(Id(_logger_), Lit("transport"), Lit("HTTP")),
					),
				),
			)
		}
		main.Line()
		main.If(Err().Op(":=").Id("g").Dot("Wait").Call(), Err().Op("!=").Nil()).Block(
			Id(_logger_).Dot("Log").Call(Lit("error"), Err()),
		)
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
	if mstrings.IsInStringSlice(nameInitLogger, t.rendered) {
		return nil
	}
	return Comment(nameInitLogger + ` initialize go-kit JSON logger with timestamp and caller.`).Line().
		Func().Id(nameInitLogger).Params(Id("writer").Qual(PackagePathIO, "Writer")).Params(Qual(PackagePathGoKitLog, "Logger")).BlockFunc(func(body *Group) {
		body.Id(_logger_).Op(":=").Qual(PackagePathGoKitLog, "NewJSONLogger").Call(Id("writer"))
		body.Id(_logger_).Op("=").Qual(PackagePathGoKitLog, "With").Call(Id(_logger_), Lit("@timestamp"), Qual(PackagePathGoKitLog, "DefaultTimestampUTC"))
		body.Id(_logger_).Op("=").Qual(PackagePathGoKitLog, "With").Call(Id(_logger_), Lit("caller"), Qual(PackagePathGoKitLog, "DefaultCaller"))
		body.Return(Id(_logger_))
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
func (t *mainTemplate) serveGrpc(ctx context.Context) *Statement {
	if !Tags(ctx).HasAny(GrpcTag, GrpcServerTag) || mstrings.IsInStringSlice(nameServeGRPC, t.rendered) {
		return nil
	}
	return Comment(nameServeGRPC+` starts new GRPC server on address and sends first error to channel.`).Line().
		Func().Id(nameServeGRPC).Params(
		ctx_contextContext,
		Id("endpoints").Op("*").Qual(filepath.Join(t.Info.OutputPackageImport, "transport"), EndpointsSetName),
		Id("addr").Id("string"),
		Id(_logger_).Qual(PackagePathGoKitLog, "Logger"),
	).Params(
		Error(),
	).BlockFunc(func(body *Group) {
		body.List(Id("listener"), Err()).Op(":=").Qual(PackagePathNet, "Listen").Call(Lit("tcp"), Id("addr"))
		body.If(Err().Op("!=").Nil()).Block(
			Return().Err(),
		)
		body.Comment(`Here you can add middlewares for grpc server.`)
		body.Id("server").Op(":=").Qual(filepath.Join(t.Info.OutputPackageImport, "transport/grpc"), "NewGRPCServer").Call(t.newServerParams(ctx))
		body.Id("grpcServer").Op(":=").Qual(PackagePathGoogleGRPC, "NewServer").Call()
		body.Qual(t.Info.ProtobufPackageImport, "Register"+mstrings.ToUpperFirst(t.Info.Iface.Name)+"Server").Call(Id("grpcServer"), Id("server"))
		body.Id(_logger_).Dot("Log").Call(Lit("listen on"), Id("addr"))
		body.Id("ch").Op(":=").Make(Id("chan error"))
		body.Go().Func().Call().Block(
			Id("ch").Op("<-").Id("grpcServer").Dot("Serve").Call(Id("listener")),
		).Call()
		body.Select().Block(
			Case(Err().Op(":= <-").Id("ch")),
			Return().Qual(PackagePathFmt, "Errorf").Call(Lit("grpc server: serve: %v"), Err()),
			Case(Op("<-").Id(_ctx_).Dot("Done").Call()),
			Id("grpcServer").Dot("GracefulStop").Call(),
			Return().Qual(PackagePathErrors, "New").Call(Lit("grpc server: context canceled")),
		)
	})
}

func (t *mainTemplate) serveHTTP(ctx context.Context) *Statement {
	if !Tags(ctx).HasAny(HttpTag, HttpServerTag) || mstrings.IsInStringSlice(nameServeHTTP, t.rendered) {
		return nil
	}
	return Comment(nameServeHTTP+` starts new HTTP server on address and sends first error to channel.`).Line().
		Func().Id(nameServeHTTP).Params(
		ctx_contextContext,
		Id("endpoints").Op("*").Qual(t.Info.OutputPackageImport+"/transport", EndpointsSetName),
		Id("addr").Id("string"),
		Id(_logger_).Qual(PackagePathGoKitLog, "Logger"),
	).Params(
		Error(),
	).BlockFunc(func(body *Group) {
		body.Id("handler").Op(":=").Qual(t.Info.OutputPackageImport+"/transport/http", "NewHTTPHandler").Call(t.newServerParams(ctx))
		body.Id("httpServer").Op(":=").Op("&").Qual(PackagePathHttp, "Server").Values(DictFunc(func(d Dict) {
			d[Id("Addr")] = Id("addr")
			d[Id("Handler")] = Id("handler")
		}))
		body.Id(_logger_).Dot("Log").Call(Lit("listen on"), Id("addr"))
		body.Id("ch").Op(":=").Make(Id("chan error"))
		body.Go().Func().Call().Block(
			Id("ch").Op("<-").Id("httpServer").Dot("ListenAndServe").Call(),
		).Call()
		body.Select().Block(
			Case(Err().Op(":= <-").Id("ch")),
			If(Err().Op("==").Qual(PackagePathHttp, "ErrServerClosed")).Block(
				Return().Nil(),
			),
			Return().Qual(PackagePathFmt, "Errorf").Call(Lit("http server: serve: %v"), Err()),
			Case(Op("<-").Id(_ctx_).Dot("Done").Call()),
			Return().Id("httpServer").Dot("Shutdown").Call(Qual(PackagePathContext, "Background").Call()),
		)
	})
}

func (t *mainTemplate) endpointsParams(ctx context.Context) *Statement {
	s := &Statement{}
	s.Id(_service_)
	/*if Tags(ctx).HasAny(TracingMiddlewareTag) {
		s.Op(",").Line().Qual(PackagePathOpenTracingGo, "NoopTracer{}").Op(",").Comment("TODO: Add tracer").Line()
	}*/
	return s
}

func (t *mainTemplate) newServerParams(ctx context.Context) *Statement {
	s := &Statement{}
	s.Id("endpoints")
	if Tags(ctx).HasAny(TracingMiddlewareTag) {
		s.Op(",").Line().Id(_logger_)
	}
	if Tags(ctx).HasAny(TracingMiddlewareTag) {
		s.Op(",").Line().Qual(PackagePathOpenTracingGo, "NoopTracer{}").Op(",").Comment("TODO: Add tracer").Line()
	}
	return s
}

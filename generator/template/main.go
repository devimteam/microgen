package template

import (
	"path/filepath"

	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/generator/write_strategy"
	"github.com/devimteam/microgen/util"
)

type mainTemplate struct {
	Info *GenerationInfo

	logging    bool
	recovering bool
	grpcServer bool
	httpServer bool
}

func NewMainTemplate(info *GenerationInfo) Template {
	return &mainTemplate{
		Info: info,
	}
}

func (t *mainTemplate) Render() write_strategy.Renderer {
	f := NewFile("main")
	f.PackageComment(FileHeader)
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
		main.Id("logger").Op(":=").Id("InitLogger").Call()
		main.Defer().Id("logger").Dot("Log").Call(Lit("goodbye"), Lit("good luck"))
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
		if t.recovering {
			main.Id("service").Op("=").
				Qual(filepath.Join(t.Info.ServiceImportPath, "middleware"), "ServiceRecovering").Call(Id("logger")).Call(Id("service")).
				Comment(`Setup service recovering.`)
		}
		main.Line()
		main.Id("endpoints").Op(":=").Op("&").Qual(t.Info.ServiceImportPath, "Endpoints").Values(DictFunc(func(p Dict) {
			for _, method := range t.Info.Iface.Methods {
				p[Id(endpointStructName(method.Name))] = Qual(t.Info.ServiceImportPath, endpointStructName(method.Name)).Call(Id("service"))
			}
		}))
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
		Func().Id("InitLogger").Params().Params(Qual(PackagePathGoKitLog, "Logger")).BlockFunc(func(body *Group) {
		body.Id("logger").Op(":=").Qual(PackagePathGoKitLog, "NewJSONLogger").Call(Qual(PackagePathOs, "Stdout"))
		body.Id("logger").Op("=").Qual(PackagePathGoKitLog, "With").Call(Id("logger"), Lit("@when"), Qual(PackagePathGoKitLog, "DefaultTimestampUTC"))
		body.Id("logger").Op("=").Qual(PackagePathGoKitLog, "With").Call(Id("logger"), Lit("@where"), Qual(PackagePathGoKitLog, "DefaultCaller"))
		body.Id("logger").Dot("Log").Call(Lit("hello"), Lit("I am alive"))
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
		body.Id("server").Op(":=").Qual(filepath.Join(t.Info.ServiceImportPath, "transport/grpc"), "NewGRPCServer").Call(Id("endpoints"))
		body.Id("grpcs").Op(":=").Qual(PackagePathGoogleGRPC, "NewServer").Call()
		body.Qual(t.Info.ProtobufPackage, "Register"+util.ToUpperFirst(t.Info.Iface.Name)+"Server").Call(Id("grpcs"), Id("server"))
		body.Id("logger").Dot("Log").Call(Lit("listen on"), Id("addr"))
		body.Id("ch").Op("<-").Id("grpcs").Dot("Serve").Call(Id("listener"))
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
		body.Id("handler").Op(":=").Qual(t.Info.ServiceImportPath+"/transport/http", "NewHTTPHandler").Call(Id("endpoints"))
		body.Id("https").Op(":=").Op("&").Qual(PackagePathHttp, "Server").Values(DictFunc(func(d Dict) {
			d[Id("Addr")] = Id("addr")
			d[Id("Handler")] = Id("handler")
		}))
		body.Id("logger").Dot("Log").Call(Lit("listen on"), Id("addr"))
		body.Id("ch").Op("<-").Id("https").Dot("ListenAndServe").Call()
	})
}

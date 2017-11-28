package template

import (
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

	f.Add(t.mainFunc())
	f.Add(t.interruptHandler())

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

	})
}
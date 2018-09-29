package plugins

import (
	"flag"
	"microgen/gen"
	"microgen/pkg/microgen"

	mstrings "github.com/devimteam/microgen/generator/strings"
)

const loggingPlugin = "logging"

type loggingMiddlewarePlugin struct {
	Name string
}

func (p *loggingMiddlewarePlugin) Generate(ctx microgen.Context, args ...string) (microgen.Context, error) {
	// parse args
	flags := flag.NewFlagSet(loggingPlugin, flag.ExitOnError)
	flags.StringVar(&p.Name, "name", "LoggingMiddleware", "")
	err := flags.Parse(args)
	if err != nil {
		return ctx, err
	}

	outfile := microgen.File{}

	// normalize args
	p.Name = mstrings.ToUpperFirst(p.Name)

	f := gen.NewFile()
	f.Wln(ctx.FileHeader)
	f.Wln()
	f.Wln(`package `, ctx.SourcePackageName)
	f.Wln()
	f.Wln(`import (
	log "github.com/go-kit/kit/log"
	)`)
	f.Wln(`func `, p.Name, `(logger`)

	outfile.Name = loggingPlugin
	outfile.Path = "./service/logging_microgen.go"
	outfile.Content = f.Bytes()
	return ctx, nil
}

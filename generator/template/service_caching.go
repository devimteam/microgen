package template

import (
	"context"

	. "github.com/dave/jennifer/jen"
	mstrings "github.com/devimteam/microgen/generator/strings"
	"github.com/devimteam/microgen/generator/write_strategy"
	"github.com/vetcher/go-astra/types"
)

const (
	cacheKeyTag = "cache-key"

	cacheInterfaceName          = "Cache"
	cachingMiddlewareStructName = "cachingMiddleware"
)

var CachingMiddlewareName = mstrings.ToUpperFirst(cachingMiddlewareStructName)

type cacheMiddlewareTemplate struct {
	info      *GenerationInfo
	cacheKeys map[string]string
	caching   map[string]bool
}

func NewCacheMiddlewareTemplate(info *GenerationInfo) Template {
	return &cacheMiddlewareTemplate{
		info: info,
	}
}

func (t *cacheMiddlewareTemplate) Render(ctx context.Context) write_strategy.Renderer {
	f := &Statement{}
	// Render type Cache
	f.Comment("Cache interface uses for middleware as key-value storage for requests.")
	f.Line().Type().Id(cacheInterfaceName).Interface(
		Id("Set").Call(Op("key, value interface{}")).Call(Op("err error")),
		Id("Get").Call(Op("key interface{}")).Call(Op("value interface{}, err error")),
	)
	f.Line()

	f.Line().Func().Id(CachingMiddlewareName).Params(Id("cache").Id(cacheInterfaceName)).Params(Id(MiddlewareTypeName)).
		Block(t.newCacheBody(t.info.Iface))

	f.Line()

	// Render middleware struct
	f.Type().Id(cachingMiddlewareStructName).Struct(
		Id("cache").Id(cacheInterfaceName),
		Id(_logger_).Qual(PackagePathGoKitLog, "Logger"),
		Id(_next_).Qual(t.info.SourcePackageImport, t.info.Iface.Name),
	)
	for _, signature := range t.info.Iface.Methods {
		f.Line()
		f.Add(t.cacheFunc(ctx, signature)).Line()
	}
	for _, signature := range t.info.Iface.Methods {
		if !t.info.AllowedMethods[signature.Name] {
			continue
		}
		f.Add(cacheEntity(ctx, signature)).Line()
	}

	file := NewFile("service")
	file.ImportAlias(t.info.SourcePackageImport, serviceAlias)
	file.HeaderComment(t.info.FileHeader)
	file.Add(f)
	return file
}

func (cacheMiddlewareTemplate) DefaultPath() string {
	return filenameBuilder(PathService, "caching")
}

func (t *cacheMiddlewareTemplate) Prepare(ctx context.Context) error {
	t.cacheKeys = make(map[string]string)
	t.caching = make(map[string]bool)
	for _, method := range t.info.Iface.Methods {
		if mstrings.HasTag(method.Docs, TagMark+CachingMiddlewareTag) {
			t.caching[method.Name] = true
			t.cacheKeys[method.Name] = `"` + method.Name + `"`
		}
		if s := mstrings.FetchTags(method.Docs, TagMark+cacheKeyTag); len(s) > 0 {
			t.cacheKeys[method.Name] = s[0]
			t.caching[method.Name] = true
		}
	}
	return nil
}

func (t *cacheMiddlewareTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
}

func (t *cacheMiddlewareTemplate) newCacheBody(i *types.Interface) *Statement {
	return Return(Func().Params(
		Id(_next_).Qual(t.info.SourcePackageImport, i.Name),
	).Params(
		Qual(t.info.SourcePackageImport, i.Name),
	).BlockFunc(func(g *Group) {
		g.Return(Op("&").Id(cachingMiddlewareStructName).Values(
			Dict{
				Id("cache"): Id("cache"),
				Id(_next_):  Id(_next_),
			},
		))
	}))
}

func (t *cacheMiddlewareTemplate) cacheFunc(ctx context.Context, signature *types.Function) *Statement {
	normalized := normalizeFunctionResults(signature)
	return methodDefinition(ctx, cachingMiddlewareStructName, &normalized.Function).
		BlockFunc(t.cacheFuncBody(signature, &normalized.Function))
}

func (t *cacheMiddlewareTemplate) cacheFuncBody(signature *types.Function, normalized *types.Function) func(g *Group) {
	return func(g *Group) {
		if !t.info.AllowedMethods[signature.Name] {
			s := &Statement{}
			if len(normalized.Results) > 0 {
				s.Return()
			}
			s.Id(rec(cachingMiddlewareStructName)).Dot(_next_).Dot(signature.Name).Call(paramNames(normalized.Args))
			g.Add(s)
			return
		}
		if t.caching[signature.Name] {
			g.List(Id("value"), Id("e")).Op(":=").Id(rec(cachingMiddlewareStructName)).Dot("cache").Dot("Get").Call(Id(t.cacheKeys[signature.Name]))
			g.If(Id("e").Op("==").Nil()).Block(
				ReturnFunc(func(group *Group) {
					for _, field := range removeErrorIfLast(signature.Results) {
						group.Id("value").Assert(Op("*").Id(cacheEntityStructName(normalized))).Op(".").Add(structFieldName(&field))
					}
					group.Id(nameOfLastResultError(normalized))
				}),
			)
			g.Defer().Func().Params().Block(
				Id(rec(cachingMiddlewareStructName)).Dot("cache").Dot("Set").Call(
					Id(t.cacheKeys[signature.Name]),
					Op("&").Id(cacheEntityStructName(normalized)).Values(dictByNormalVariables(
						removeErrorIfLast(signature.Results),
						removeErrorIfLast(normalized.Results),
					)),
				),
			).Call()
		}
		g.Return().Id(rec(cachingMiddlewareStructName)).Dot(_next_).Dot(signature.Name).Call(paramNames(normalized.Args))
	}
}

func cacheEntityStructName(signature *types.Function) string {
	return mstrings.ToLowerFirst(responseStructName(signature) + "CacheEntity")
}

func cacheEntity(ctx context.Context, signature *types.Function) *Statement {
	s := &Statement{}
	s.Type().Id(cacheEntityStructName(signature)).StructFunc(func(l *Group) {
		for _, field := range removeErrorIfLast(signature.Results) {
			l.Add(structFieldName(&field)).Add(fieldType(ctx, field.Type, false))
		}
	})
	return s
}

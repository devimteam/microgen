package template

import (
	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/generator/write_strategy"
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/godecl/types"
)

const (
	cacheKeyTag = "cache-key"

	cacheInterfaceName        = "Cache"
	cacheMiddlewareStructName = "serviceCache"
)

type cacheMiddlewareTemplate struct {
	Info      *GenerationInfo
	cacheKeys map[string]string
	caching   map[string]bool
}

func NewCacheMiddlewareTemplate(info *GenerationInfo) Template {
	return &cacheMiddlewareTemplate{
		Info: info,
	}
}

func (t *cacheMiddlewareTemplate) Render() write_strategy.Renderer {
	f := NewFile("middleware")
	f.PackageComment(t.Info.FileHeader)
	f.PackageComment(`Please, do not edit.`)

	// Render type Cache
	f.Comment("Cache interface uses for middleware as key-value storage for requests.")
	f.Type().Id(cacheInterfaceName).Interface(
		Id("Set").Call(Op("key, value interface{}")).Call(Op("err error")),
		Id("Get").Call(Op("key interface{}")).Call(Op("value interface{}, err error")),
	)
	f.Line()

	f.Line().Func().Id(util.ToUpperFirst(cacheMiddlewareStructName)).Params(Id("cache").Id(cacheInterfaceName)).Params(Id(MiddlewareTypeName)).
		Block(t.newCacheBody(t.Info.Iface))

	f.Line()

	// Render middleware struct
	f.Type().Id(cacheMiddlewareStructName).Struct(
		Id("cache").Id(cacheInterfaceName),
		Id(loggerVarName).Qual(PackagePathGoKitLog, "Logger"),
		Id(nextVarName).Qual(t.Info.ServiceImportPath, t.Info.Iface.Name),
	)
	for _, signature := range t.Info.Iface.Methods {
		f.Line()
		f.Add(t.cacheFunc(signature)).Line()
	}
	for _, signature := range t.Info.Iface.Methods {
		if t.caching[signature.Name] {
			f.Add(t.cacheEntity(signature)).Line()
		}
	}
	return f
}

func (cacheMiddlewareTemplate) DefaultPath() string {
	return "./middleware/cache.go"
}

func (t *cacheMiddlewareTemplate) Prepare() error {
	t.cacheKeys = make(map[string]string)
	t.caching = make(map[string]bool)
	for _, method := range t.Info.Iface.Methods {
		if util.HasTag(method.Docs, TagMark+CacheTag) {
			t.caching[method.Name] = true
			t.cacheKeys[method.Name] = `"` + method.Name + `"`
		}
		if s := util.FetchTags(method.Docs, TagMark+cacheKeyTag); len(s) > 0 {
			t.cacheKeys[method.Name] = s[0]
			t.caching[method.Name] = true
		}
	}
	return nil
}

func (t *cacheMiddlewareTemplate) ChooseStrategy() (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.Info.AbsOutPath, t.DefaultPath()), nil
}

func (t *cacheMiddlewareTemplate) newCacheBody(i *types.Interface) *Statement {
	return Return(Func().Params(
		Id(nextVarName).Qual(t.Info.ServiceImportPath, i.Name),
	).Params(
		Qual(t.Info.ServiceImportPath, i.Name),
	).BlockFunc(func(g *Group) {
		g.Return(Op("&").Id(cacheMiddlewareStructName).Values(
			Dict{
				Id("cache"):     Id("cache"),
				Id(nextVarName): Id(nextVarName),
			},
		))
	}))
}

func (t *cacheMiddlewareTemplate) cacheFunc(signature *types.Function) *Statement {
	normalized := normalizeFunctionResults(signature)
	return methodDefinition(cacheMiddlewareStructName, &normalized.Function).
		BlockFunc(t.cacheFuncBody(signature, &normalized.Function))
}

func (t *cacheMiddlewareTemplate) cacheFuncBody(signature *types.Function, normalized *types.Function) func(g *Group) {
	return func(g *Group) {
		if t.caching[signature.Name] {
			g.List(Id("value"), Id("e")).Op(":=").Id(util.LastUpperOrFirst(cacheMiddlewareStructName)).Dot("cache").Dot("Get").Call(Id(t.cacheKeys[signature.Name]))
			g.If(Id("e").Op("==").Nil()).Block(
				ReturnFunc(func(group *Group) {
					for _, field := range removeErrorIfLast(signature.Results) {
						group.Id("value").Assert(Op("*").Id(cacheEntityStructName(normalized))).Op(".").Add(structFieldName(&field))
					}
					group.Id(nameOfLastResultError(normalized))
				}),
			)
			g.Defer().Func().Params().Block(
				Id(util.LastUpperOrFirst(cacheMiddlewareStructName)).Dot("cache").Dot("Set").Call(
					Id(t.cacheKeys[signature.Name]),
					Op("&").Id(cacheEntityStructName(normalized)).Values(dictByNormalVariables(
						removeErrorIfLast(signature.Results),
						removeErrorIfLast(normalized.Results),
					)),
				),
			).Call()
		}
		g.Return().Id(util.LastUpperOrFirst(cacheMiddlewareStructName)).Dot(nextVarName).Dot(signature.Name).Call(paramNames(normalized.Args))
	}
}

func cacheEntityStructName(signature *types.Function) string {
	return util.ToLowerFirst(responseStructName(signature) + "CacheEntity")
}

func (t *cacheMiddlewareTemplate) cacheEntity(signature *types.Function) *Statement {
	s := &Statement{}
	s.Type().Id(cacheEntityStructName(signature)).StructFunc(func(l *Group) {
		for _, field := range removeErrorIfLast(signature.Results) {
			l.Add(structFieldName(&field)).Add(fieldType(field.Type, false))
		}
	})
	return s
}

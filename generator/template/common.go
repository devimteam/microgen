package template

import (
	"strconv"
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/godecl/types"
)

const (
	PackagePathGoKitEndpoint         = "github.com/go-kit/kit/endpoint"
	PackagePathContext               = "context"
	PackagePathGoKitLog              = "github.com/go-kit/kit/log"
	PackagePathTime                  = "time"
	PackagePathGoogleGRPC            = "google.golang.org/grpc"
	PackagePathGoogleGRPCStatus      = "google.golang.org/grpc/status"
	PackagePathGoogleGRPCCodes       = "google.golang.org/grpc/codes"
	PackagePathNetContext            = "golang.org/x/net/context"
	PackagePathGoKitTransportGRPC    = "github.com/go-kit/kit/transport/grpc"
	PackagePathHttp                  = "net/http"
	PackagePathGoKitTransportHTTP    = "github.com/go-kit/kit/transport/http"
	PackagePathBytes                 = "bytes"
	PackagePathJson                  = "encoding/json"
	PackagePathIOUtil                = "io/ioutil"
	PackagePathIO                    = "io"
	PackagePathStrings               = "strings"
	PackagePathUrl                   = "net/url"
	PackagePathEmptyProtobuf         = "github.com/golang/protobuf/ptypes/empty"
	PackagePathFmt                   = "fmt"
	PackagePathOs                    = "os"
	PackagePathOsSignal              = "os/signal"
	PackagePathSyscall               = "syscall"
	PackagePathErrors                = "errors"
	PackagePathNet                   = "net"
	PackagePathGorillaMux            = "github.com/gorilla/mux"
	PackagePathPath                  = "path"
	PackagePathStrconv               = "strconv"
	PackagePathOpenTracingGo         = "github.com/opentracing/opentracing-go"
	PackagePathGoKitTracing          = "github.com/go-kit/kit/tracing/opentracing"
	PackagePathGoKitTransportJSONRPC = "github.com/go-kit/kit/transport/http/jsonrpc"

	TagMark         = "// @"
	MicrogenMainTag = "microgen"
	ForceTag        = "force"
)

const (
	MiddlewareTag             = "middleware"
	LoggingMiddlewareTag      = "logging"
	RecoverMiddlewareTag      = "recover"
	HttpTag                   = "http"
	HttpServerTag             = "http-server"
	HttpClientTag             = "http-client"
	GrpcTag                   = "grpc"
	GrpcServerTag             = "grpc-server"
	GrpcClientTag             = "grpc-client"
	MainTag                   = "main"
	ErrorLoggingMiddlewareTag = "error-logging"
	TracingTag                = "tracing"
	CacheTag                  = "cache"
	JSONRPCTag                = "json-rpc"
	JSONRPCServerTag          = "json-rpc-server"
	JSONRPCClientTag          = "json-rpc-client"
)

type WriteStrategyState int

const (
	FileStrat WriteStrategyState = iota + 1
	AppendStrat
)

type GenerationInfo struct {
	ServiceImportPackageName string
	Iface                    *types.Interface
	ServiceImportPath        string
	Force                    bool
	AbsOutPath               string
	SourceFilePath           string
	FileHeader               string

	ProtobufPackage string
	GRPCRegAddr     string
}

func (info GenerationInfo) Copy() *GenerationInfo {
	return &GenerationInfo{
		Iface: info.Iface,
		Force: info.Force,
		ServiceImportPackageName: info.ServiceImportPackageName,
		ServiceImportPath:        info.ServiceImportPath,
		AbsOutPath:               info.AbsOutPath,
		SourceFilePath:           info.SourceFilePath,

		GRPCRegAddr:     info.GRPCRegAddr,
		ProtobufPackage: info.ProtobufPackage,
		FileHeader:      info.FileHeader,
	}
}

func structFieldName(field *types.Variable) *Statement {
	return Id(util.ToUpperFirst(field.Name))
}

// Remove from function fields context if it is first in slice
func RemoveContextIfFirst(fields []types.Variable) []types.Variable {
	if IsContextFirst(fields) {
		return fields[1:]
	}
	return fields
}

func IsContextFirst(fields []types.Variable) bool {
	if len(fields) == 0 {
		return false
	}
	name := types.TypeName(fields[0].Type)
	return name != nil &&
		types.TypeImport(fields[0].Type) != nil &&
		types.TypeImport(fields[0].Type).Package == PackagePathContext &&
		*name == "Context"
}

// Remove from function fields error if it is last in slice
func removeErrorIfLast(fields []types.Variable) []types.Variable {
	if IsErrorLast(fields) {
		return fields[:len(fields)-1]
	}
	return fields
}

func IsErrorLast(fields []types.Variable) bool {
	if len(fields) == 0 {
		return false
	}
	name := types.TypeName(fields[len(fields)-1].Type)
	return name != nil &&
		types.TypeImport(fields[len(fields)-1].Type) == nil &&
		*name == "error"
}

// Return name of error, if error is last result, else return `err`
func nameOfLastResultError(fn *types.Function) string {
	if IsErrorLast(fn.Results) {
		return fn.Results[len(fn.Results)-1].Name
	}
	return "err"
}

// Renders struct field.
//
//  	Visit *entity.Visit `json:"visit"`
//
func structField(field *types.Variable) *Statement {
	s := structFieldName(field)
	s.Add(fieldType(field.Type, false))
	s.Tag(map[string]string{"json": util.ToSnakeCase(field.Name)})
	if types.IsEllipsis(field.Type) {
		s.Comment("This field was defined with ellipsis (...).")
	}
	return s
}

// Renders func params for definition.
//
//  	visit *entity.Visit, err error
//
func funcDefinitionParams(fields []types.Variable) *Statement {
	c := &Statement{}
	c.ListFunc(func(g *Group) {
		for _, field := range fields {
			g.Id(util.ToLowerFirst(field.Name)).Add(fieldType(field.Type, true))
		}
	})
	return c
}

// Renders field type for given func field.
//
//  	*repository.Visit
//
func fieldType(field types.Type, useEllipsis bool) *Statement {
	c := &Statement{}
	for field != nil {
		switch f := field.(type) {
		case types.TImport:
			if f.Import != nil {
				c.Qual(f.Import.Package, "")
			}
			field = f.Next
		case types.TName:
			c.Id(f.TypeName)
			field = nil
		case types.TArray:
			if f.IsSlice {
				c.Index()
			} else if f.ArrayLen > 0 {
				c.Index(Lit(f.ArrayLen))
			}
			field = f.Next
		case types.TMap:
			return c.Map(fieldType(f.Key, false)).Add(fieldType(f.Value, false))
		case types.TPointer:
			c.Op(strings.Repeat("*", f.NumberOfPointers))
			field = f.Next
		case types.TInterface:
			mhds := interfaceType(f.Interface)
			return c.Interface(mhds...)
		case types.TEllipsis:
			if useEllipsis {
				c.Op("...")
			} else {
				c.Index()
			}
			field = f.Next
		default:
			return c
		}
	}
	return c
}

func interfaceType(p *types.Interface) (code []Code) {
	for _, x := range p.Methods {
		code = append(code, functionDefinition(x))
	}
	return
}

// Renders key/value pairs wrapped in Dict for provided fields.
//
//		Err:    err,
//		Result: result,
//
func dictByVariables(fields []types.Variable) Dict {
	return DictFunc(func(d Dict) {
		for _, field := range fields {
			d[structFieldName(&field)] = Id(util.ToLowerFirst(field.Name))
		}
	})
}

// Render list of function receivers by signature.Result.
//
//		Ans1, ans2, AnS3 -> ans1, ans2, anS3
//
func paramNames(fields []types.Variable) *Statement {
	var list []Code
	for _, field := range fields {
		v := Id(util.ToLowerFirst(field.Name))
		if types.IsEllipsis(field.Type) {
			v.Op("...")
		}
		list = append(list, v)
	}
	return List(list...)
}

// Render full method definition with receiver, method name, args and results.
//
//		func (e *Endpoints) Count(ctx context.Context, text string, symbol string) (count int)
//
func methodDefinition(obj string, signature *types.Function) *Statement {
	return Func().
		Params(Id(util.LastUpperOrFirst(obj)).Op("*").Id(obj)).
		Add(functionDefinition(signature))
}

// Render full method definition with receiver, method name, args and results.
//
//		func Count(ctx context.Context, text string, symbol string) (count int)
//
func functionDefinition(signature *types.Function) *Statement {
	return Id(signature.Name).
		Params(funcDefinitionParams(signature.Args)).
		Params(funcDefinitionParams(signature.Results))
}

func defaultNameFormer(f *types.Function) string {
	return f.Name
}

// Remove from generating functions that already in existing.
func removeAlreadyExistingFunctions(existing []types.Function, generating *[]*types.Function, nameFormer func(*types.Function) string) {
	x := (*generating)[:0]
	for _, fn := range *generating {
		if f := util.FindFunctionByName(existing, nameFormer(fn)); f == nil {
			x = append(x, fn)
		}
	}
	*generating = x
}

type normalizedFunction struct {
	types.Function
	parent *types.Function
}

const (
	normalArgPrefix    = "arg"
	normalResultPrefix = "res"
)

func normalizeFunction(signature *types.Function) *normalizedFunction {
	newFunc := &normalizedFunction{parent: signature}
	newFunc.Name = signature.Name
	newFunc.Args = normalizeVariables(signature.Args, normalArgPrefix)
	newFunc.Results = normalizeVariables(signature.Results, normalResultPrefix)
	return newFunc
}

func normalizeFunctionArgs(signature *types.Function) *normalizedFunction {
	newFunc := &normalizedFunction{parent: signature}
	newFunc.Name = signature.Name
	newFunc.Args = normalizeVariables(signature.Args, normalArgPrefix)
	newFunc.Results = signature.Results
	return newFunc
}

func normalizeFunctionResults(signature *types.Function) *normalizedFunction {
	newFunc := &normalizedFunction{parent: signature}
	newFunc.Name = signature.Name
	newFunc.Args = signature.Args
	newFunc.Results = normalizeVariables(signature.Results, normalResultPrefix)
	return newFunc
}

func normalizeVariables(old []types.Variable, prefix string) (new []types.Variable) {
	for i := range old {
		v := old[i]
		v.Name = prefix + strconv.Itoa(i)
		new = append(new, v)
	}
	return
}

func dictByNormalVariables(fields []types.Variable, normals []types.Variable) Dict {
	if len(fields) != len(normals) {
		panic("len of fields and normals not the same")
	}
	return DictFunc(func(d Dict) {
		for i, field := range fields {
			d[structFieldName(&field)] = Id(util.ToLowerFirst(normals[i].Name))
		}
	})
}

type Rendered struct {
	slice []string
}

func (r *Rendered) Add(s string) {
	r.slice = append(r.slice, s)
}

func (r *Rendered) Contain(s string) bool {
	return util.IsInStringSlice(s, r.slice)
}

func (r *Rendered) NotContain(s string) bool {
	return !r.Contain(s)
}

// Hard
// Soft
// Nop?

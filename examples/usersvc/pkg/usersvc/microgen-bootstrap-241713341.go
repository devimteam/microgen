// +build microgen-ignore

// TEMPORARY microgen FILE. DO NOT EDIT.

package main

import (
	// List of imported plugins
	pkg "github.com/devimteam/microgen/examples/usersvc/pkg/usersvc"
	microgen "github.com/devimteam/microgen/pkg/microgen"
	_ "github.com/devimteam/microgen/pkg/plugins"
	"reflect"
)

func main() {
	microgen.RegisterPackage("github.com/devimteam/microgen/examples/usersvc/pkg/usersvc")
	targetInterface := microgen.Interface{
		Name: "UserService",
		Docs: []string{"//microgen", "// @microgen middleware, logging, grpc, http, recovering, error-logging, tracing, caching, metrics", "// @protobuf github.com/devimteam/microgen/examples/protobuf"},
		Methods: []microgen.Method{
			microgen.Method{
				Name:    "CreateUser",
				Docs:    []string{},
				Args:    []microgen.Var{{Name: "ctx"}, {Name: "user"}},
				Results: []microgen.Var{{Name: "id"}, {Name: "err"}},
			},
			microgen.Method{
				Name:    "UpdateUser",
				Docs:    []string{},
				Args:    []microgen.Var{{Name: "ctx"}, {Name: "user"}},
				Results: []microgen.Var{{Name: "err"}},
			},
			microgen.Method{
				Name:    "GetUser",
				Docs:    []string{},
				Args:    []microgen.Var{{Name: "ctx"}, {Name: "id"}},
				Results: []microgen.Var{{Name: "user"}, {Name: "err"}},
			},
			microgen.Method{
				Name:    "FindUsers",
				Docs:    []string{},
				Args:    []microgen.Var{{Name: "ctx"}},
				Results: []microgen.Var{{Name: "results"}, {Name: "err"}},
			},
			microgen.Method{
				Name:    "CreateComment",
				Docs:    []string{},
				Args:    []microgen.Var{{Name: "ctx"}, {Name: "comment"}},
				Results: []microgen.Var{{Name: "id"}, {Name: "err"}},
			},
			microgen.Method{
				Name:    "GetComment",
				Docs:    []string{},
				Args:    []microgen.Var{{Name: "ctx"}, {Name: "id"}},
				Results: []microgen.Var{{Name: "comment"}, {Name: "err"}},
			},
			microgen.Method{
				Name:    "GetUserComments",
				Docs:    []string{},
				Args:    []microgen.Var{{Name: "ctx"}, {Name: "userId"}},
				Results: []microgen.Var{{Name: "list"}, {Name: "err"}},
			},
		}}
	// Add reflect data
	targetInterface.Type = reflect.TypeOf(new(pkg.UserService)).Elem()
	// CreateUser
	targetInterface.Methods[0].Type = methodToType(targetInterface.Type.MethodByName("CreateUser"))
	targetInterface.Methods[0].Args[0].Type = targetInterface.Methods[0].Type.In(0)     // ctx
	targetInterface.Methods[0].Args[1].Type = targetInterface.Methods[0].Type.In(1)     // user
	targetInterface.Methods[0].Results[0].Type = targetInterface.Methods[0].Type.Out(0) // id
	targetInterface.Methods[0].Results[1].Type = targetInterface.Methods[0].Type.Out(1) // err
	// UpdateUser
	targetInterface.Methods[1].Type = methodToType(targetInterface.Type.MethodByName("UpdateUser"))
	targetInterface.Methods[1].Args[0].Type = targetInterface.Methods[1].Type.In(0)     // ctx
	targetInterface.Methods[1].Args[1].Type = targetInterface.Methods[1].Type.In(1)     // user
	targetInterface.Methods[1].Results[0].Type = targetInterface.Methods[1].Type.Out(0) // err
	// GetUser
	targetInterface.Methods[2].Type = methodToType(targetInterface.Type.MethodByName("GetUser"))
	targetInterface.Methods[2].Args[0].Type = targetInterface.Methods[2].Type.In(0)     // ctx
	targetInterface.Methods[2].Args[1].Type = targetInterface.Methods[2].Type.In(1)     // id
	targetInterface.Methods[2].Results[0].Type = targetInterface.Methods[2].Type.Out(0) // user
	targetInterface.Methods[2].Results[1].Type = targetInterface.Methods[2].Type.Out(1) // err
	// FindUsers
	targetInterface.Methods[3].Type = methodToType(targetInterface.Type.MethodByName("FindUsers"))
	targetInterface.Methods[3].Args[0].Type = targetInterface.Methods[3].Type.In(0)     // ctx
	targetInterface.Methods[3].Results[0].Type = targetInterface.Methods[3].Type.Out(0) // results
	targetInterface.Methods[3].Results[1].Type = targetInterface.Methods[3].Type.Out(1) // err
	// CreateComment
	targetInterface.Methods[4].Type = methodToType(targetInterface.Type.MethodByName("CreateComment"))
	targetInterface.Methods[4].Args[0].Type = targetInterface.Methods[4].Type.In(0)     // ctx
	targetInterface.Methods[4].Args[1].Type = targetInterface.Methods[4].Type.In(1)     // comment
	targetInterface.Methods[4].Results[0].Type = targetInterface.Methods[4].Type.Out(0) // id
	targetInterface.Methods[4].Results[1].Type = targetInterface.Methods[4].Type.Out(1) // err
	// GetComment
	targetInterface.Methods[5].Type = methodToType(targetInterface.Type.MethodByName("GetComment"))
	targetInterface.Methods[5].Args[0].Type = targetInterface.Methods[5].Type.In(0)     // ctx
	targetInterface.Methods[5].Args[1].Type = targetInterface.Methods[5].Type.In(1)     // id
	targetInterface.Methods[5].Results[0].Type = targetInterface.Methods[5].Type.Out(0) // comment
	targetInterface.Methods[5].Results[1].Type = targetInterface.Methods[5].Type.Out(1) // err
	// GetUserComments
	targetInterface.Methods[6].Type = methodToType(targetInterface.Type.MethodByName("GetUserComments"))
	targetInterface.Methods[6].Args[0].Type = targetInterface.Methods[6].Type.In(0)     // ctx
	targetInterface.Methods[6].Args[1].Type = targetInterface.Methods[6].Type.In(1)     // userId
	targetInterface.Methods[6].Results[0].Type = targetInterface.Methods[6].Type.Out(0) // list
	targetInterface.Methods[6].Results[1].Type = targetInterface.Methods[6].Type.Out(1) // err

	microgen.RegisterInterface(targetInterface)
	microgen.Exec("-debug", "-config=micro.toml", "-keep", "UserService")
}
func methodToType(m reflect.Method, ok bool) reflect.Type {
	return m.Type
}

// +build microgen-ignore

// TEMPORARY microgen FILE. DO NOT EDIT.

package main

import (
	// List of imported plugins
	pkg "github.com/devimteam/microgen/examples/usersvc/pkg/usersvc"
	microgen "github.com/devimteam/microgen/pkg/microgen"
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
				Args:    []string{"ctx", "user"},
				Results: []string{"id", "err"},
			},
			microgen.Method{
				Name:    "UpdateUser",
				Docs:    []string{},
				Args:    []string{"ctx", "user"},
				Results: []string{"err"},
			},
			microgen.Method{
				Name:    "GetUser",
				Docs:    []string{},
				Args:    []string{"ctx", "id"},
				Results: []string{"user", "err"},
			},
			microgen.Method{
				Name:    "FindUsers",
				Docs:    []string{},
				Args:    []string{"ctx"},
				Results: []string{"results", "err"},
			},
			microgen.Method{
				Name:    "CreateComment",
				Docs:    []string{},
				Args:    []string{"ctx", "comment"},
				Results: []string{"id", "err"},
			},
			microgen.Method{
				Name:    "GetComment",
				Docs:    []string{},
				Args:    []string{"ctx", "id"},
				Results: []string{"comment", "err"},
			},
			microgen.Method{
				Name:    "GetUserComments",
				Docs:    []string{},
				Args:    []string{"ctx", "userId"},
				Results: []string{"list", "err"},
			},
		}}
	targetInterface.Value = reflect.ValueOf(new(pkg.UserService)).Elem()
	microgen.RegisterInterface(targetInterface)
	microgen.Exec("-config=micro.toml", "-debug", "UserService")
}

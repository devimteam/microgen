// +build microgen-ignore

// TEMPORARY microgen FILE. DO NOT EDIT.

package main

import (
	// List of imported plugins
	_ "github.com/devimteam/microgen/pkg/plugins"

	pkg "github.com/devimteam/microgen/examples/usersvc/pkg/usersvc"
	microgen "github.com/devimteam/microgen/pkg/microgen"
)

func main() {
	microgen.RegisterInterface("UserService", pkg.UserService(nil))
	microgen.Run()
}

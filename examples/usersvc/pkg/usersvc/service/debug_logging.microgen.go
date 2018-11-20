// Code generated by microgen. DO NOT EDIT.

package service

import (
	"context"
	service "github.com/devimteam/microgen/examples/usersvc/pkg/usersvc"
	log "github.com/go-kit/kit/log"
)

//go:generate easyjson -all debug_logging.microgen.go

var _ service.UserService = &debugLogging{}

func DebugLogging(logger log.Logger) func(service.UserService) service.UserService {
	return func(next service.UserService) service.UserService {
		return &debugLogging{
			logger: logger,
			next:   next,
		}
	}
}

type debugLogging struct {
	logger log.Logger
	next   service.UserService
}

func (L debugLogging) CreateUser(arg_0 context.Context, arg_1 service.User) (res_0 string, res_1 error) {
	defer func() {
		L.logger.Log(
			"method", "CreateUser",
			"user", arg_1,
			"id", res_0,
			"err", res_1)
	}()
	return L.next.CreateUser(arg_0, arg_1)
}

func (L debugLogging) UpdateUser(arg_0 context.Context, arg_1 service.User) (res_0 error) {
	defer func() {
		L.logger.Log(
			"method", "UpdateUser",
			"user", arg_1,
			"err", res_0)
	}()
	return L.next.UpdateUser(arg_0, arg_1)
}

func (L debugLogging) GetUser(arg_0 context.Context, arg_1 string) (res_0 service.User, res_1 error) {
	defer func() {
		L.logger.Log(
			"method", "GetUser",
			"id", arg_1,
			"user", res_0,
			"err", res_1)
	}()
	return L.next.GetUser(arg_0, arg_1)
}

func (L debugLogging) FindUsers(arg_0 context.Context) (res_0 []*service.User, res_1 error) {
	defer func() {
		L.logger.Log(
			"method", "FindUsers",
			"results", res_0,
			"len(results)", len(res_0),
			"err", res_1)
	}()
	return L.next.FindUsers(arg_0)
}

func (L debugLogging) CreateComment(arg_0 context.Context, arg_1 service.Comment) (res_0 string, res_1 error) {
	defer func() {
		L.logger.Log(
			"method", "CreateComment",
			"comment", arg_1,
			"id", res_0,
			"err", res_1)
	}()
	return L.next.CreateComment(arg_0, arg_1)
}

func (L debugLogging) GetComment(arg_0 context.Context, arg_1 string) (res_0 service.Comment, res_1 error) {
	defer func() {
		L.logger.Log(
			"method", "GetComment",
			"id", arg_1,
			"comment", res_0,
			"err", res_1)
	}()
	return L.next.GetComment(arg_0, arg_1)
}

func (L debugLogging) GetUserComments(arg_0 context.Context, arg_1 string) (res_0 []service.Comment, res_1 error) {
	defer func() {
		L.logger.Log(
			"method", "GetUserComments",
			"userId", arg_1,
			"list", res_0,
			"err", res_1)
	}()
	return L.next.GetUserComments(arg_0, arg_1)
}
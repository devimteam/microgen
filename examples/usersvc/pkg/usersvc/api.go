package usersvc

import "context"

// @microgen middleware, logging, grpc, http, recovering, error-logging, tracing, caching, metrics
// @protobuf github.com/devimteam/microgen/examples/protobuf
type UserService interface {
	CreateUser(ctx context.Context, user User) (id string, err error)
	UpdateUser(ctx context.Context, user User) (err error)
	GetUser(ctx context.Context, id string) (user User, err error)
	FindUsers(ctx context.Context) (results map[string]User, err error)
	CreateComment(ctx context.Context, comment Comment) (id string, err error)
	GetComment(ctx context.Context, id string) (comment Comment, err error)
	GetUserComments(ctx context.Context, userId string) (list []Comment, err error)
}

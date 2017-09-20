package stringsvc

import (
	"context"

	"github.com/devimteam/microgen/example/svc/entity"
)

// @microgen middleware, logging, grpc
// @grpc-addr devim.string.team
// @protobuf github.com/devimteam/protobuf/stringsvc
type StringService interface {
	//!log ans, err
	Uppercase(ctx context.Context, str string) (ans string, err error)
	Count(ctx context.Context, text string, symbol string) (count int, positions []int)
	TestCase(ctx context.Context, comments []*entity.Comment) (tree map[string]int, err error)
}

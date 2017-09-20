package stringsvc

import (
	"context"
)

// @microgen middleware, logging, grpc
// @grpc-addr devim.string.team
// @protobuf gitlab.devim.team/protobuf/stringsvc
type StringService interface {
	Uppercase(ctx context.Context, str string) (ans string, err error)
	Count(ctx context.Context, text string, symbol string) (count int, positions []int)
	//TestCase(ctx context.Context, comments []*drive.Comment) (tree map[string]interface{}, err error)
}

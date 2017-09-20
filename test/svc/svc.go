package stringsvc

import (
	"context"

	drive "google.golang.org/api/drive/v2"
)

// @microgen middleware, logging, grpc
// @grpc-addr devim.string.team
// @protobuf gitlab.devim.team/protobuf/stringsvc
type StringService interface {
	//!log ans, err
	Uppercase(ctx context.Context, str string) (ans string, err error)
	Count(ctx context.Context, text string, symbol string) (count int, positions []int)
	TestCase(ctx context.Context, comments []*drive.Comment) (tree map[string]int, err error)
}

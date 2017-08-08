package visitsvc

import (
	"context"

	"google.golang.org/api/drive/v3"
)

type StringService interface {
	Uppercase(ctx context.Context, str string) (ans string, err error)
	Count(ctx context.Context, text string) (count int)
	TestCase(ctx context.Context, comments drive.Comment) (err error)
}

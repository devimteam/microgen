// This file was automatically generated by "microgen 0.7.0b" utility.
// Please, do not edit.
package middleware

import (
	context "context"
	generated "github.com/devimteam/microgen/example/generated"
	entity "github.com/devimteam/microgen/example/svc/entity"
	log "github.com/go-kit/kit/log"
	time "time"
)

// ServiceLogging writes params, results and working time of method call to provided logger after its execution.
func ServiceLogging(logger log.Logger) Middleware {
	return func(next generated.StringService) generated.StringService {
		return &serviceLogging{
			logger: logger,
			next:   next,
		}
	}
}

type serviceLogging struct {
	logger log.Logger
	next   generated.StringService
}

func (L *serviceLogging) Uppercase(ctx context.Context, stringsMap map[string]string) (ans string, err error) {
	defer func(begin time.Time) {
		L.logger.Log(
			"method", "Uppercase",
			"request", logUppercaseRequest{StringsMap: stringsMap},
			"took", time.Since(begin))
	}(time.Now())
	return L.next.Uppercase(ctx, stringsMap)
}

func (L *serviceLogging) Count(ctx context.Context, text string, symbol string) (count int, positions []int, err error) {
	defer func(begin time.Time) {
		L.logger.Log(
			"method", "Count",
			"request", logCountRequest{
				Symbol: symbol,
				Text:   text,
			},
			"response", logCountResponse{
				Count:     count,
				Positions: positions,
			},
			"error", err,
			"took", time.Since(begin))
	}(time.Now())
	return L.next.Count(ctx, text, symbol)
}

func (L *serviceLogging) TestCase(ctx context.Context, comments []*entity.Comment) (tree map[string]int, err error) {
	defer func(begin time.Time) {
		L.logger.Log(
			"method", "TestCase",
			"request", logTestCaseRequest{
				Comments:    comments,
				LenComments: len(comments),
			},
			"response", logTestCaseResponse{Tree: tree},
			"error", err,
			"took", time.Since(begin))
	}(time.Now())
	return L.next.TestCase(ctx, comments)
}

type logUppercaseRequest struct {
	StringsMap map[string]string
}

type logCountRequest struct {
	Text   string
	Symbol string
}

type logCountResponse struct {
	Count     int
	Positions []int
}

type logTestCaseRequest struct {
	Comments    []*entity.Comment
	LenComments int `json:"len(Comments)"`
}

type logTestCaseResponse struct {
	Tree map[string]int
}

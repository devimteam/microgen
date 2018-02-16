// This file was automatically generated by "microgen 0.8.0alpha" utility.
// Please, do not edit.
package middleware

import (
	context "context"
	generated "github.com/devimteam/microgen/example/generated"
	entity "github.com/devimteam/microgen/example/svc/entity"
	log "github.com/go-kit/kit/log"
)

// Cache interface uses for middleware as key-value storage for requests.
type Cache interface {
	Set(key, value interface{}) (err error)
	Get(key interface{}) (value interface{}, err error)
}

func ServiceCache(cache Cache) Middleware {
	return func(next generated.StringService) generated.StringService {
		return &serviceCache{
			cache: cache,
			next:  next,
		}
	}
}

type serviceCache struct {
	cache  Cache
	logger log.Logger
	next   generated.StringService
}

func (C *serviceCache) Uppercase(ctx context.Context, stringsMap map[string]string) (res0 string, res1 error) {
	value, e := C.cache.Get("Uppercase")
	if e == nil {
		return value.(*uppercaseResponseCacheEntity).Ans, res1
	}
	defer func() {
		C.cache.Set("Uppercase", &uppercaseResponseCacheEntity{Ans: res0})
	}()
	return C.next.Uppercase(ctx, stringsMap)
}

func (C *serviceCache) Count(ctx context.Context, text string, symbol string) (res0 int, res1 []int, res2 error) {
	value, e := C.cache.Get(text)
	if e == nil {
		return value.(*countResponseCacheEntity).Count, value.(*countResponseCacheEntity).Positions, res2
	}
	defer func() {
		C.cache.Set(text, &countResponseCacheEntity{
			Count:     res0,
			Positions: res1,
		})
	}()
	return C.next.Count(ctx, text, symbol)
}

func (C *serviceCache) TestCase(ctx context.Context, comments []*entity.Comment) (res0 map[string]int, res1 error) {
	return C.next.TestCase(ctx, comments)
}

type uppercaseResponseCacheEntity struct {
	Ans string
}

type countResponseCacheEntity struct {
	Count     int
	Positions []int
}

package pkg

import (
	"context"

	"github.com/devimteam/microgen/examples/clientsvc/pkg/entity"
)

//go:generate microgen -v=15 ClientService

type ListRequestFilters struct {
	Ids         []entity.UUID
	FirstName   *string
	LastName    *string
	InBlackList *bool
}

type ClientService interface {
	Create(ctx context.Context, client *entity.Client) (id entity.UUID, err error)
	Read(ctx context.Context, id entity.UUID) (client *entity.Client, err error)
	Update(ctx context.Context, client *entity.Client) (err error)
	Delete(ctx context.Context, ids ...entity.UUID) (err error)
	List(ctx context.Context, pag *Pagination, filters *ListRequestFilters) (clients []*entity.Client, err error)
}

type Pagination struct {
	Limit  *uint
	Offset uint
}

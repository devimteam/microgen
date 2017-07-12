package svc

import (
	"context"

	"gitlab.devim.team/microservices/workplacesvc/entity"
)

// This is an interface of the service.
// Yay
type WorkplaceService interface {
	CreateWorkplace(ctx context.Context, title string, teamViewerNumber string, pointOfSaleId string, equipmentId string) (workplace *entity.Workplace, err error)
	UpdateWorkplace(ctx context.Context, workplaceId string, workplace *entity.Workplace) (workplaceRes *entity.Workplace, err error)
	ReadWorkplace(ctx context.Context, id string) (workplace *entity.Workplace, err error)
	DeleteWorkplace(ctx context.Context, id string) (workplace *entity.Workplace, err error)
	ListWorkplace(ctx context.Context) (workplaces []*entity.Workplace, err error)
	ListWorkplaceByPointOfSale(ctx context.Context, pointOfSaleId string) (workplaces []*entity.Workplace, err error)
	IssueCertificate(ctx context.Context, workplaceId, certificateName string) (cert *entity.Certificate, err error)
	RevokeCertificate(ctx context.Context, workplaceId string) (cert *entity.Certificate, err error)
	CheckCertificate(ctx context.Context, serial uint64) (isValid bool, abc string, err error)
	ReissueCertificate(ctx context.Context, workplaceId, certificateName string) (cert *entity.Certificate, err error)
}

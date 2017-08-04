package visitsvc

import (
	"gitlab.devim.team/microservices/visitsvc/entity"
)

type VisitService interface {
	CreateVisit(visit *entity.Visit) (result *entity.Visit, err error)
}

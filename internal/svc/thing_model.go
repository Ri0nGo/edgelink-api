package svc

import "edgelink-api/internal/repo"

type IThingModelSvc interface {
	GetThingModelList()
}

type ThingModelSvc struct {
	tmRepo repo.IThingModelRepo
}

func (t *ThingModelSvc) GetThingModelList() {
	//TODO implement me
	panic("implement me")
}

func NewThingModelSvc(tmRepo repo.IThingModelRepo) IThingModelSvc {
	return &ThingModelSvc{
		tmRepo: tmRepo,
	}
}

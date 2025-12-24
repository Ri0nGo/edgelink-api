package repo

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type IThingModelRepo interface {
	GetThingModelList()
}

type ThingModelRepo struct {
	db  *gorm.DB
	cmd redis.Cmdable
}

func (t *ThingModelRepo) GetThingModelList() {
	//TODO implement me
	panic("implement me")
}

func NewThingModelRepo(db *gorm.DB, cmd redis.Cmdable) IThingModelRepo {
	return &ThingModelRepo{db: db, cmd: cmd}
}

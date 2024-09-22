package repository

import (
	"n1h41/apk_builder_v3/entity"
	"n1h41/apk_builder_v3/types"
)

type appRepo struct{}

func NewAppRepo() types.AppRepo {
	return &appRepo{}
}

func (a *appRepo) BuildApp(conf entity.BuildConfig) {
	panic("unimplemented")
}

package types

import "n1h41/apk_builder_v3/entity"

type AppRepo interface {
	BuildApp(entity.BuildConfig)
}

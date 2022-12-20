package orm

import (
	"github.com/jackycsl/geektime-go-practical/orm/internal/valuer"
	"github.com/jackycsl/geektime-go-practical/orm/model"
)

type core struct {
	model   *model.Model
	dialect Dialect
	creator valuer.Creator
	r       model.Registry
}

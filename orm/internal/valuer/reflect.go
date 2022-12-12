package valuer

import (
	"database/sql"
	"reflect"

	"github.com/jackycsl/geektime-go-practical/orm/internal/errs"
	"github.com/jackycsl/geektime-go-practical/orm/model"
)

type reflectValue struct {
	model *model.Model
	// 对应于 T 的指针
	// val any
	val reflect.Value
}

var _ Creator = NewReflectValue

func NewReflectValue(model *model.Model, val any) Value {
	return reflectValue{
		model: model,
		val:   reflect.ValueOf(val).Elem(),
	}
}

func (r reflectValue) SetColumns(rows *sql.Rows) error {
	// 在这里，继续处理结果集

	// 我怎么知道你 SELECT 出来了哪些列？
	// 拿到了 SELECT 出来的列
	cs, err := rows.Columns()
	if err != nil {
		return err
	}

	// 怎么利用 cs 来解决顺序问题和类型问题

	// 通过 cs 来构造 vals
	// 怎么构造呢？
	vals := make([]any, 0, len(cs))
	valElem := make([]reflect.Value, 0, len(cs))
	for _, c := range cs {
		fd, ok := r.model.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		// 反射创建一个实例
		// 这里创建的实例是原本类型的指针类型
		// 例如 fd.Type = int，那么val 是 *int
		val := reflect.New(fd.Typ)
		vals = append(vals, val.Interface())
		// 记得要调用 Elem，因为 fd.Type = int，那么val 是 *int
		valElem = append(valElem, val.Elem())
	}

	// 第一个问题：类型要匹配
	// 第二个问题：顺序要匹配

	// SELECT id, first_name, age, last_name
	// SELECT first_name, age, last_name, id
	err = rows.Scan(vals...)
	if err != nil {
		return err
	}

	tpValueElem := r.val
	for i, c := range cs {
		fd, ok := r.model.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		if fd.ColName == c {
			tpValueElem.FieldByName(fd.GoName).Set(valElem[i])
		}
	}

	return nil
}

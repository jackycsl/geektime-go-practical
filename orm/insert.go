package orm

import (
	"reflect"
	"strings"

	"github.com/jackycsl/geektime-go-practical/orm/internal/errs"
	"github.com/jackycsl/geektime-go-practical/orm/model"
)

type OnDuplicateKeyBuilder[T any] struct {
	i *Inserter[T]
}

type OnDuplicateKey[T any] struct {
	assigns []Assignable
}

func (o *OnDuplicateKeyBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.onDuplicateKey = &OnDuplicateKey[T]{
		assigns: assigns,
	}
	return o.i
}

type Assignable interface {
	assign()
}

type Inserter[T any] struct {
	values  []*T
	db      *DB
	columns []string

	// onDuplicateKey []Assignable
	onDuplicateKey *OnDuplicateKey[T]
}

func NewInserter[T any](db *DB) *Inserter[T] {
	return &Inserter[T]{
		db: db,
	}
}

// func (i *Inserter[T]) OnDuplicateKey(assigns ...Assignable) *Inserter[T] {
// 	i.onDuplicateKey = assigns
// 	return i
// }

func (i *Inserter[T]) OnDuplicateKey() *OnDuplicateKeyBuilder[T] {
	return &OnDuplicateKeyBuilder[T]{
		i: i,
	}
}

func (i *Inserter[T]) Values(vals ...*T) *Inserter[T] {
	i.values = vals
	return i
}

func (i *Inserter[T]) Columns(cols ...string) *Inserter[T] {
	i.columns = cols
	return i
}

func (i *Inserter[T]) Build() (*Query, error) {
	if len(i.values) == 0 {
		return nil, errs.ErrInsertZeroRow
	}
	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	m, err := i.db.r.Get(i.values[0])
	if err != nil {
		return nil, err
	}
	// 拼接表名
	sb.WriteByte('`')
	sb.WriteString(m.TableName)
	sb.WriteByte('`')
	// 一定要显示指定列的顺序，不然我们不知道数据库中默认的顺序
	// 我们要构造 `test_model`(col1, col2...)
	sb.WriteByte('(')

	fields := m.Fields
	if len(i.columns) > 0 {
		fields = make([]*model.Field, 0, len(i.columns))
		for _, fd := range i.columns {
			fdMeta, ok := m.FieldMap[fd]
			if !ok {
				return nil, errs.NewErrUnknownField(fd)
			}
			fields = append(fields, fdMeta)
		}
	}

	for idx, field := range fields {
		if idx > 0 {
			sb.WriteByte(',')
		}
		sb.WriteByte('`')
		sb.WriteString(field.ColName)
		sb.WriteByte('`')
	}
	sb.WriteByte(')')

	// 拼接 Values
	sb.WriteString(" VALUES ")
	// 预估的参数数量是：我有多少行乘以我有多少个字段
	args := make([]any, 0, len(i.values)*len(fields))
	for j, val := range i.values {
		if j > 0 {
			sb.WriteByte(',')
		}
		sb.WriteByte('(')
		for idx, field := range fields {
			if idx > 0 {
				sb.WriteByte(',')
			}
			sb.WriteByte('?')
			arg := reflect.ValueOf(val).Elem().FieldByName(field.GoName).Interface()
			args = append(args, arg)
		}
		sb.WriteByte(')')
	}

	if i.onDuplicateKey != nil {
		sb.WriteString(" ON DUPLICATE KEY UPDATE ")
		for idx, assign := range i.onDuplicateKey.assigns {
			if idx > 0 {
				sb.WriteByte(',')
			}
			switch a := assign.(type) {
			case Assignment:
				fd, ok := m.FieldMap[a.col]
				// 字段不对，或者说列不对
				if !ok {
					return nil, errs.NewErrUnknownField(a.col)
				}
				sb.WriteByte('`')
				sb.WriteString(fd.ColName)
				sb.WriteByte('`')
				sb.WriteString("=?")
				args = append(args, a.val)
			case Column:
				fd, ok := m.FieldMap[a.name]
				// 字段不对，或者说列不对
				if !ok {
					return nil, errs.NewErrUnknownField(a.name)
				}
				sb.WriteByte('`')
				sb.WriteString(fd.ColName)
				sb.WriteByte('`')
				sb.WriteString("=VALUES(")
				sb.WriteByte('`')
				sb.WriteString(fd.ColName)
				sb.WriteByte('`')
				sb.WriteByte(')')
			default:
				return nil, errs.NewErrUnsupportedAssignable(assign)
			}
		}
	}

	sb.WriteByte(';')
	return &Query{SQL: sb.String(), Args: args}, nil
}

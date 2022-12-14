package orm

import (
	"context"
	"strings"

	"github.com/jackycsl/geektime-go-practical/orm/internal/errs"
	"github.com/jackycsl/geektime-go-practical/orm/model"
)

// Selectable 是一个标记接口
// 它代表的是查找的列，或者聚合函数等
// SELECT XXX 部分
type Selectable interface {
	selectable()
}

type Selector[T any] struct {
	table   string
	model   *model.Model
	where   []Predicate
	sb      *strings.Builder
	args    []any
	columns []Selectable

	db *DB
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		sb: &strings.Builder{},
		db: db,
	}
}

func (s *Selector[T]) Build() (*Query, error) {
	s.sb = &strings.Builder{}
	var err error
	s.model, err = s.db.r.Get(new(T))
	if err != nil {
		return nil, err
	}
	sb := s.sb
	sb.WriteString("SELECT ")

	if err = s.buildColumns(); err != nil {
		return nil, err
	}

	sb.WriteString(" FROM ")
	// 我怎么把表名拿到
	if s.table == "" {
		sb.WriteByte('`')
		sb.WriteString(s.model.TableName)
		sb.WriteByte('`')
	} else {
		// segs := strings.Split(s.table, ".")
		// sb.WriteByte('`')
		// sb.WriteString(segs[0])
		// sb.WriteByte('`')
		// sb.WriteByte('.')
		// sb.WriteByte('`')
		// sb.WriteString(segs[1])
		// sb.WriteByte('`')
		sb.WriteString(s.table)
	}

	if len(s.where) > 0 {
		sb.WriteString(" WHERE ")
		p := s.where[0]
		for i := 1; i < len(s.where); i++ {
			p = p.And(s.where[i])
		}
		if err := s.buildExpression(p); err != nil {
			return nil, err
		}
	}

	sb.WriteByte(';')
	return &Query{
		SQL:  sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) buildExpression(expr Expression) error {
	switch exp := expr.(type) {
	case nil:
	case Predicate:
		// 在这里处理 p
		// p.left 构建好
		// p.op 构建好
		// p.right 构建好
		_, ok := exp.left.(Predicate)
		if ok {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(exp.left); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')
		}

		s.sb.WriteByte(' ')
		s.sb.WriteString(exp.op.String())
		s.sb.WriteByte(' ')

		_, ok = exp.right.(Predicate)
		if ok {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(exp.right); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')
		}
	case Column:
		return s.buildColumn(exp.name)
	case value:
		s.sb.WriteByte('?')
		s.addArg(exp.val)
	default:
		return errs.NewErrUnsupportedExpression(expr)
	}
	return nil
}

func (s *Selector[T]) buildColumns() error {
	if len(s.columns) == 0 {
		// 没有指定列
		s.sb.WriteByte('*')
		return nil
	}

	if len(s.columns) > 0 {
		for i, col := range s.columns {
			if i > 0 {
				s.sb.WriteByte(',')
			}
			switch c := col.(type) {
			case Column:
				err := s.buildColumn(c.name)
				if err != nil {
					return err
				}
			case Aggregate:
				// 聚合函数名
				s.sb.WriteString(c.fn)
				s.sb.WriteByte('(')
				err := s.buildColumn(c.arg)
				if err != nil {
					return err
				}
				s.sb.WriteByte(')')
				// 聚合函数本身的别名
			}

		}
	} else {
		s.sb.WriteByte('*')
	}
	return nil
}

func (s *Selector[T]) buildColumn(c string) error {
	fd, ok := s.model.FieldMap[c]
	// 字段不对，或者说列不对
	if !ok {
		return errs.NewErrUnknownField(c)
	}
	s.sb.WriteByte('`')
	s.sb.WriteString(fd.ColName)
	s.sb.WriteByte('`')
	return nil
}

func (s *Selector[T]) addArg(val any) {
	if s.args == nil {
		s.args = make([]any, 0, 8)
	}
	s.args = append(s.args, val)
}

// 这种也是可行
// s.Select("first_name,last_name")
// func (s *Selector[T]) SelectV1(cols string) *Selector[T] {
// 	return s
// }

// func (s *Selector[T]) Select(cols ...string) *Selector[T] {
// 	s.columns = cols
// 	return s
// }

func (s *Selector[T]) Select(cols ...Selectable) *Selector[T] {
	s.columns = cols
	return s
}

func (s *Selector[T]) From(table string) *Selector[T] {
	s.table = table
	return s
}

// ids :=[]int{1,2,3,}
// s.Where("id in (?, ?, ?)", ids...)
// golint-ci
// func (s *Selector[T]) Where(query string, args ...any) *Selector[T] {

// }

func (s *Selector[T]) Where(ps ...Predicate) *Selector[T] {
	s.where = ps
	return s
}

// func (s *Selector[T]) GetV1(ctx context.Context) (*T, error) {
// 	q, err := s.Build()
// 	if err != nil {
// 		return nil, err
// 	}

// 	db := s.db.db
// 	// 在这里，就是要发起查询，并且处理结果集
// 	rows, err := db.QueryContext(ctx, q.SQL, q.Args...)
// 	// 这个是查询错误
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 你要确认有没有数据
// 	if !rows.Next() {
// 		// 要不要返回 error？
// 		// 返回 error，和 sql 包语义保持一致
// 		return nil, ErrNoRows
// 	}

// 	cs, err := rows.Columns()
// 	if err != nil {
// 		return nil, err
// 	}

// 	var vals []any
// 	tp := new(T)
// 	// 起始地址
// 	address := reflect.ValueOf(tp).UnsafePointer()
// 	for _, c := range cs {
// 		fd, ok := s.model.ColumnMap[c]
// 		if !ok {
// 			return nil, errs.NewErrUnknownColumn(c)
// 		}
// 		// 是不是要计算字段的地址？
// 		// 起始地址 + 偏移量
// 		fdAddress := unsafe.Pointer(uintptr(address) + fd.Offset)
// 		// 反射在特定的地址上，创建一个特定类型的实例
// 		// 这里创建的实例是原本类型的指针类型
// 		// 例如 fd.Type = int，那么val 是 *int
// 		val := reflect.NewAt(fd.Typ, fdAddress)
// 		vals = append(vals, val.Interface())
// 	}

// 	err = rows.Scan(vals...)
// 	return tp, err
// }

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	q, err := s.Build()
	if err != nil {
		return nil, err
	}

	db := s.db.db
	// 在这里，就是要发起查询，并且处理结果集
	rows, err := db.QueryContext(ctx, q.SQL, q.Args...)
	// 这个是查询错误
	if err != nil {
		return nil, err
	}

	// if flag {
	// 	val := valuer.NewReflectValue()
	// } else {
	// 	val := valuer.NewUnsafeValue()
	// }
	//

	// 你要确认有没有数据
	if !rows.Next() {
		// 要不要返回 error？
		// 返回 error，和 sql 包语义保持一致
		return nil, ErrNoRows
	}

	tp := new(T)
	val := s.db.creator(s.model, tp)
	err = val.SetColumns(rows)

	return tp, err
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	panic("implement me")
}

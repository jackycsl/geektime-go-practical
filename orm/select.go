package orm

import (
	"context"
	"reflect"
	"strings"

	"github.com/jackycsl/geektime-go-practical/orm/internal/errs"
)

type Selector[T any] struct {
	table string
	model *Model
	where []Predicate
	sb    *strings.Builder
	args  []any

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
	sb.WriteString("SELECT * FROM ")
	// 我怎么把表名拿到
	if s.table == "" {
		sb.WriteByte('`')
		sb.WriteString(s.model.tableName)
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
		fd, ok := s.model.FieldMap[exp.name]
		// 字段不对，或者说列不对
		if !ok {
			return errs.NewErrUnknownField(exp.name)
		}
		s.sb.WriteByte('`')
		s.sb.WriteString(fd.colName)
		s.sb.WriteByte('`')
	case value:
		s.sb.WriteByte('?')
		s.addArg(exp.val)
	default:
		return errs.NewErrUnsupportedExpression(expr)
	}
	return nil
}

func (s *Selector[T]) addArg(val any) {
	if s.args == nil {
		s.args = make([]any, 0, 8)
	}
	s.args = append(s.args, val)
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

	// 你要确认有没有数据
	if !rows.Next() {
		// 要不要返回 error？
		// 返回 error，和 sql 包语义保持一致
		return nil, ErrNoRows
	}

	cs, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	tp := new(T)

	vals := make([]any, 0, len(cs))
	valElem := make([]reflect.Value, 0, len(cs))
	for _, c := range cs {
		fd, ok := s.model.ColumnMap[c]
		if !ok {
			return nil, errs.NewErrUnknownColumn(c)
		}
		// fd.Type = int, val 是 *int
		val := reflect.New(fd.typ)
		vals = append(vals, val.Interface())
		valElem = append(valElem, val.Elem())
	}

	err = rows.Scan(vals...)
	if err != nil {
		return nil, err
	}

	tpValueElem := reflect.ValueOf(tp).Elem()
	for i, c := range cs {
		fd, ok := s.model.ColumnMap[c]
		if !ok {
			return nil, errs.NewErrUnknownColumn(c)
		}
		if fd.colName == c {
			tpValueElem.FieldByName(fd.goName).Set(valElem[i])
		}
	}

	// 接口定义好之后，就两件事，一个是用新接口的方法改造上层，
	// 一个就是提供不同的实现
	return tp, nil
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	panic("implement me")
}

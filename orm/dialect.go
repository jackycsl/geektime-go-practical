package orm

import (
	"github.com/jackycsl/geektime-go-practical/orm/internal/errs"
)

var (
	DialectMySQL      Dialect = mysqlDialect{}
	DialectSQLite     Dialect = sqliteDialect{}
	DialectPostgreSQL Dialect = postgreDialect{}
)

type Dialect interface {
	// quoter 就是为了解决引号问题
	// MySQL `
	quoter() byte

	buildOnDuplicateKey(b *builder, odk *OnDuplicateKey) error
}

type standardSQL struct {
}

func (s standardSQL) quoter() byte {
	panic("not implemented") // TODO: Implement
}

func (standardSQL) buildOnDuplicateKey(b *builder, odk *OnDuplicateKey) error {
	panic("not implemented") // TODO: Implement
}

type mysqlDialect struct {
	standardSQL
}

func (s mysqlDialect) quoter() byte {
	return '`'
}

func (s mysqlDialect) buildOnDuplicateKey(b *builder, odk *OnDuplicateKey) error {
	b.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
	for idx, assign := range odk.assigns {
		if idx > 0 {
			b.sb.WriteByte(',')
		}
		switch a := assign.(type) {
		case Assignment:
			fd, ok := b.model.FieldMap[a.col]
			// 字段不对，或者说列不对
			if !ok {
				return errs.NewErrUnknownField(a.col)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=?")
			b.addArg(a.val)
		case Column:
			fd, ok := b.model.FieldMap[a.name]
			// 字段不对，或者说列不对
			if !ok {
				return errs.NewErrUnknownField(a.name)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=VALUES(")
			b.quote(fd.ColName)
			b.sb.WriteByte(')')
		default:
			return errs.NewErrUnsupportedAssignable(assign)
		}
	}

	return nil
}

type sqliteDialect struct {
	standardSQL
}

type postgreDialect struct {
	standardSQL
}

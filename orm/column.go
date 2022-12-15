package orm

type Column struct {
	name  string
	alias string
}

func C(name string) Column {
	return Column{name: name}
}

// func (c Column) assign() {}

func (c Column) As(alias string) Column {
	return Column{
		name:  c.name,
		alias: alias,
	}
}

// Eq 代表相等
// C("id").Eq(12)
// sub.C("id").Eq(12)
func (c Column) Eq(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opEq,
		right: valueOf(arg),
	}
}

func valueOf(arg any) Expression {
	switch val := arg.(type) {
	case Expression:
		return val
	default:
		return value{val: val}
	}
}

func (c Column) expr()       {}
func (c Column) selectable() {}

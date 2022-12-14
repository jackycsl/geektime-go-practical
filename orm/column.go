package orm

type Column struct {
	name string
}

func C(name string) Column {
	return Column{name: name}
}

// Eq 代表相等
// C("id").Eq(12)
// sub.C("id").Eq(12)
func (c Column) Eq(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opEq,
		right: value{val: arg},
	}
}

func (c Column) expr()       {}
func (c Column) selectable() {}

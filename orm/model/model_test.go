package model

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/jackycsl/geektime-go-practical/orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_Register(t *testing.T) {
	testCases := []struct {
		name      string
		entity    any
		wantModel *Model
		fields    []*Field
		wantErr   error
	}{
		{
			name:    "struct",
			entity:  TestModel{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:   "pointer",
			entity: &TestModel{},
			wantModel: &Model{
				TableName: "test_model",
			},
			fields: []*Field{
				{
					ColName: "id",
					GoName:  "Id",
					Typ:     reflect.TypeOf(int64(0)),
				},
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Typ:     reflect.TypeOf(""),
					Offset:  8,
				},
				{
					ColName: "age",
					GoName:  "Age",
					Typ:     reflect.TypeOf(int8(0)),
					Offset:  24,
				},
				{
					ColName: "last_name",
					GoName:  "LastName",
					Typ:     reflect.TypeOf(&sql.NullString{}),
					Offset:  32,
				},
			},
		},
		{
			name:    "map",
			entity:  map[string]string{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "slice",
			entity:  []int{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "basic Types",
			entity:  0,
			wantErr: errs.ErrPointerOnly,
		},
	}

	r := &registry{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Register(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			fieldMap := make(map[string]*Field)
			columnMap := make(map[string]*Field)
			for _, f := range tc.fields {
				fieldMap[f.GoName] = f
				columnMap[f.ColName] = f
			}
			tc.wantModel.FieldMap = fieldMap
			tc.wantModel.ColumnMap = columnMap
			assert.Equal(t, tc.wantModel, m)
		})
	}
}

func TestRegistry_get(t *testing.T) {
	testCases := []struct {
		name string

		entity    any
		wantModel *Model
		fields    []*Field
		wantErr   error
		cacheSize int
	}{
		{
			name:   "pointer",
			entity: &TestModel{},
			wantModel: &Model{
				TableName: "test_model",
			},
			fields: []*Field{
				{
					ColName: "id",
					GoName:  "Id",
					Typ:     reflect.TypeOf(int64(0)),
				},
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Typ:     reflect.TypeOf(""),
					Offset:  8,
				},
				{
					ColName: "age",
					GoName:  "Age",
					Typ:     reflect.TypeOf(int8(0)),
					Offset:  24,
				},
				{
					ColName: "last_name",
					GoName:  "LastName",
					Typ:     reflect.TypeOf(&sql.NullString{}),
					Offset:  32,
				},
			},
			cacheSize: 1,
		},
		{
			name: "tag",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column=first_name_t"`
				}
				return &TagTable{}
			}(),
			wantModel: &Model{
				TableName: "tag_table",
			},
			fields: []*Field{
				{
					ColName: "first_name_t",
					GoName:  "FirstName",
					Typ:     reflect.TypeOf(""),
				},
			},
		},
		{
			name: "empty column",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column="`
				}
				return &TagTable{}
			}(),
			wantModel: &Model{
				TableName: "tag_table",
			},
			fields: []*Field{
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Typ:     reflect.TypeOf(""),
				},
			},
		},
		{
			name: "column only",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column"`
				}
				return &TagTable{}
			}(),
			wantErr: errs.NewErrInvalidTagContent("column"),
		},
		{
			name: "ignore tag",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"abc=abc"`
				}
				return &TagTable{}
			}(),
			wantModel: &Model{
				TableName: "tag_table",
			},
			fields: []*Field{
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Typ:     reflect.TypeOf(""),
				},
			},
		},
		{
			name:   "table name",
			entity: &CustomTableName{},
			wantModel: &Model{
				TableName: "custom_table_name_t",
			},
			fields: []*Field{
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Typ:     reflect.TypeOf(""),
				},
			},
		},
		{
			name:   "table name ptr",
			entity: &CustomTableNamePtr{},
			wantModel: &Model{
				TableName: "custom_table_name_ptr_t",
			},
			fields: []*Field{
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Typ:     reflect.TypeOf(""),
				},
			},
		},
		{
			name:   "empty table name",
			entity: &EmptyTableName{},
			wantModel: &Model{
				TableName: "empty_table_name",
			},
			fields: []*Field{
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Typ:     reflect.TypeOf(""),
				},
			},
		},
	}

	r := NewRegistry()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Get(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			fieldMap := make(map[string]*Field)
			columnMap := make(map[string]*Field)
			for _, f := range tc.fields {
				fieldMap[f.GoName] = f
				columnMap[f.ColName] = f
			}
			tc.wantModel.FieldMap = fieldMap
			tc.wantModel.ColumnMap = columnMap

			assert.Equal(t, tc.wantModel, m)

			typ := reflect.TypeOf(tc.entity)
			cache, ok := r.(*registry).models.Load(typ)
			assert.True(t, ok)

			assert.Equal(t, tc.wantModel, cache)
		})
	}
}

type CustomTableName struct {
	FirstName string
}

func (c CustomTableName) TableName() string {
	return "custom_table_name_t"
}

type CustomTableNamePtr struct {
	FirstName string
}

func (c *CustomTableNamePtr) TableName() string {
	return "custom_table_name_ptr_t"
}

type EmptyTableName struct {
	FirstName string
}

func (c *EmptyTableName) TableName() string {
	return ""
}

func TestModelWithTableName(t *testing.T) {
	r := NewRegistry()
	m, err := r.Register(&TestModel{}, WithTableName("test_model_ttt"))
	require.NoError(t, err)
	assert.Equal(t, "test_model_ttt", m.TableName)
}

func TestModelWithColumnName(t *testing.T) {
	testCases := []struct {
		name    string
		field   string
		ColName string

		wantColName string
		wantErr     error
	}{
		{
			name:        "column name",
			field:       "FirstName",
			ColName:     "first_name_ccc",
			wantColName: "first_name_ccc",
		},
		{
			name:    "invalid column name",
			field:   "XXX",
			ColName: "first_name_ccc",
			wantErr: errs.NewErrUnknownField("XXX"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.ColName, func(t *testing.T) {
			r := NewRegistry()
			m, err := r.Register(&TestModel{}, WithColumneName(tc.field, tc.ColName))
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			fd, ok := m.FieldMap[tc.field]
			require.True(t, ok)
			assert.Equal(t, tc.wantColName, fd.ColName)
		})
	}
}

type TestModel struct {
	Id int64
	// ""
	FirstName string
	Age       int8
	LastName  *sql.NullString
}

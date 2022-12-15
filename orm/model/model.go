package model

import (
	"reflect"
	"strings"
	"sync"
	"unicode"

	"github.com/jackycsl/geektime-go-practical/orm/internal/errs"
)

const (
	tagKeyColumn = "column"
)

type Registry interface {
	Get(val any) (*Model, error)
	Register(val any, opts ...Option) (*Model, error)
}

type Model struct {
	TableName string
	Fields    []*Field
	// 上面是字段名到字段定义的映射
	FieldMap map[string]*Field
	// 列名到字段定义的映射
	ColumnMap map[string]*Field
}

type Option func(*Model) error

type Field struct {
	GoName string
	// 列名
	ColName string
	// 代表的是字段的类型
	Type reflect.Type

	// 字段相对于结构体本身的偏移量
	Offset uintptr
}

// var models = map[reflect.Type]*Model{}

// 全局默认的
// var defaultRegistry = &registry{
// 	models: map[reflect.Type]*Model{},
// }

// registry 代表的是元数据的注册中心
type registry struct {
	// 读写锁
	// lock   sync.RWMutex
	// models map[reflect.Type]*Model
	models sync.Map
}

func NewRegistry() Registry {
	return &registry{}
}

func (r *registry) Get(val any) (*Model, error) {
	typ := reflect.TypeOf((val))
	m, ok := r.models.Load(typ)
	if ok {
		return m.(*Model), nil
	}
	m, err := r.Register(val)
	if err != nil {
		return nil, err
	}

	return m.(*Model), nil
}

// func (r *registry) get1(val any) (*Model, error) {
// 	typ := reflect.TypeOf(val)

// 	r.lock.RLock()
// 	m, ok := r.Models[typ]
// 	r.lock.RUnlock()
// 	if ok {
// 		return m, nil
// 	}

// 	r.lock.Lock()
// 	defer r.lock.Unlock()
// 	m, ok = r.Models[typ]
// 	if ok {
// 		return m, nil
// 	}

// 	m, err := r.parseModel(val)
// 	if err != nil {
// 		return nil, err
// 	}
// 	r.Models[typ] = m
// 	return m, nil
// }

// Register 限制只能用一级指针
func (r *registry) Register(entity any, opts ...Option) (*Model, error) {
	typ := reflect.TypeOf(entity)
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}
	elemTyp := typ.Elem()
	// for elemTyp.Kind() == reflect.Pointer {
	// 	elemTyp = elemTyp.Elem()
	// }
	numField := elemTyp.NumField()
	fieldMap := make(map[string]*Field, numField)
	columnMap := make(map[string]*Field, numField)
	fields := make([]*Field, 0, numField)
	for i := 0; i < numField; i++ {
		fd := elemTyp.Field(i)
		pair, err := r.parseTag(fd.Tag)
		if err != nil {
			return nil, err
		}
		colName := pair[tagKeyColumn]
		if colName == "" {
			colName = underscoreName(fd.Name)
		}
		fdMeta := &Field{
			GoName:  fd.Name,
			ColName: colName,
			Type:    fd.Type,
			Offset:  fd.Offset,
		}
		fieldMap[fd.Name] = fdMeta
		columnMap[colName] = fdMeta
		fields = append(fields, fdMeta)
	}

	var tableName string
	if tbl, ok := entity.(TableName); ok {
		tableName = tbl.TableName()
	}
	if tableName == "" {
		tableName = underscoreName(elemTyp.Name())
	}

	res := &Model{
		TableName: tableName,
		FieldMap:  fieldMap,
		ColumnMap: columnMap,
		Fields:    fields,
	}

	for _, opt := range opts {
		err := opt(res)
		if err != nil {
			return nil, err
		}
	}

	r.models.Store(typ, res)
	return res, nil
}

func WithTableName(tableName string) Option {
	return func(m *Model) error {
		m.TableName = tableName
		return nil
	}
}

func WithColumneName(field string, colName string) Option {
	return func(m *Model) error {
		fd, ok := m.FieldMap[field]
		if !ok {
			return errs.NewErrUnknownField(field)
		}
		fd.ColName = colName
		return nil
	}
}

// type User struct {
// 	ID uint64 `orm:"column=id,xxx=bbb`
// }

func (r *registry) parseTag(tag reflect.StructTag) (map[string]string, error) {
	ormTag, ok := tag.Lookup("orm")
	if !ok {
		return map[string]string{}, nil
	}
	pairs := strings.Split(ormTag, ",")
	res := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		segs := strings.Split(pair, "=")
		if len(segs) != 2 {
			return nil, errs.NewErrInvalidTagContent(pair)
		}
		key := segs[0]
		val := segs[1]
		res[key] = val
	}
	return res, nil
}

// underscoreName 驼峰转字符串命名
func underscoreName(tableName string) string {
	var buf []byte
	for i, v := range tableName {
		if unicode.IsUpper(v) {
			if i != 0 {
				buf = append(buf, '_')
			}
			buf = append(buf, byte(unicode.ToLower(v)))
		} else {
			buf = append(buf, byte(v))
		}

	}
	return string(buf)
}

type TableName interface {
	TableName() string
}

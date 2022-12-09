package orm

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
	Register(val any, opts ...ModelOpt) (*Model, error)
}

type Model struct {
	tableName string
	// fields    map[string]*Field
	// 上面是字段名到字段定义的映射
	FieldMap map[string]*Field
	// 列名到字段定义的映射
	ColumnMap map[string]*Field
}

type ModelOpt func(*Model) error

type Field struct {
	goName string
	// 列名
	colName string
	// 代表的是字段的类型
	typ reflect.Type
}

// var Models = map[reflect.Type]*Model{}

// 全局默认的
// var defaultRegistry = &registry{
// 	Models: map[reflect.Type]*Model{},
// }

// registry 代表的是元数据的注册中心
type registry struct {
	// 读写锁
	// lock   sync.RWMutex
	// Models map[reflect.Type]*Model
	Models sync.Map
}

func NewRegistry() *registry {
	return &registry{}
}

func (r *registry) Get(val any) (*Model, error) {
	typ := reflect.TypeOf((val))
	m, ok := r.Models.Load(typ)
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
func (r *registry) Register(entity any, opts ...ModelOpt) (*Model, error) {
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
			goName:  fd.Name,
			colName: colName,
			typ:     fd.Type,
		}
		fieldMap[fd.Name] = fdMeta
		columnMap[colName] = fdMeta
	}

	var tableName string
	if tbl, ok := entity.(TableName); ok {
		tableName = tbl.TableName()
	}
	if tableName == "" {
		tableName = underscoreName(elemTyp.Name())
	}

	res := &Model{
		tableName: tableName,
		FieldMap:  fieldMap,
		ColumnMap: columnMap,
	}

	for _, opt := range opts {
		err := opt(res)
		if err != nil {
			return nil, err
		}
	}

	r.Models.Store(typ, res)
	return res, nil
}

func ModelWithTableName(tableName string) ModelOpt {
	return func(m *Model) error {
		m.tableName = tableName
		return nil
	}
}

func ModelWithColumneName(field string, colName string) ModelOpt {
	return func(m *Model) error {
		fd, ok := m.FieldMap[field]
		if !ok {
			return errs.NewErrUnknownField(field)
		}
		fd.colName = colName
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

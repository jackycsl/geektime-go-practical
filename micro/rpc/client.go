package rpc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

// InitService 要为 GetById 之类的函数类型的字段赋值
func InitClientProxy(service Service) error {
	// 在这里初始化一个 Proxy
	return setFuncField(service, nil)
}

func setFuncField(service Service, p Proxy) error {
	if service == nil {
		return errors.New("rpc: 不支持 nil")
	}
	val := reflect.ValueOf(service)
	typ := val.Type()
	// 只支持指向结构体的一级指针
	if typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
		return errors.New("rpc: 只支持指向结构体的一级指针")
	}

	val = val.Elem()
	typ = typ.Elem()

	numfield := typ.NumField()
	for i := 0; i < numfield; i++ {
		fieldTyp := typ.Field(i)
		fieldVal := val.Field(i)

		if fieldVal.CanSet() {
			// 这个地方才是真正的将本地调用捕捉到的地方
			fn := func(args []reflect.Value) (results []reflect.Value) {
				// args[0] 是 context
				ctx := args[0].Interface().(context.Context)
				// args[1] 是 req
				req := &Request{
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,
					Arg:         args[1].Interface(),
				}

				retVal := reflect.New(fieldTyp.Type.Out(0)).Elem()
				// 要真的发起调用了
				resp, err := p.Invoke(ctx, req)
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				fmt.Println(resp)
				return []reflect.Value{retVal, reflect.Zero(reflect.TypeOf(new(error)).Elem())}
			}
			// 我要设置值给 GetById
			fnVal := reflect.MakeFunc(fieldTyp.Type, fn)
			fieldVal.Set(fnVal)
		}
	}

	return nil
}

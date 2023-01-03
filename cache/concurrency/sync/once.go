package sync

import "sync"

type MyBiz struct {
	once sync.Once
}

func (m *MyBiz) Init() {
	m.once.Do(func() {

	})
}

// func (m MyBiz) InitV1() {
// 	m.once.Do(func() {

// 	})
// }

type MyBizV1 struct {
	once *sync.Once
}

func (m MyBizV1) Init() {
	m.once.Do(func() {

	})
}

type MyBusiness interface {
	DoSomething()
}

type singleton struct {
}

func (s *singleton) DoSomething() {
	panic("not implemented") // TODO: Implement
}

var s *singleton
var singletonOnce sync.Once

func GetSingleton() *singleton {
	singletonOnce.Do(func() {
		s = &singleton{}
	})
	return s
}

package test

type Logger interface {
	Output(maxdepth int, s string) error
}

type tbLog interface {
	Log(...interface{})
}

// 接口继承， 接口继承的目的是，为了拓展原有接口定义的功能，实际例子如sort包 Reverse方法及定义
/*
	1. 接口继承时，不必重新全部定义继承接口的方法，只需要更改部分目标方法即可
	2. 接口继承只能更改传入的参数的逻辑（或者说拓展原有接口功能就是通过在传入参数部分做新的逻辑处理达成的）
*/
type testLogger struct {
	tbLog
}

// 接口继承的结构体中,可以引用原本接口定义好的方法，去组合拓展原有接口方法实现
func (tl *testLogger) Output(maxdepth int, s string) error {
	tl.Log(s)
	return nil
}

// 通过构造的方式，将原本已经实现了接口的对象传入，然后将拓展过功能的对象返回（类似于接口对象的装饰器模式）
func NewTestLogger(tbl tbLog) Logger {
	return &testLogger{tbl}
}

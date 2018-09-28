package util

import (
	"sync"
)

type WaitGroupWrapper struct {
	sync.WaitGroup
}

// 为了保证协程会执行， 避免主协程退出导致子协程不执行
func (w *WaitGroupWrapper) Wrap(cb func()) {
	w.Add(1)
	go func() {
		cb()
		w.Done()
	}()
}

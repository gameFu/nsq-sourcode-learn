package dirlock

import (
	"fmt"
	"os"
	"syscall"
)

// 用来锁定目录（datapath）
type DirLock struct {
	// 要锁定的路径
	dir string
	// 文件对象，os.Open可以打开dir，并吧dir当成文件对象返回
	f *os.File
}

func New(dir string) *DirLock {
	return &DirLock{
		dir: dir,
	}
}

func (lock *DirLock) Lock() error {
	f, err := os.Open(lock.dir)
	if err != nil {
		return err
	}
	lock.f = f
	/*
		syscall.Flock是文件建议锁.当使用这个锁时，文件会被上建议锁，建议锁并不会真正阻止其他进程修改文件内容，而仅仅是它用来告诉检查该文件是否被加锁的进程已经枷锁了
		syscall.Flock这个方法首先会检查这个文件是否已经枷锁，如果已经加锁就会进行，等待或者直接报错结束，syscall.LOCK_NB是直接结束的意思
		LOCK.EX是排他锁，常呗用做写锁
	*/
	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		return fmt.Errorf("cannot flock directory %s - %s", lock.dir, err)
	}
	return nil
}

func (lock *DirLock) Unlock() error {
	defer lock.f.Close()
	return syscall.Flock(int(lock.f.Fd()), syscall.LOCK_UN)
}

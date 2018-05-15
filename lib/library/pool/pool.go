package pool

import (
	"sync"
)

type Pool struct {
	pool  chan bool      //控制数量
	group sync.WaitGroup //用于等待所有goroutine执行完毕
}

//添加等待goroutine的数量
func (pool *Pool) Add() {
	pool.pool <- true
	pool.group.Add(1)
}

//减少等待groutine的数量
func (pool *Pool) Done() {
	<-pool.pool
	pool.group.Done()
}

//等待groutine的数量为0
func (pool *Pool) Wait() {
	pool.group.Wait()
}

//初始化等待goroutine的最大数量
func (pool *Pool) Init(maxNum int) {
	pool.pool = make(chan bool, maxNum)
}

//新建并发控制器
func NewPool(maxNum int) *Pool {
	return &Pool{
		pool: make(chan bool, maxNum),
	}
}


//goroutine池中已分配的数量
func (pool *Pool) Num() int {
	return len(pool.pool)
}

//goroutine池容量
func (pool *Pool) Cap() int {
	return cap(pool.pool)
}


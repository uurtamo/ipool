// Copyright 2020 Steve Uurtamo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ipool

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

type Pool struct {
	Value sync.Map
	size  uint64
	queue *twoLockQueue
	pmux  sync.Mutex
}

type Interval struct {
	Begin     uint64
	End       uint64
	allocated bool
	Handle    uint64
	Hook      uint64
}

type node struct {
	value *Interval
	next  *node
}

type twoLockQueue struct {
	hmux sync.Mutex
	tmux sync.Mutex
	head *node
	tail *node
}

func (Q *twoLockQueue) Init() *twoLockQueue {

	i_node := new(node)
	i_node.next = nil
	Q.head = i_node
	Q.tail = i_node
	return Q
}

func (Q *twoLockQueue) enqueue(item *Interval) {

	t_node := new(node)
	t_node.value = item
	t_node.next = nil

	Q.tmux.Lock()

	Q.tail.next = t_node
	Q.tail = t_node

	Q.tmux.Unlock()
	return
}

func (Q *twoLockQueue) dequeue() *Interval {

	// lock the queue

	Q.hmux.Lock()
	defer Q.hmux.Unlock()

	h_new := Q.head.next

	if h_new == nil {
		return nil
	}

	value := h_new.value

	Q.head = h_new

	return value
}

func (iPool *Pool) Init() *Pool {

	var Queue = new(twoLockQueue)
	var Locations sync.Map
	var tmpInterval *Interval
	var newSize uint64

	Queue.Init()
	iPool.queue = Queue

	iPool.pmux.Lock()
	newSize = atomic.AddUint64(&iPool.size, 1)

	tmpInterval = new(Interval)
	tmpInterval.allocated = false
	tmpInterval.Handle = 8 * newSize
	Locations.Store(tmpInterval.Handle, tmpInterval)

	iPool.queue.enqueue(tmpInterval)
	iPool.Value = Locations
	iPool.pmux.Unlock()

	return iPool
}

func (iPool *Pool) Alloc() (*Interval, uint64) {

	var Item *Interval
	var IntervalTmp *Interval
	var IntervalRet *Interval
	var handleTmp uint64
	var counter uint64
	var startSize uint64
	var newSize uint64

	Item = iPool.queue.dequeue()

	if Item == nil {
		// the queue is out of items to give

		iPool.pmux.Lock()
		newSize = atomic.AddUint64(&iPool.size, 1)
		IntervalRet = new(Interval)
		handleTmp = 8 * newSize
		iPool.Value.Store(handleTmp, IntervalRet)
		iPool.pmux.Unlock()

		IntervalRet.Handle = handleTmp

		startSize = atomic.LoadUint64(&iPool.size)

		iPool.pmux.Lock()

		for counter = 1; counter <= startSize; counter++ {

			IntervalTmp = new(Interval)
			IntervalTmp.allocated = false
			newSize = atomic.AddUint64(&iPool.size, 1)
			handleTmp = 8 * newSize
			iPool.Value.Store(handleTmp, IntervalTmp)
			IntervalTmp.Handle = handleTmp
			iPool.queue.enqueue(IntervalTmp)
		}

		iPool.pmux.Unlock()

		if IntervalRet.allocated == true {
			fmt.Printf("this should never happen. reallocateed object %v\n", IntervalRet)
			os.Exit(13)
		}
		IntervalRet.allocated = true
		return IntervalRet, IntervalRet.Handle
	} else {
		IntervalRet = Item
		if IntervalRet.allocated == true {
			fmt.Printf("this should never happen. reallocateed object %v\n", IntervalRet)
			os.Exit(13)
		}
		IntervalRet.allocated = true
		return IntervalRet, IntervalRet.Handle
	}
}

func (iPool *Pool) Free(PassedInterval *Interval) {

	var handle uint64

	// clear the bottom 3 bits

	handle = PassedInterval.Handle
	handle &^= 7

	PassedInterval.allocated = false
	iPool.queue.enqueue(PassedInterval)
}

func (iPool *Pool) FreeHandle(handle uint64) {

	var IntervalTmp *Interval

	// clear the bottom 3 bits

	handle &^= 7

	tmpVal, _ := iPool.Value.Load(handle)
	IntervalTmp = tmpVal.(*Interval)

	if IntervalTmp == nil {
		fmt.Printf("this should never happen. Freed interval %v is not in the pool location map for pool %v\n", handle, iPool)
		os.Exit(13)
	}
	IntervalTmp.allocated = false
	iPool.queue.enqueue(IntervalTmp)
}

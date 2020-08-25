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

type pool struct {
	locations sync.Map
	//	locations map[uint64]*interval
	size  uint64
	queue *twoLockQueue
	pmux  sync.Mutex
}

type interval struct {
	begin     uint64
	end       uint64
	allocated bool
	handle    uint64
}

type node struct {
	value *interval
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

func (Q *twoLockQueue) enqueue(item *interval) {

	t_node := new(node)
	t_node.value = item
	t_node.next = nil

	Q.tmux.Lock()

	Q.tail.next = t_node
	Q.tail = t_node

	Q.tmux.Unlock()
	return
}

func (Q *twoLockQueue) dequeue() *interval {

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

func (Pool *pool) Init() *pool {

	var Queue = new(twoLockQueue)
	var Locations sync.Map
	var Interval *interval
	var newSize uint64

	Queue.Init()
	Pool.queue = Queue

	Pool.pmux.Lock()
	newSize = atomic.AddUint64(&Pool.size, 1)

	Interval = new(interval)
	Interval.allocated = false
	Interval.handle = 8 * newSize
	Locations.Store(Interval.handle, Interval)

	Pool.queue.enqueue(Interval)
	Pool.locations = Locations
	Pool.pmux.Unlock()

	return Pool
}

func (Pool *pool) Alloc() *interval {

	var Item *interval
	var IntervalTmp *interval
	var IntervalRet *interval
	var handleTmp uint64
	var counter uint64
	var startSize uint64
	var newSize uint64

	Item = Pool.queue.dequeue()

	if Item == nil {
		// the queue is out of items to give

		Pool.pmux.Lock()
		newSize = atomic.AddUint64(&Pool.size, 1)
		IntervalRet = new(interval)
		handleTmp = 8 * newSize
		Pool.locations.Store(handleTmp, IntervalRet)
		Pool.pmux.Unlock()

		IntervalRet.handle = handleTmp

		startSize = atomic.LoadUint64(&Pool.size)

		Pool.pmux.Lock()

		for counter = 1; counter <= startSize; counter++ {

			IntervalTmp = new(interval)
			IntervalTmp.allocated = false
			newSize = atomic.AddUint64(&Pool.size, 1)
			handleTmp = 8 * newSize
			Pool.locations.Store(handleTmp, IntervalTmp)
			IntervalTmp.handle = handleTmp
			Pool.queue.enqueue(IntervalTmp)
		}

		Pool.pmux.Unlock()

		if IntervalRet.allocated == true {
			fmt.Printf("this should never happen. reallocateed object %v\n", IntervalRet)
			os.Exit(13)
		}
		IntervalRet.allocated = true
		return IntervalRet
	} else {
		IntervalRet = Item
		if IntervalRet.allocated == true {
			fmt.Printf("this should never happen. reallocateed object %v\n", IntervalRet)
			os.Exit(13)
		}
		IntervalRet.allocated = true
		return IntervalRet
	}
}

func (Pool *pool) Free(Interval *interval) {

	var handle uint64

	// clear the bottom 3 bits

	handle = Interval.handle
	handle &^= 7

	Interval.allocated = false
	Pool.queue.enqueue(Interval)
}

func (Pool *pool) FreeHandle(handle uint64) {

	var IntervalTmp *interval

	// clear the bottom 3 bits

	handle &^= 7

	//	IntervalTmp = Pool.locations[handle]

	tmpVal, _ := Pool.locations.Load(handle)
	IntervalTmp = tmpVal.(*interval)

	if IntervalTmp == nil {
		fmt.Printf("this should never happen. Freed interval %v is not in the pool location map for pool %v\n", handle, Pool)
		os.Exit(13)
	}
	IntervalTmp.allocated = false
	Pool.queue.enqueue(IntervalTmp)
}

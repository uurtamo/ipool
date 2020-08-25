# ipool

This is a memory pool for intervals


It simply inserts freed intervals into a queue, and allocates intervals from that same queue.


If the queue is empty (for instance when you initialize one, or if all allocated intervals are in use), it will make some unallocated intervals and insert them for later use.


The idea is that once the queue size stabilizes to be bigger the size that represents
your max flow rate in and out of the queue, it should be the speed of the enqueue/dequeue
operation that limits how quickly you can get one rather than the speed of allocating it using the system memory manager.


It uses a 2-lock queue behind the scenes, so is efficient when allocations and frees are roughly in balance.


The queue is based upon the michael+scott 1996 PODC locking queue. There is a map (using golang's sync.Map) that maintains a mapping between handles and intervals. Handles act as virtual pointers to the intervals, are unique uint64's, and are arranged so as to have their bottom three bits cleared upon allocation and after freeing them.


On a 2.4GHz 4 core i7:

goos: darwin

goarch: amd64

pkg: math/ipool

(with no contention)		BenchmarkAlloc-4           	13164860	 89.4 ns/op

(4x contention)      		BenchmarkAllocParallel-4   	 6800242	167 ns/op

(no contention, not parallel) 	BenchmarkNew-4             	 4072016	328 ns/op



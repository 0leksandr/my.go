package my

import "io"

type Queue[T any] interface {
	io.Closer
	Put(T)
	Get() T
}

type ChannelQueue[T any] struct {
	channel chan T
}
func (ChannelQueue[T]) New(size int) ChannelQueue[T] {
	return ChannelQueue[T]{make(chan T, size)}
}
func (queue ChannelQueue[T]) Close() error {
	close(queue.channel)
	return nil
}
func (queue ChannelQueue[T]) Put(value T) {
	queue.channel <- value
}
func (queue ChannelQueue[T]) Get() T {
	return <-queue.channel
}
func (queue ChannelQueue[T]) Channel() chan T {
	return queue.channel
}

//// Sync1Queue SingularQueue useless, because holds only 1 value
//type Sync1Queue[T any] struct {
//	value    T
//	received Locker
//	sent     Locker
//	putFirst bool
//}
//func (*Sync1Queue[T]) New(putFirst bool) *Sync1Queue[T] {
//	queue := &Sync1Queue[T]{
//		putFirst: putFirst,
//	}
//	queue.received.Lock()
//	return queue
//}
//func (queue *Sync1Queue[T]) Close() error {
//	return nil
//}
//func (queue *Sync1Queue[T]) Put(value T) {
//	if queue.putFirst {
//		queue.sent.Wait()
//		queue.sent.Lock()
//		queue.value = value
//		queue.received.Unlock()
//	} else {
//		queue.sent.Lock()
//		queue.value = value
//		queue.received.Unlock()
//		queue.sent.Wait()
//	}
//}
//func (queue *Sync1Queue[T]) Get() T {
//	queue.received.Wait()
//	queue.received.Lock()
//	defer queue.sent.Unlock()
//	return queue.value
//}

type zeroQueueReceiver[T any] interface {
	Put(T)
	Get() T
}

type zeroQueueSingularReceiver[T any] struct {
	zeroQueueReceiver[T]
	value  T
	locker Locker
}
func (*zeroQueueSingularReceiver[T]) New() *zeroQueueSingularReceiver[T] {
	receiver := zeroQueueSingularReceiver[T]{}
	receiver.locker.Lock()
	return &receiver
}
func (receiver *zeroQueueSingularReceiver[T]) Put(value T) {
	receiver.value = value
	receiver.locker.Unlock()
}
func (receiver *zeroQueueSingularReceiver[T]) Get() T {
	receiver.locker.Wait()
	return receiver.value
}

type ReservoirQueue[T any] struct {
	in  chan<- T
	out <-chan T
}
func (ReservoirQueue[T]) New() ReservoirQueue[T] {
	in := make(chan T)

	return ReservoirQueue[T]{
		in:  in,
		out: Reservoir(in, 0),
	}
}
func (queue ReservoirQueue[T]) Close() error {
	close(queue.in)
	return nil
}
func (queue ReservoirQueue[T]) Put(value T) {
	queue.in <- value
}
func (queue ReservoirQueue[T]) Get() T {
	return <-queue.out
}

//// FairChannelQueue useless, because regular ChannelQueue seems to be already fair
//type FairChannelQueue[T any] struct {
//	receivers ReservoirQueue[*zeroQueueSingularReceiver[T]]
//	values    ReservoirQueue[T]
//	mutex     sync.Mutex
//	balance   int
//}
//func (*FairChannelQueue[T]) New() *FairChannelQueue[T] {
//	return &FairChannelQueue[T]{
//		receivers: ReservoirQueue[*zeroQueueSingularReceiver[T]]{}.New(),
//		values:    ReservoirQueue[T]{}.New(),
//	}
//}
//func (queue *FairChannelQueue[T]) Close() error {
//	return ComboError(
//		queue.receivers.Close(),
//		queue.values.Close(),
//	)
//}
//func (queue *FairChannelQueue[T]) Put(value T) {
//	queue.mutex.Lock()
//	queue.balance++
//
//	if queue.balance <= 0 {
//		queue.receivers.Get().Put(value)
//	} else {
//		queue.values.Put(value)
//	}
//	queue.mutex.Unlock()
//}
//func (queue *FairChannelQueue[T]) Get() T {
//	queue.mutex.Lock()
//	queue.balance--
//
//	if queue.balance >= 0 {
//		value := queue.values.Get()
//		queue.mutex.Unlock()
//		return value
//	} else {
//		receiver := (*zeroQueueSingularReceiver[T])(nil).New()
//		queue.receivers.Put(receiver)
//		queue.mutex.Unlock()
//		return receiver.Get()
//	}
//}

type zeroQueueZeroReceiver[T any] struct {
	zeroQueueReceiver[T]
	value          T
	received, sent Locker
}
func (*zeroQueueZeroReceiver[T]) New() *zeroQueueZeroReceiver[T] {
	receiver := &zeroQueueZeroReceiver[T]{}
	receiver.received.Lock()
	return receiver
}
func (receiver *zeroQueueZeroReceiver[T]) Put(value T) {
	receiver.value = value
	receiver.sent.Lock()
	receiver.received.Unlock()
	receiver.sent.Wait()
}
func (receiver *zeroQueueZeroReceiver[T]) Get() T {
	receiver.received.Wait()
	defer receiver.sent.Unlock()
	return receiver.value
}

type ZeroQueue[T any] struct {
	receivers ReservoirQueue[zeroQueueReceiver[T]]
}
func (ZeroQueue[T]) New() ZeroQueue[T] {
	return ZeroQueue[T]{
		receivers: ReservoirQueue[zeroQueueReceiver[T]]{}.New(),
	}
}
func (queue ZeroQueue[T]) Close() error {
	return queue.receivers.Close()
}
func (queue ZeroQueue[T]) Put(value T) {
	receiver := queue.receivers.Get()
	receiver.Put(value)
}
func (queue ZeroQueue[T]) Get() T {
	//receiver := (*zeroQueueSingularReceiver[T])(nil).New()
	//receiver := (*Sync1Queue[T])(nil).New(false)
	receiver := (*zeroQueueZeroReceiver[T])(nil).New()
	queue.receivers.Put(receiver)
	return receiver.Get()
}

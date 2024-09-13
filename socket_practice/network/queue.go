package network

import (
	"container/list"
	"sync"
)

const (
	CAPACITY_INFINITE = -1
)

/* Session Index 관리 */
type Deque struct {
	sync.RWMutex
	container *list.List
	capacity  int
}

/*
item을 interface 형태로 넣는다.
컨테이너 뒤에 값을 넣는다.
*/

func (d *Deque) Append(item interface{}) (int, bool) {
	// interface{}는 모든 자료형이 될 수 있다.
	// 꺼낼 때만 형을 지정해주면 된다.
	d.Lock()
	defer d.Unlock()

	count := d.container.Len()
	if d.capacity < 0 || count < d.capacity {
		d.container.PushBack(item)
		count += 1
		return count, true
	}

	return 0, false
}

/*
item을 interface형태로 넣는다.
컨테이너 앞에 값을 넣는다.
*/
func (d *Deque) Prepend(item interface{}) (int, bool) {
	d.Lock()
	defer d.Unlock()

	count := d.container.Len()
	if d.capacity < 0 || count < d.capacity {
		d.container.PushFront(item)
		count += 1
		return count, true
	}

	return 0, false
}

/*
컨테이너 뒤에서 값을 뺀다.
*/
func (d *Deque) pop() interface{} {
	d.Lock()
	defer d.Unlock()

	var item interface{} = nil
	var lastContainerItem = d.container.Back()
	if lastContainerItem != nil {
		item = d.container.Remove(lastContainerItem)
	}

	return item
}

/*
컨테이너 앞에서 값을 뺀다.
*/
func (d *Deque) Shift() interface{} {
	d.Lock()
	defer d.Unlock()

	var item interface{} = nil
	var firstContainerItem *list.Element = nil

	firstContainerItem = d.container.Front()
	if firstContainerItem != nil {
		item = d.container.Remove(firstContainerItem)
	}

	return item
}

/*
컨테이너 앞의 값을 확인한다.
(지우진 않는다.)
*/
func (d *Deque) First() interface{} {
	d.RLock()
	defer d.RUnlock()

	item := d.container.Front()
	if item != nil {
		return item.Value
	}
	return nil
}

/*
컨테이너 뒤의 값을 확인한다.
(지우진 않는다.)
*/
func (d *Deque) Last() interface{} {
	d.Lock()
	defer d.RUnlock()

	item := d.container.Back()
	if item != nil {
		return item.Value
	}

	return nil
}

/*
컨테이너 크기를 확인한다.
*/
func (d *Deque) Size() int {
	d.RLock()
	defer d.RUnlock()

	return d.container.Len()
}

/*
컨테이너 용량을 확인한다.
*/
func (d *Deque) Capacity() int {
	d.RLock()
	defer d.RUnlock()

	return d.capacity
}

/*
컨테이너가 현재 비워져있는지 확인한다.
*/
func (d *Deque) Empty() bool {
	d.RLock()
	defer d.RUnlock()

	return d.container.Len() == 0
}

/*
컨테이너가 현재 가득 찬 상태인지 확인한다.
*/
func (d *Deque) Full() bool {
	d.RLock()
	defer d.RUnlock()

	return d.capacity >= 0 && d.container.Len() >= d.capacity
}

/*
새로운 컨테이너를 생성한다.
용량은 인자로 설정한다.
*/
func NewDeque() *Deque {
	return NewCappedDeque(CAPACITY_INFINITE)
}

func NewCappedDeque(capacity int) *Deque {
	return &Deque{
		container: list.New(),
		capacity:  capacity, // POOL_SIZE로 초기화됨.
	}
}

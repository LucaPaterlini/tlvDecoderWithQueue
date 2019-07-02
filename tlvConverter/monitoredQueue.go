package tlvconverter

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// MonitoredQueue contains the implementation of the struct a monitored queue.
type MonitoredQueue struct {
	in         int64
	out        int64
	base       chan PacketType
	lastCheck  time.Time
	mu         sync.RWMutex
	readDelay  time.Duration
	writeDelay time.Duration
}

// NewMonitoredQueue initialize and return a new monitored queue.
func NewMonitoredQueue(size int) MonitoredQueue {
	return MonitoredQueue{base: make(chan PacketType, size), lastCheck: time.Now()}
}

// Add put a new packet on the monitored queue and increase the size of in items.
func (mq *MonitoredQueue) Add(p PacketType) {
	atomic.AddInt64(&mq.in, 1)
	mq.base <- p
	return
}

// GetFirst get the first element from its channel of packets and inc the number of out elements.
func (mq *MonitoredQueue) GetFirst() (p PacketType, ok bool) {
	atomic.AddInt64(&mq.out, 1)
	p, ok = <-mq.base
	return
}

// Load returns the load of the monitored Queue.
func (mq *MonitoredQueue) Load() float64 {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	return float64(len(mq.base)) / float64(cap(mq.base))
}

// RwRate returns the rate of in packet and out packet from the monitoredQueue from the last measurement.
func (mq *MonitoredQueue) RwRate() (float64, float64) {
	now := time.Now()
	mq.mu.Lock()
	defer mq.mu.Unlock()
	timeSpan := now.Sub(mq.lastCheck)
	inC, outC := mq.in, mq.out
	mq.in, mq.out = 0, 0
	return (float64(inC * int64(time.Second))) / float64(timeSpan), (float64(outC * int64(time.Second))) / float64(timeSpan)
}

// Close close the monitored queue channel.
func (mq *MonitoredQueue) Close() {
	close(mq.base)
}

// MonitorLoad checks if the queue is full increase the delay if its empty decrease the delay.
func (mq *MonitoredQueue) MonitorLoad() {
	switch mq.Load() {
	case 0:
		mq.writeDelay -= time.Millisecond
		if mq.writeDelay < 0 {
			mq.writeDelay = 0
		}
		log.Println("The monitoredQueue is empty, decreasing the delay of 1 ms")
	case 1:
		mq.writeDelay += time.Millisecond
		log.Println("The monitoredQueue is full, increasing the delay of 1 ms")

	}
	inRate, outRate := mq.RwRate()
	log.Printf("item in %.5f/s , item out %.5f/s\n", inRate, outRate)
}

// SetReadDelay sets the delay of the reader, its usefull to make the testing more realistic
func (mq *MonitoredQueue) SetReadDelay(duration time.Duration) {
	mq.readDelay = duration
}

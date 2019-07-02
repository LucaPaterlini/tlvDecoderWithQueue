package tlvconverter

import (
	"reflect"
	"testing"
	"time"
)

func TestNewMonitoredQueue(t *testing.T) {
	nMq := NewMonitoredQueue(10)
	gotCap := cap(nMq.base)
	if gotCap != 10 {
		t.Errorf("initialization error expected cap=10 : cap=%d", gotCap)
	}
}

func TestMonitoredQueue_Add(t *testing.T) {
	nMq := NewMonitoredQueue(10)
	nPk := PacketType{A: 5, B: 1.1, C: 2.2, D: 3.3, E: 4.4, F: true}
	nMq.Add(nPk)
	if !(int64(len(nMq.base)) == nMq.in && nMq.in == 1) {
		t.Error("wrong accounting of add")
	}
}

func TestMonitoredQueue_GetFirst(t *testing.T) {
	nMq := NewMonitoredQueue(10)
	nPk := PacketType{A: 5, B: 1.1, C: 2.2, D: 3.3, E: 4.4, F: true}
	nMq.Add(nPk)
	got, ok := nMq.GetFirst()
	if !ok {
		t.Error("channel closed")
	}
	if !reflect.DeepEqual(got, nPk) {
		t.Errorf("got: %v, expected %v ", got, nPk)
	}
}

func TestMonitoredQueue_Load(t *testing.T) {

	// initialization
	nMq := NewMonitoredQueue(10)
	for i := 0; i < 6; i++ {
		nMq.Add(PacketType{})
	}
	for i := 0; i < 3; i++ {
		nMq.GetFirst()
	}
	expectedLoad := 0.3
	// measuring
	gotLoad := nMq.Load()
	if gotLoad != expectedLoad {
		t.Errorf("got: %v, expected: %v", gotLoad, expectedLoad)
	}
}

func TestMonitoredQueue_RwRate(t *testing.T) {
	// initialization
	nMq := NewMonitoredQueue(10)
	for i := 0; i < 6; i++ {
		nMq.Add(PacketType{})
	}
	for i := 0; i < 3; i++ {
		nMq.GetFirst()
	}
	// measuring
	time.Sleep(1 * time.Second)
	if rateIn, rateOut := nMq.RwRate(); rateIn > 6 || rateIn < 5.98 || rateOut > 3 || rateOut < 2.98 {
		t.Errorf("got: rateIn=%f rateOut=%f ,expected: rateIn=[5.98-6] rateOut=[2.98-3]", rateIn, rateOut)
	}
}

func TestMonitoredQueue_Close(t *testing.T) {
	nMq := NewMonitoredQueue(10)
	nMq.Close()
	if _, ok := nMq.GetFirst(); ok {
		t.Error("the channel is still open")
	}
}

func TestMonitoredQueue_MonitorLoad(t *testing.T) {
	nMq := NewMonitoredQueue(2)
	nMq.MonitorLoad()
	// initially empty
	if nMq.writeDelay != 0 {
		t.Errorf("got %d, expected delay 0", nMq.writeDelay)
	}
	nMq.Add(PacketType{A: 1})
	nMq.Add(PacketType{A: 2})
	nMq.MonitorLoad()
	// queue full
	if nMq.writeDelay != time.Millisecond {
		t.Errorf("got %v, expected delay 1ms", nMq.writeDelay)
	}
	nMq.GetFirst()
	nMq.GetFirst()
	nMq.MonitorLoad()
	// empty again
	if nMq.writeDelay != 0 {
		t.Errorf("got %d, expected delay 0", nMq.writeDelay)
	}
}

func TestMonitoredQueue_SetReadDelay(t *testing.T) {
	nMq := MonitoredQueue{readDelay: 100}
	nMq.SetReadDelay(10 * time.Second)
	if nMq.readDelay != 10*time.Second {
		t.Errorf("got %v, expected 10s", nMq.readDelay)
	}
}

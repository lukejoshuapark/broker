package broker

import (
	"fmt"
	"testing"
	"time"

	"github.com/lukejoshuapark/test"
	"github.com/lukejoshuapark/test/does"
	"github.com/lukejoshuapark/test/is"
)

func TestBrokerUnsubscribe(t *testing.T) {
	// Arrange.
	broker := NewBroker[int]()
	out1 := broker.Subscribe()
	out2 := broker.Subscribe()

	// Act.
	broker.In() <- 1
	broker.In() <- 2
	time.Sleep(time.Millisecond * 50)

	broker.Unsubscribe(out2)
	broker.In() <- 3
	broker.Close()

	// Assert.
	sum1 := 0
	for v := range out1 {
		sum1 += v
	}

	sum2 := 0
	for v := range out2 {
		sum2 += v
	}

	test.That(t, sum1, is.EqualTo(6))
	test.That(t, sum2, is.EqualTo(3))
}

func TestBrokerReceiveBeforeClose(t *testing.T) {
	// Arrange.
	broker := NewBroker[int]()
	out := broker.Subscribe()

	// Act.
	broker.In() <- 1
	value := <-out
	broker.Close()
	_, ok := <-out

	// Assert.
	test.That(t, value, is.EqualTo(1))
	test.That(t, ok, is.False)
}

func TestBrokerReceiveAfterClose(t *testing.T) {
	// Arrange.
	broker := NewBroker[int]()
	out := broker.Subscribe()

	// Act.
	broker.In() <- 1
	broker.Close()
	value := <-out
	_, ok := <-out

	// Assert.
	test.That(t, value, is.EqualTo(1))
	test.That(t, ok, is.False)
}

func TestBrokerNoSendAfterClose(t *testing.T) {
	// Arrange.
	broker := NewBroker[int]()
	out := broker.Subscribe()

	// Act.
	broker.Close()
	_, ok := <-out
	test.That(t, ok, is.False)

	action := func() {
		broker.In() <- 1
		fmt.Println("asd")
	}

	// Assert.
	test.That(t, action, does.Panic)
}

func TestBrokerNoSubscribeAfterClose(t *testing.T) {
	// Arrange.
	broker := NewBroker[int]()
	out := broker.Subscribe()

	// Act.
	broker.Close()
	_, ok := <-out
	test.That(t, ok, is.False)

	action := func() {
		broker.Subscribe()
	}

	// Assert.
	test.That(t, action, does.Panic)
}

func TestReadmeScenario(t *testing.T) {
	// Create a broker that broadcasts ints.
	broker := NewBroker[int]()

	// The type of these are just <-chan int.
	out1 := broker.Subscribe()
	out2 := broker.Subscribe()

	// Broadcast a number, then close the broker.
	broker.In() <- 42
	broker.Close()

	// It's still safe to read from these after the broker has closed.
	v1 := <-out1
	v2 := <-out2

	test.That(t, v1, is.EqualTo(42))
	test.That(t, v2, is.EqualTo(42))

	// The output channels will automatically close when the broker is closed
	// and there are no more values to be read from them.
	_, ok1 := <-out1
	_, ok2 := <-out2

	test.That(t, ok1, is.False)
	test.That(t, ok2, is.False)
}

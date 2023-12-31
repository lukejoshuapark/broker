![](icon.png)

# broker

Broker exposes a buffered, mpmc "fan-out" broker.

## Usage Example

```go
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
```

## Notes

It is important to never manually close any of the channels returned from the
various methods on the broker.  If you want to shutdown the broker, simply call
`Close()` on it.

The underlying implementation of this broker uses
[infchan](https://github.com/lukejoshuapark/infchan).

package event

var stream *eventStream

func init() {
	stream = newStream()
}

package gobot_sim

const PIN_ON = 0x01
const PIN_OFF = 0x00

type PinChangedEvent struct {
	Pin       string
	LastValue int
	Value     int
	source    interface{}
}

func (p PinChangedEvent) Source() interface{} {
	return p.source
}

type PinReadFunc func(pin string) (int, error)
type PinWriteFunc func(pin string, val byte) error
type PinChangedFunc func(ev PinChangedEvent) error

package gobot_sim

type AcceptedAdapterIntf interface {
	DigitalRead(pin string) (val int, err error)
	DigitalWrite(pin string, val byte) (err error)
}

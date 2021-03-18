package gobot_sim

import (
	"errors"
	"time"
)

const (
	GPIOACTION_UNDEFINED = iota
	GPIOACTION_ON
	GPIOACTION_OFF
	GPIOACTION_TOGGLE
	GPIOACTION_BUTTONPRESS
)

type PinFuncs struct {
	On   func(action *PinWriteAction) error
	Off  func(action *PinWriteAction) error
	Read func(action *PinWriteAction) (int, error)
}

type PinWriteAction struct {
	pin    string
	action int

	onValue  byte
	offValue byte

	pinFuncs *PinFuncs
}

func NewPinWriteAction(pin string, action int, pinFuncs *PinFuncs) PinWriteAction {
	return PinWriteAction{
		pin:      pin,
		action:   action,
		pinFuncs: pinFuncs,
		onValue:  0x01,
		offValue: 0x00,
	}
}

func (ac PinWriteAction) Pin() string {
	return ac.pin
}

func (ac PinWriteAction) Action() int {
	return ac.action
}

func (ac PinWriteAction) OnValue() byte {
	return ac.onValue
}

func (ac PinWriteAction) OffValue() byte {
	return ac.offValue
}

func (ac *PinWriteAction) Execute() error {
	switch ac.action {
	case GPIOACTION_ON:
		return ac.pinFuncs.On(ac)
	case GPIOACTION_OFF:
		return ac.pinFuncs.Off(ac)
	case GPIOACTION_TOGGLE:
		val, _ := ac.pinFuncs.Read(ac)
		//if err != nil {
		//	return err
		//}
		var byteNum byte = 0x00
		byteNum = byte(val)
		if byteNum == ac.offValue {

			return ac.pinFuncs.On(ac)
		}
		return ac.pinFuncs.Off(ac)
	case GPIOACTION_BUTTONPRESS:
		ac.pinFuncs.On(ac)
		time.Sleep(time.Millisecond * 300)
		return ac.pinFuncs.Off(ac)
	}
	return errors.New("action not handled")
}

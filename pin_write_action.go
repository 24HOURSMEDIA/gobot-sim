package gobot_sim

import (
	"errors"
	"time"
)

const (
	PWACTION_UNDEFINED = iota
	PWACTION_ON
	PWACTION_OFF
	PWACTION_TOGGLE
	PWACTION_BUTTONPRESS
)

type PinFuncs struct {
	Write PinWriteFunc
	Read  PinReadFunc
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
		onValue:  PIN_ON,
		offValue: PIN_OFF,
	}
}

// Pin returns the pin number
func (ac PinWriteAction) Pin() string {
	return ac.pin
}

// Pin returns the action constant (i.e. PWACTION_TOGGLE etc)
func (ac PinWriteAction) Action() int {
	return ac.action
}

// Execute is called by the owner when an acton on
// a pin must be executed
func (ac *PinWriteAction) Execute() error {
	switch ac.action {
	case PWACTION_ON:
		return ac.pinFuncs.Write(ac.pin, ac.onValue)
	case PWACTION_OFF:
		return ac.pinFuncs.Write(ac.pin, ac.offValue)
	case PWACTION_TOGGLE:
		val, _ := ac.pinFuncs.Read(ac.pin)
		//if err != nil {
		//	return err
		//}
		var byteNum byte = 0x00
		byteNum = byte(val)
		if byteNum == ac.offValue {
			return ac.pinFuncs.Write(ac.pin, ac.onValue)
		}
		return ac.pinFuncs.Write(ac.pin, ac.offValue)
	case PWACTION_BUTTONPRESS:
		ac.pinFuncs.Write(ac.pin, ac.onValue)
		time.Sleep(time.Millisecond * 300)
		return ac.pinFuncs.Write(ac.pin, ac.offValue)
	}
	return errors.New("action not handled")
}

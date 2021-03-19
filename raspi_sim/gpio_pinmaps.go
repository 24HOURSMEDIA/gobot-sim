package raspi_sim

import "errors"

// GPIOToPinMap maps gpio numbers to pins for a Raspberry revision
type PinToGPIOMap struct {
	mapping map[string]string
}

func NewPinToGPIOMap(mapping map[string]string) *PinToGPIOMap {
	m := &PinToGPIOMap{
		mapping: mapping,
	}
	return m
}

func (m *PinToGPIOMap) ToGPIO(pin string) (string, error) {
	gpio, valid := m.mapping[pin]
	if !valid {
		return "", errors.New("Pin does not support GPIO")
	}
	return gpio, nil
}

// RPI3GPIOPinMap is a mapping for the latest 40 pin raspberry revisions
var RPI3PinGPIOMap = NewPinToGPIOMap(map[string]string{
	// pin 3 -> gpio 2
	"3":  "2",
	"5":  "3",
	"7":  "4",
	"8":  "14",
	"10": "15",
	"11": "17",
	"12": "18",
	"13": "27",
	"15": "22",
	"16": "23",
	"18": "24",
	"19": "10",
	"21": "9",
	"22": "25",
	"23": "11",
	"24": "8",
	"26": "7",
	"27": "0",
	"28": "1",
	"29": "5",
	"31": "6",
	"32": "12",
	"33": "13",
	"35": "19",
	"36": "16",
	"37": "26",
	"38": "20",
	"40": "21",
})

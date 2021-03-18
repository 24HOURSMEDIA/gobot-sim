package raspi_sim

import (
	"fmt"
	"github.com/24hoursmedia/gobot-sim"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/keyboard"
	"gobot.io/x/gobot/platforms/raspi"
	"gobot.io/x/gobot/sysfs"
	"strconv"
)

type GobotSimulator struct {
	gpioKeymap map[rune]gobot_sim.PinWriteAction
	verbosity  int
	adapter    *raspi.Adaptor
	logger     gobot_sim.VerbosityLogger
	name       string
}

// NewGobotSimulator creates a bot that makes your machine
// behave like a raspberry pi in some ways
func NewGobotSimulator(adapter *raspi.Adaptor) *GobotSimulator {
	sim := &GobotSimulator{}
	sim.gpioKeymap = map[rune]gobot_sim.PinWriteAction{}
	sim.adapter = adapter
	sim.name = "GobotSim"
	sim.logger.Prefix = sim.name
	return sim
}

// Verbosity sets the verbosity level of messages to stdout
func (sim *GobotSimulator) Verbosity(verbosity int) {
	sim.logger.Verbosity = verbosity
}

// AllGPIOPins returns an array of all GPIO pins that can be fed to EnterSimulationMode
// for quick testing
func (sim *GobotSimulator) AllGPIOPins() []string {
	count := 27
	pins := make([]string, count)
	for i := 1; i <= count; i++ {
		pins[i-1] = fmt.Sprintf("%d", i)
	}
	return pins
}

// EnterSimulationMode sets up the local machine and hooks into the file system
// to intercept specific gpio pins. Note that these are GPIO pin numbers, not board pin numbers
func (sim *GobotSimulator) EnterSimulationMode(gpioPins []string) {
	var files = make([]string, 0)
	files = append(files, "/sys/class/gpio/export")
	files = append(files, "/sys/class/gpio/unexport")
	for _, gpioPinNum := range gpioPins {
		files = append(files, fmt.Sprintf("/sys/class/gpio/gpio%s/direction", gpioPinNum))
		files = append(files, fmt.Sprintf("/sys/class/gpio/gpio%s/value", gpioPinNum))
	}
	fs := sysfs.NewMockFilesystem(files)
	sysfs.SetFilesystem(fs)
	sysfs.SetSyscall(&sysfs.MockSyscall{})
}

// MapKeyPressToGPIOAction maps a key press to a specific action on a pin, for example
// to turn it on or simulate a button press and release
func (sim *GobotSimulator) MapKeyPressToGPIOAction(key rune, pin string, action int) error {
	sim.logger.Debug("Mapping key %s to pin %s", strconv.QuoteRune(key), pin)

	pinFuncs := &gobot_sim.PinFuncs{
		On:   sim.actionPinOn,
		Off:  sim.actionPinOff,
		Read: sim.actionPinRead,
	}

	sim.gpioKeymap[key] = gobot_sim.NewPinWriteAction(pin, action, pinFuncs)
	return nil
}

// actionPinOn is the handler passed to PinWriteActions so it has access to the local context
func (sim *GobotSimulator) actionPinOn(ac *gobot_sim.PinWriteAction) error {
	return sim.adapter.DigitalWrite(ac.Pin(), ac.OnValue())
}

// actionPinOff is the handler passed to PinWriteActions so it has access to the local context
func (sim *GobotSimulator) actionPinOff(ac *gobot_sim.PinWriteAction) error {
	return sim.adapter.DigitalWrite(ac.Pin(), ac.OffValue())
}

// actionPinRead is the handler passed to PinWriteActions so it has access to the local context
func (sim *GobotSimulator) actionPinRead(ac *gobot_sim.PinWriteAction) (int, error) {
	return sim.adapter.DigitalRead(ac.Pin())
}

// Run sets up the simulator bot and starts it
func (sim *GobotSimulator) Run() error {
	keys := keyboard.NewDriver()
	work := func() {
		keys.On(keyboard.Key, func(data interface{}) {
			key := data.(keyboard.KeyEvent)
			if action, ok := sim.gpioKeymap[rune(key.Key)]; ok {
				sim.logger.Debug("Key %s pressed, Pin %s, Action %d", strconv.QuoteRune(rune(key.Key)), action.Pin(), action.Action())
				err := action.Execute()
				if err != nil {
					sim.logger.Error("Error %v", err)
				}
			}
		})
	}

	robot := gobot.NewRobot(sim.name,
		[]gobot.Connection{},
		[]gobot.Device{keys},
		work,
	)
	sim.logger.Info("Waiting for keypress")
	robot.Start()

	return nil
}

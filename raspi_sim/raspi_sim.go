package raspi_sim

import (
	"errors"
	"fmt"
	"github.com/24hoursmedia/gobot-sim"
	"github.com/24hoursmedia/gobot-sim/hybrid_sysfs"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/keyboard"
	"gobot.io/x/gobot/platforms/raspi"
	"gobot.io/x/gobot/sysfs"
	"strconv"
	"time"
)

type GobotSimulator struct {
	name          string
	adapter       *raspi.Adaptor
	pinToGPIOMap  *PinToGPIOMap
	gpioKeymap    map[rune]*gobot_sim.PinWriteAction
	gpioWatchers  []*gobot_sim.PinWatcher
	watchInterval time.Duration
	usedGPIOPins  map[string]bool
	logger        gobot_sim.VerbosityLogger
}

// NewGobotSimulator creates a bot that makes your machine
// behave like a raspberry pi in some ways
func NewGobotSimulator(adapter *raspi.Adaptor) *GobotSimulator {
	sim := &GobotSimulator{}
	sim.name = "GobotSim"
	sim.pinToGPIOMap = RPI3PinGPIOMap
	sim.gpioKeymap = map[rune]*gobot_sim.PinWriteAction{}
	sim.adapter = adapter
	sim.logger.Prefix = sim.name + " "
	sim.watchInterval = time.Millisecond * 10
	sim.usedGPIOPins = make(map[string]bool)
	return sim
}

// Verbosity sets the verbosity level of messages to stdout
func (sim *GobotSimulator) Verbosity(verbosity int) {
	sim.logger.Verbosity = verbosity
}

// SetPinToGPIOMap sets a pin mapping to gpio numbers for the platform (defaults to RPI3 mapping).
func (sim *GobotSimulator) SetPinToGPIOMap(pinToGPIO *PinToGPIOMap) {
	sim.pinToGPIOMap = pinToGPIO
}

// enterSimulationMode sets up the local machine and hooks into the file system
// to intercept specific gpio pins. Note that these are GPIO pin numbers, not board pin numbers
func (sim *GobotSimulator) enterSimulationMode() {
	fs := hybrid_sysfs.NewHybridFs(
		&sysfs.NativeFilesystem{},
		sysfs.NewMockFilesystem([]string{}),
	)
	fs.AddMockablePath("/sys/class/gpio/export")
	fs.AddMockablePath("/sys/class/gpio/unexport")
	for gpioPinNum, _ := range sim.usedGPIOPins {
		sim.logger.Debug("entersim - hooking into GPIO%s", gpioPinNum)

		fs.AddMockablePath(fmt.Sprintf("/sys/class/gpio/gpio%s/direction", gpioPinNum))
		fs.AddMockablePath(fmt.Sprintf("/sys/class/gpio/gpio%s/value", gpioPinNum))
	}
	sysfs.SetFilesystem(fs)
	sysfs.SetSyscall(&hybrid_sysfs.HybridSyscall{})
}

// usePinForGPIO tells the simulator to use a pin for GPIO
// this runs the pin through the simulator instead of the HW board
func (sim *GobotSimulator) usePinForGPIO(pin string) error {
	// translate pin to gpio num and map it so we know it is used
	gpioPin, pinErr := sim.pinToGPIOMap.ToGPIO(pin)
	if pinErr != nil {
		sim.logger.Error("usepin  %s - no GPIO", pin)
		return pinErr
	}
	sim.logger.Debug("usepin  %s - this is GPIO %s", pin, gpioPin)
	sim.usedGPIOPins[gpioPin] = true
	return nil
}

// MapKeyPressToGPIOAction maps a key press to a specific action on a pin, for example
// to turn it on or simulate a button press and release
func (sim *GobotSimulator) MapKeyPressToGPIOAction(key rune, pin string, action int) (*gobot_sim.PinWriteAction, error) {
	sim.logger.Debug("Mapping key %s to pin %s", strconv.QuoteRune(key), pin)

	usePinErr := sim.usePinForGPIO(pin)
	if usePinErr != nil {
		return nil, usePinErr
	}

	pinFuncs := &gobot_sim.PinFuncs{Write: sim.pinWrite, Read: sim.pinRead}
	sim.gpioKeymap[key] = gobot_sim.NewPinWriteAction(pin, action, pinFuncs)
	return sim.gpioKeymap[key], nil
}

// WatchPin intercepts writes to a pin and calls a function if the value changed
func (sim *GobotSimulator) WatchPin(pin string, handler gobot_sim.PinChangedFunc) (*gobot_sim.PinWatcher, error) {
	sim.logger.Debug("add watcher for pin %s", pin)

	usePinErr := sim.usePinForGPIO(pin)
	if usePinErr != nil {
		return nil, usePinErr
	}

	watchFuncs := &gobot_sim.WatchFuncs{Read: sim.pinRead, Changed: handler}
	watcher := gobot_sim.NewPinWatcher(pin, watchFuncs)
	sim.gpioWatchers = append(sim.gpioWatchers, watcher)
	return watcher, nil
}

// pinWrite is the handler passed to PinWrite/ReadActions so it has access to the local context
func (sim *GobotSimulator) pinWrite(pin string, v byte) error {
	return sim.adapter.DigitalWrite(pin, v)
}

// pinRead is the handler passed to PinWriteActions so it has access to the local context
func (sim *GobotSimulator) pinRead(pin string) (int, error) {
	return sim.adapter.DigitalRead(pin)
}

// Run sets up the simulator bot and starts it
func (sim *GobotSimulator) Run() error {
	sim.enterSimulationMode()
	go sim.goRun()
	return nil
}

func Stop() error {
	return errors.New("Not supported")
}

// goRun is the go routine for running
func (sim *GobotSimulator) goRun() error {
	keys := keyboard.NewDriver()
	work := func() {
		if len(sim.gpioKeymap) > 0 {
			sim.logger.Debug("Setup keypress handlers")
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
		if len(sim.gpioWatchers) > 0 {
			sim.logger.Debug("Setup watchers")
			gobot.Every(sim.watchInterval, func() {
				for _, w := range sim.gpioWatchers {
					w.Observe()
				}
			})
		}
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

func Close() error {
	return nil
}

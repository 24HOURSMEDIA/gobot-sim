package raspi_sim

import (
	"fmt"
	"github.com/24hoursmedia/gobot-sim"
	"github.com/24hoursmedia/gobot-sim/hybrid_sysfs"
	"github.com/rs/zerolog/log"
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
}

// NewGobotSimulator creates a bot that makes your machine
// behave like a raspberry pi in some ways
func NewGobotSimulator(adapter *raspi.Adaptor) *GobotSimulator {
	sim := &GobotSimulator{}
	sim.name = "GobotSim"
	sim.pinToGPIOMap = RPI3PinGPIOMap
	sim.gpioKeymap = map[rune]*gobot_sim.PinWriteAction{}
	sim.adapter = adapter
	sim.watchInterval = time.Millisecond * 20
	sim.usedGPIOPins = make(map[string]bool)
	log.Debug().Str("name", sim.name).Msg("Created new Gobot-Sim")
	return sim
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
		log.Debug().Str("gpio", gpioPinNum).Msg("entersim - hooking into GPIO")
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
		log.Error().Str("pin", pin).Msg("entersim - hooking pin into GPIO")
		return pinErr
	}

	sim.usedGPIOPins[gpioPin] = true
	return nil
}

// AddKeyPressPWAction writes something to a pin when a key is pressed.
// It maps a key press to a specific action on a pin, for example
// to turn it on or simulate a button press and release
func (sim *GobotSimulator) AddKeyPressPWAction(key rune, pin string, action int) (*gobot_sim.PinWriteAction, error) {
	log.Debug().Str("pin", pin).Msg("entersim - hooking pin into GPIO")
	log.Debug().Str("key", strconv.QuoteRune(key)).Str("pin", pin).
		Msg("Mapping key")

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
	log.Debug().Msgf("add watcher for pin %s", pin)

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
	return nil
}

// goRun is the go routine for running
func (sim *GobotSimulator) goRun() error {
	keys := keyboard.NewDriver()
	work := func() {
		if len(sim.gpioKeymap) > 0 {
			log.Debug().Int("count", len(sim.gpioKeymap)).Msg("Setup keypress handlers")
			keys.On(keyboard.Key, func(data interface{}) {
				key := data.(keyboard.KeyEvent)
				if action, ok := sim.gpioKeymap[rune(key.Key)]; ok {
					log.Debug().Str("key", strconv.QuoteRune(rune(key.Key))).Str("pin", action.Pin()).
						Int("action", action.Action()).Msg("Key pressed")
					err := action.Execute()
					if err != nil {
						log.Err(err).Msg("")
					}
				}
			})
		}
		if len(sim.gpioWatchers) > 0 {
			log.Info().Msg("Setup watchers")
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
	log.Info().
		Int("num_pin_watchers", len(sim.gpioWatchers)).
		Int("num_keypress_watchers", len(sim.gpioKeymap)).
		Msg("Simulator ready")
	robot.Start()
	return nil
}

func Close() error {
	return nil
}

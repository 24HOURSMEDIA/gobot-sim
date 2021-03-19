// +build example
//
// Do not build by default.

package main

import (
	"fmt"
	"github.com/24hoursmedia/gobot-sim"
	"github.com/24hoursmedia/gobot-sim/raspi_sim"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
	"time"
)

func main() {
	fmt.Println("Setting up simulator - watcher")

	ledPin := "11"

	// set up gobot
	r := raspi.NewAdaptor()
	led := gpio.NewLedDriver(r, ledPin)
	work := func() {
		gobot.Every(1*time.Second, func() {
			led.Toggle()
		})
	}
	robot := gobot.NewRobot("ledBot",
		[]gobot.Connection{r},
		[]gobot.Device{led},
		work,
	)

	// hook in the simulator.
	sim := raspi_sim.NewGobotSimulator(r)
	sim.Verbosity(gobot_sim.VERBOSITY_VVV)
	// a simple watcher for the pin that on a real board has a led attached
	sim.WatchPin(ledPin, func(ev gobot_sim.PinChangedEvent) error {
		fmt.Println("LED BLINKS")
		return nil
	})

	// a more advanced, generic watcher
	ledWatcher, _ := sim.WatchPin(ledPin, func(ev gobot_sim.PinChangedEvent) error {
		source := ev.Source().(*gobot_sim.PinWatcher)
		fmt.Printf("%s, pin %s, state = %d\n",
			source.Name(),
			ev.Pin,
			ev.Value,
		)
		return nil
	})
	ledWatcher.SetName("LED blink watcher")
	sim.Run()

	// start the 'real' robot
	robot.Start()
}

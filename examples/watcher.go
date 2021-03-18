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
	ledGPIO := "17"

	// set up gobopt
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

	// hook in the simulator. It links keypress '1' to a simulation of a button press and release
	// on pin 11 (GPIO 17)
	sim := raspi_sim.NewGobotSimulator(r)
	sim.Verbosity(gobot_sim.VERBOSITY_VVV)
	sim.EnterSimulationMode([]string{ledGPIO})
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
	go sim.Run()

	// start the 'real' robot
	robot.Start()
}

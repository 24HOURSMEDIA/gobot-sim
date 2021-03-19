// +build example
//
// Do not build by default.

package main

import (
	"github.com/24hoursmedia/gobot-sim"
	"github.com/24hoursmedia/gobot-sim/raspi_sim"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
	"os"
	"time"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	log.Info().Msg("Setting up simulator - button press")

	// set up rasberry pi gobot with a button attached to pin 11 (GPIO17)
	r := raspi.NewAdaptor()
	button := gpio.NewButtonDriver(r, "11", time.Millisecond*20)
	work := func() {
		button.On(gpio.ButtonPush, func(data interface{}) {
			log.Info().Msg("button pressed")
		})
		button.On(gpio.ButtonRelease, func(data interface{}) {
			log.Info().Msg("button released")
		})
	}

	robot := gobot.NewRobot("buttonBot",
		[]gobot.Connection{r},
		[]gobot.Device{button},
		work,
	)

	// hook in the simulator. It links keypress '1' to a simulation of a button press and release
	// on pin 11 (GPIO 17)
	sim := raspi_sim.NewGobotSimulator(r)
	sim.AddKeyPressPWAction('1', "11", gobot_sim.PW_ACTION_BUTTONPRESS)
	sim.Run()

	// start the 'real' robot
	robot.Start()
}

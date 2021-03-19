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
	log.Info().Msg("Setting up simulator - watcher")

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
	// a simple watcher for the pin that on a real board has a led attached
	sim.WatchPin(ledPin, func(ev gobot_sim.PinChangedEvent) error {
		log.Info().Msg("LED BLINKS (1)")
		return nil
	})

	// a more advanced, generic watcher
	ledWatcher, _ := sim.WatchPin(ledPin, func(ev gobot_sim.PinChangedEvent) error {
		source := ev.Source().(*gobot_sim.PinWatcher)
		log.Info().Str("source", source.Name()).Str("pin", ev.Pin).Int("pinVal", ev.Value).
			Msg("LED BLINKS (2)")

		return nil
	})
	ledWatcher.SetName("LED blink watcher")
	sim.Run()

	// start the 'real' robot
	robot.Start()
}

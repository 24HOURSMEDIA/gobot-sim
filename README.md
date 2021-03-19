# gobot-sim

**Simulate GPIO for [Gobot](https://gobot.io) with your keyboard and your own code**  

* Log keystrokes to send vitual pin inputs to your Gobot application  
  (your keyboard replaces an IN pin)
* Hook into OUT pins so they pass through your code instead of the hardware  
  (your code replaces a device connected to an OUT pin)
  
[View the example code.](examples/)

Gobot-Sim can simulate Raspberry Pi GPIO pin inputs on a local development
machine such as a Mac.
It links specific keyboard input to GPIO inputs captured by GoBot.

It facilitates local development and testing of Gobot applications
on non-raspberry machines.

For example, consider a push button which should be wired to GPIO 11.

With Gobot-sim, you can run your application in the terminal, and
activate the button press by linking it to a keyboard shortcut.

## Examples

* [Simulate a button connected to a GPIO pin with your keyboard](examples/button.go)
* [Log the status from a led to the console instead of sending it to real GPIO](examples/led.go)

Run this on your Mac or Linux machine.
You can toggle the button by pressing the '1' key.

```go
package main

import (
	"fmt"
	"github.com/24hoursmedia/gobot_sim"
	"github.com/24hoursmedia/gobot_sim/raspi_sim"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
	"time"
)

func main() {
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
```

![Example](resources/doc/example.png "Example output")

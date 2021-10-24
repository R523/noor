package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pterm/pterm"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/host/v3"
	"periph.io/x/host/v3/rpi"
)

const (
	// the i2c address which can be find by i2cdetect -y 1.
	I2CAddr = 0x48
	A0      = 0x40

	LEDFreq = 100 * physic.Hertz
)

func main() {
	if err := pterm.DefaultBigText.WithLetters(
		pterm.NewLettersFromStringWithStyle("No", pterm.NewStyle(pterm.FgCyan)),
		pterm.NewLettersFromStringWithStyle("or", pterm.NewStyle(pterm.FgLightRed)),
	).Render(); err != nil {
		_ = err
	}

	if _, err := host.Init(); err != nil {
		pterm.Error.Printf("host initiation failed %s\n", err)

		return
	}

	b, err := i2creg.Open("/dev/i2c-1")
	if err != nil {
		pterm.Error.Printf("cannot open i2c device %s\n", err)

		return
	}
	defer b.Close()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-quit:
			_ = rpi.P1_33.Halt()

			return
		case <-ticker.C:
			d := make([]byte, 1)

			if err := b.Tx(I2CAddr, []byte{A0, 0x0}, d); err != nil {
				pterm.Error.Printf("cannot communicate with i2c device %s\n", err)

				return
			}

			pterm.Info.Printf("light: %d\n", d[0])

			if err := rpi.P1_33.PWM(gpio.DutyMax>>d[0], LEDFreq); err != nil {
				pterm.Error.Printf("cannot setup pwm for led %s\n", err)

				return
			}
		}
	}
}

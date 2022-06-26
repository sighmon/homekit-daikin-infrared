package main

import (
	"github.com/brutella/hap"
	"github.com/brutella/hap/accessory"
	"github.com/chbmuc/lirc"

	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var developmentMode bool
var currentHeaterCoolerState int
var currentHeatingThresholdTemperature float64

func init() {
	flag.BoolVar(&developmentMode, "dev", false, "development mode, so ignore LIRC setup")
	flag.Parse()
}

func main() {
	// Initialize with path to lirc socket
	ir, err := lirc.Init("/var/run/lirc/lircd")
	if err != nil && developmentMode == false {
		panic(err)
	}

	// Create the Daikin heater accessory.
	a := accessory.NewHeater(accessory.Info {
		Name: "Daikin air conditioner",
		SerialNumber: "FTXS50KAVMA",
		Manufacturer: "Daikin",
		Model: "FTXS50KAVMA",
		Firmware: "1.0.0",
	})

	// TODO: read room temperature from a sensor
	a.Heater.CurrentTemperature.SetValue(19)

	// Set target state to auto
	currentHeaterCoolerState = 0
	a.Heater.TargetHeaterCoolerState.SetValue(currentHeaterCoolerState)

	// Set target temperature
	currentHeatingThresholdTemperature = 23.0
	a.Heater.HeatingThresholdTemperature.SetValue(currentHeatingThresholdTemperature)
	a.Heater.HeatingThresholdTemperature.SetStepValue(1.0)
	a.Heater.HeatingThresholdTemperature.SetMinValue(18)
	a.Heater.HeatingThresholdTemperature.SetMaxValue(26)

	a.Heater.Active.OnValueRemoteUpdate(func(on int) {
		if on == 1 {
			log.Println("Sending power on command")
			powerOnCommand := "daikin POWER_ON"
			if currentHeaterCoolerState == 1 {
				powerOnCommand = "daikin POWER_ON_HEAT"
				currentHeatingThresholdTemperature = 25.0
				a.Heater.HeatingThresholdTemperature.SetValue(currentHeatingThresholdTemperature)
			}
			err = ir.Send(powerOnCommand)
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Println("Sending power off command")
			err = ir.Send("daikin POWER_OFF")
			if err != nil {
				log.Println(err)
			}
		}
	})

	a.Heater.HeatingThresholdTemperature.OnValueRemoteUpdate(func(value float64) {
		currentHeatingThresholdTemperature = value
		state := "AUTO"
		if currentHeaterCoolerState == 1 {
			state = "HEAT"
		}
		log.Println(fmt.Sprintf("Sending target temperature command: %f°C %s", value, state))
		err = ir.Send(fmt.Sprintf("daikin TEMPERATURE_%s_%d", state, int(value)))
		if err != nil {
			log.Println(err)
		}
	})

	a.Heater.TargetHeaterCoolerState.OnValueRemoteUpdate(func(value int) {
		currentHeaterCoolerState = value
		if value == 0 {
			log.Println("Sending target state command: Auto")
			err = ir.Send("daikin MODE_AUTO")
			if err != nil {
				log.Println(err)
			}
		} else if value == 1 {
			log.Println("Sending target state command: Heat")
			err = ir.Send("daikin MODE_HEAT")
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Println("Target state command: Unknown")
		}
	})

	// Store the data in the "./db" directory.
	fs := hap.NewFsStore("./db")

	// Create the hap server.
	server, err := hap.NewServer(fs, a.A)
	if err != nil {
		// stop if an error happens
		log.Panic(err)
	}

	// Setup a listener for interrupts and SIGTERM signals
	// to stop the server.
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<- c
		// Stop delivering signals.
		signal.Stop(c)

		// Cancel the context to stop the server.
		cancel()
	}()

	// Run the server.
	server.ListenAndServe(ctx)
}

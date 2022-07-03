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

var currentHeaterCoolerState int
var currentHeatingThresholdTemperature float64
var developmentMode bool
var dyson bool
var name string

func init() {
	flag.BoolVar(&developmentMode, "dev", false, "development mode, so ignore LIRC setup")
	flag.BoolVar(&dyson, "dyson", false, "Dyson AM09 mode")
	flag.Parse()
}

func main() {
	// Initialize with path to lirc socket
	ir, err := lirc.Init("/var/run/lirc/lircd")
	if err != nil && developmentMode == false {
		panic(err)
	}

	info := accessory.Info{
		Name: "Daikin air conditioner",
		SerialNumber: "FTXS50KAVMA",
		Manufacturer: "Daikin",
		Model: "FTXS50KAVMA",
		Firmware: "1.0.0",
	}
	name = "daikin"

	if dyson {
		info.Name = "Dyson Hot+Cool"
		info.SerialNumber = "AM09"
		info.Manufacturer = "Dyson"
		info.Model = "AM09"
		name = "dyson-am09"
	}

	// Create the heater accessory.
	a := accessory.NewHeater(info)

	// TODO: read room temperature from a sensor
	// a.Heater.CurrentTemperature.SetValue(19)

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
			powerOnCommand := fmt.Sprintf("%s POWER_ON", name)
			if currentHeaterCoolerState == 1 {
				powerOnCommand = fmt.Sprintf("%s POWER_ON_HEAT", name)
				currentHeatingThresholdTemperature = 25.0
				a.Heater.HeatingThresholdTemperature.SetValue(currentHeatingThresholdTemperature)
			}
			err = ir.Send(powerOnCommand)
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Println("Sending power off command")
			err = ir.Send(fmt.Sprintf("%s POWER_OFF", name))
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
		log.Println(fmt.Sprintf("Sending target temperature command: %f°C %s", currentHeatingThresholdTemperature, state))
		err = ir.Send(fmt.Sprintf("daikin TEMPERATURE_%s_%d", state, int(currentHeatingThresholdTemperature)))
		if err != nil {
			log.Println(err)
		}
	})

	a.Heater.TargetHeaterCoolerState.OnValueRemoteUpdate(func(value int) {
		currentHeaterCoolerState = value
		state := "AUTO"
		if currentHeaterCoolerState == 1 {
			state = "HEAT"
		}
		log.Println(fmt.Sprintf("Sending target mode command: %f°C %s", currentHeatingThresholdTemperature, state))
		err = ir.Send(fmt.Sprintf("daikin TEMPERATURE_%s_%d", state, int(currentHeatingThresholdTemperature)))
		if err != nil {
			log.Println(err)
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

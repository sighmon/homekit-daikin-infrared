package main

import (
	"github.com/brutella/hap"
	"github.com/brutella/hap/accessory"

	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Create the Daikin heater accessory.
	a := accessory.NewHeater(accessory.Info {
		Name: "Daikin air conditioner",
		SerialNumber: "FTXS50KAVMA",
		Manufacturer: "Daikin",
		Model: "FTXS50KAVMA",
		Firmware: "1.0.0",
	})

	// TODO: read from temperature sensor
	a.Heater.CurrentTemperature.SetValue(19)

	// Set target state to auto
	a.Heater.TargetHeaterCoolerState.SetValue(0)

	// Set target temperature
	a.Heater.HeatingThresholdTemperature.SetValue(23)

	a.Heater.Active.OnValueRemoteUpdate(func(on int) {
		if on == 1 {
			log.Println("TODO: send on command")
		} else {
			log.Println("TODO: send off command")
		}
	})

	a.Heater.HeatingThresholdTemperature.OnValueRemoteUpdate(func(value float64) {
		log.Println(fmt.Sprintf("TODO: send target temperature command: %f°C", value))
	})

	a.Heater.TargetHeaterCoolerState.OnValueRemoteUpdate(func(value int) {
		if value == 0 {
			log.Println("TODO: send target state command: Auto")
		} else if value == 1 {
			log.Println("TODO: send target state command: Heat")
		} else {
			log.Println("TODO: send target state command: Cool")
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

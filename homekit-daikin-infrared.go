package main

import (
	"github.com/brutella/hap"
	"github.com/brutella/hap/accessory"

	"context"
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

	a.Heater.TargetHeaterCoolerState.SetValue(23)

	a.Heater.Active.OnValueRemoteUpdate(func(on int) {
		if on == 1 {
			log.Println("TODO: send on command")
		} else {
			log.Println("TODO: send off command")
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

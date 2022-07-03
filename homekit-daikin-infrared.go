package main

import (
	"github.com/brutella/hap"
	"github.com/brutella/hap/accessory"
	"github.com/chbmuc/lirc"
	"github.com/d2r2/go-dht"

	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var acc *accessory.Heater
var currentHeaterCoolerState int
var currentHeatingThresholdTemperature float64
var developmentMode bool
var temperaturePin int

func init() {
	flag.BoolVar(&developmentMode, "dev", false, "development mode, so ignore LIRC setup")
	flag.IntVar(&temperaturePin, "temperaturePin", 0, "tempearture sensor GPIO pin, an int")
	flag.Parse()
}

func readTemperature() {
	if temperaturePin != 0 {
		for {
			temperature, humidity, retried, err := dht.ReadDHTxxWithRetry(dht.DHT22, temperaturePin, false, 10)
			if err != nil {
				log.Println(fmt.Sprintf("Failed to get temperature with error: %s", err))
			} else {
				log.Println(fmt.Sprintf("Temperature = %f°C, Humidity = %f% (retried %d times)", temperature, humidity, retried))
				acc.Heater.CurrentTemperature.SetValue(float64(temperature))
			}
			time.Sleep(5 * time.Second)
		}
	}
}

func main() {
	// Initialize with path to lirc socket
	ir, err := lirc.Init("/var/run/lirc/lircd")
	if err != nil && developmentMode == false {
		panic(err)
	}

	// Create the Daikin heater accessory.
	acc = accessory.NewHeater(accessory.Info {
		Name: "Daikin air conditioner",
		SerialNumber: "FTXS50KAVMA",
		Manufacturer: "Daikin",
		Model: "FTXS50KAVMA",
		Firmware: "1.0.0",
	})

	// Set target state to auto
	currentHeaterCoolerState = 0
	acc.Heater.TargetHeaterCoolerState.SetValue(currentHeaterCoolerState)

	// Set target temperature
	currentHeatingThresholdTemperature = 23.0
	acc.Heater.HeatingThresholdTemperature.SetValue(currentHeatingThresholdTemperature)
	acc.Heater.HeatingThresholdTemperature.SetStepValue(1.0)
	acc.Heater.HeatingThresholdTemperature.SetMinValue(18)
	acc.Heater.HeatingThresholdTemperature.SetMaxValue(26)

	acc.Heater.Active.OnValueRemoteUpdate(func(on int) {
		if on == 1 {
			log.Println("Sending power on command")
			powerOnCommand := "daikin POWER_ON"
			if currentHeaterCoolerState == 1 {
				powerOnCommand = "daikin POWER_ON_HEAT"
				currentHeatingThresholdTemperature = 25.0
				acc.Heater.HeatingThresholdTemperature.SetValue(currentHeatingThresholdTemperature)
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

	acc.Heater.HeatingThresholdTemperature.OnValueRemoteUpdate(func(value float64) {
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

	acc.Heater.TargetHeaterCoolerState.OnValueRemoteUpdate(func(value int) {
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
	server, err := hap.NewServer(fs, acc.A)
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

	// Read room temperature from a DHT22 temperature sensor
	go readTemperature()

	// Run the server.
	server.ListenAndServe(ctx)
}

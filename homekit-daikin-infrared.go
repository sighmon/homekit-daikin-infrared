package main

import (
	"github.com/brutella/hap"
	"github.com/brutella/hap/accessory"
	"github.com/brutella/hap/characteristic"
	"github.com/chbmuc/lirc"

	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

var currentFanSpeed float64
var currentHeaterCoolerState int
var currentHeatingThresholdTemperature float64
var currentSwingMode int
var developmentMode bool
var dyson bool
var fs hap.Store
var ir *lirc.Router
var lircName string
var fanSpeed *characteristic.RotationSpeed
var swingMode *characteristic.SwingMode

func init() {
	flag.BoolVar(&developmentMode, "dev", false, "development mode, so ignore LIRC setup")
	flag.BoolVar(&dyson, "dyson", false, "Dyson AM09 mode")
	flag.Parse()

	// Initialize with path to lirc socket
	lircIr, err := lirc.Init("/var/run/lirc/lircd")
	if err != nil && developmentMode == false {
		panic(err)
	}
	ir = lircIr

	// Store the data in the "./db" directory.
	fs = hap.NewFsStore("./db")

	// Load the previous state, or create defaults
	storedTemperature, err := fs.Get("currentHeatingThresholdTemperature")
	if err != nil {
		fs.Set("currentHeatingThresholdTemperature", []byte("23"))
		storedTemperature = []byte("23")
	}
	storedTemperatureInt, _ := strconv.Atoi(string(storedTemperature))
	currentHeatingThresholdTemperature = float64(storedTemperatureInt)
	storedHeaterState, err := fs.Get("currentHeaterCoolerState")
	if err != nil {
		fs.Set("currentHeaterCoolerState", []byte("0"))
		storedHeaterState = []byte("0")
	}
	storedHeaterCoolerStateInt, _ := strconv.Atoi(string(storedHeaterState))
	currentHeaterCoolerState = storedHeaterCoolerStateInt
	storedFanSpeed, err := fs.Get("currentFanSpeed")
	if err != nil {
		fs.Set("currentFanSpeed", []byte("5"))
		storedFanSpeed = []byte("5")
	}
	storedFanSpeedInt, _ := strconv.Atoi(string(storedFanSpeed))
	currentFanSpeed = float64(storedFanSpeedInt)
	storedSwingMode, err := fs.Get("currentSwingMode")
	if err != nil {
		fs.Set("currentSwingMode", []byte("0"))
		storedSwingMode = []byte("0")
	}
	storedSwingModeInt, _ := strconv.Atoi(string(storedSwingMode))
	currentSwingMode = storedSwingModeInt
}

func sendLircCommand(command string) {
	err := ir.Send(command)
	if err != nil {
		log.Println(err)
	}
}

func main() {
	info := accessory.Info{
		Name: "Daikin air conditioner",
		SerialNumber: "FTXS50KAVMA",
		Manufacturer: "Daikin",
		Model: "FTXS50KAVMA",
		Firmware: "1.0.0",
	}
	lircName = "daikin"

	if dyson {
		info.Name = "Dyson Hot+Cool"
		info.SerialNumber = "AM09"
		info.Manufacturer = "Dyson"
		info.Model = "AM09"
		lircName = "dyson-am09"
	}

	log.Println(fmt.Sprintf(
		"Starting up %s, state: %d, temperature: %f, fan: %f, swing mode: %d",
		lircName,
		currentHeaterCoolerState,
		currentHeatingThresholdTemperature,
		currentFanSpeed,
		currentSwingMode,
	))

	// Create the heater accessory.
	a := accessory.NewHeater(info)

	// TODO: read room temperature from a sensor
	// a.Heater.CurrentTemperature.SetValue(19)

	a.Heater.TargetHeaterCoolerState.SetValue(currentHeaterCoolerState)

	// Set target temperature
	a.Heater.HeatingThresholdTemperature.SetValue(currentHeatingThresholdTemperature)
	a.Heater.HeatingThresholdTemperature.SetStepValue(1.0)
	a.Heater.HeatingThresholdTemperature.SetMinValue(18)
	a.Heater.HeatingThresholdTemperature.SetMaxValue(26)
	if dyson {
		a.Heater.HeatingThresholdTemperature.SetMinValue(1)
		a.Heater.HeatingThresholdTemperature.SetMaxValue(37)
	}

	a.Heater.Active.OnValueRemoteUpdate(func(on int) {
		if on == 1 {
			log.Println("Sending power on command")
			powerOnCommand := fmt.Sprintf("%s POWER_ON", lircName)
			if currentHeaterCoolerState == 1 && dyson == false {
				powerOnCommand = fmt.Sprintf("%s POWER_ON_HEAT", lircName)
				currentHeatingThresholdTemperature = 25.0
			}
			a.Heater.HeatingThresholdTemperature.SetValue(currentHeatingThresholdTemperature)
			a.Heater.TargetHeaterCoolerState.SetValue(currentHeaterCoolerState)
			fanSpeed.SetValue(currentFanSpeed * 10)
			swingMode.SetValue(currentSwingMode)
			sendLircCommand(powerOnCommand)
		} else {
			log.Println("Sending power off command")
			sendLircCommand(fmt.Sprintf("%s POWER_OFF", lircName))
		}
	})

	a.Heater.HeatingThresholdTemperature.OnValueRemoteUpdate(func(value float64) {
		state := "AUTO"
		if currentHeaterCoolerState == 1 {
			state = "HEAT"
		}
		if dyson {
			command := fmt.Sprintf("%s HEAT_DOWN", lircName)
			if value > currentHeatingThresholdTemperature {
				command = fmt.Sprintf("%s HEAT_UP", lircName)
			}
			for i := 1; float64(i) <= math.Abs(value - currentHeatingThresholdTemperature); i++ {
				sendLircCommand(command)
			}
		} else {
			sendLircCommand(fmt.Sprintf("daikin TEMPERATURE_%s_%d", state, int(value)))
		}
		log.Println(fmt.Sprintf("Sending %s target temperature command: %f°C %s", lircName, value, state))
		currentHeatingThresholdTemperature = value
		fs.Set("currentHeatingThresholdTemperature", []byte(fmt.Sprintf("%d", int(currentHeatingThresholdTemperature))))
	})

	a.Heater.TargetHeaterCoolerState.OnValueRemoteUpdate(func(value int) {
		state := "AUTO"
		if value == 1 {
			state = "HEAT"
		}
		if dyson {
			if currentHeaterCoolerState != value {
				if value == 1 {
					sendLircCommand(fmt.Sprintf("%s HEAT_UP", lircName))
				} else {
					sendLircCommand(fmt.Sprintf("%s MODE_FAN", lircName))
				}
			}
		} else {
			sendLircCommand(fmt.Sprintf("daikin TEMPERATURE_%s_%d", state, int(value)))
		}
		log.Println(fmt.Sprintf("Sending %s target mode command: %f°C %s", lircName, currentHeatingThresholdTemperature, state))
		currentHeaterCoolerState = value
		fs.Set("currentHeaterCoolerState", []byte(fmt.Sprintf("%d", currentHeaterCoolerState)))
	})

	// Add Fan speed control
	fanSpeed = characteristic.NewRotationSpeed()
	fanSpeed.SetStepValue(10)
	fanSpeed.SetValue(currentFanSpeed * 10)
	fanSpeed.OnValueRemoteUpdate(func(value float64) {
		percentageToSpeed := value / 10
		if dyson {
			command := fmt.Sprintf("%s FAN_DOWN", lircName)
			if percentageToSpeed > currentFanSpeed {
				command = fmt.Sprintf("%s FAN_UP", lircName)
			}
			speedDifference := int(math.Abs(percentageToSpeed - currentFanSpeed))
			// Add 1 as the first fan speed command just shows the current speed on the display
			speedDifference += 1
			for i := 0; i < speedDifference; i++ {
				sendLircCommand(command)
			}
		}
		log.Println(fmt.Sprintf("Sending %s target fan speed command: %f%%", lircName, value))
		currentFanSpeed = percentageToSpeed
		fs.Set("currentFanSpeed", []byte(fmt.Sprintf("%d", int(currentFanSpeed))))
	})
	a.Heater.AddC(fanSpeed.C)

	// Add swing mode
	swingMode = characteristic.NewSwingMode()
	swingMode.SetValue(currentSwingMode)
	swingMode.OnValueRemoteUpdate(func(value int) {
		if dyson {
			if currentSwingMode != value {
				sendLircCommand(fmt.Sprintf("%s OSCILLATION", lircName))
			}
		} else {
			// TODO: Daikin swing mode
		}
		log.Println(fmt.Sprintf("Sending %s swing mode command: %d", lircName, value))
		currentSwingMode = value
		fs.Set("currentSwingMode", []byte(fmt.Sprintf("%d", currentSwingMode)))
	})
	a.Heater.AddC(swingMode.C)

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

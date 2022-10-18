package main

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
	"tinygo.org/x/bluetooth"
)

var mgr *Manager

var gyro = false

var suggestions = []prompt.Suggest{}

func completer(in prompt.Document) []prompt.Suggest {
	return nil
}

func executor(in string) {
	in = strings.TrimSpace(in)
	if in == "" {
		return
	}

	arg := strings.Split(in, " ")
	argc := len(arg)

	cmd := arg[0]

	switch cmd {
	case "light": // set the lights
		if argc != 2 {
			color.HiRed("usage: light n\n e.g. light 7")
			return
		}
		n, e := strconv.Atoi(arg[1])
		if e != nil {
			color.HiRed("pattern '%s' is not number", arg[1])
			return
		}

		for jc := range mgr.connected {
			// joycon.SetPlayerLights(jc, byte((n)<<4)) // flashing
			jc.SetLights(byte((n))) // stay on, 1 frame anyway
		}

		return
	case "connect": // connect to my right joycon
		var adapter = bluetooth.DefaultAdapter
		if e := adapter.Enable(); e != nil {
			panic(e)
		}

		mac, _ := bluetooth.ParseMAC("70:48:F7:76:BC:87")
		a := bluetooth.Address{MACAddress: bluetooth.MACAddress{MAC: mac}}
		_, e := adapter.Connect(a, bluetooth.ConnectionParams{})
		if e != nil {
			color.HiRed("connect fail: %s", e.Error())
		}

		return
	case "list": // list all connected devices
		mgr.mu.Lock()
		defer mgr.mu.Unlock()

		fmt.Println("Connected JoyCons:")
		for jc := range mgr.connected {
			fmt.Printf("<%s> %s %s\n", jc.Side().String(), jc.Mac(), renderBattery(jc.Battery()))
		}
		return
	case "calib": // calibrate stick
		for jc := range mgr.connected {
			jc.CalibrateStick()
		}
		return

	case "gyro": // toggle gyro printing
		gyro = !gyro
		for jc := range mgr.connected {
			jc.EnableGyro(gyro)
		}
		return

	case "rumble": // make JoyCon viberate for 32 frames.
		for jc := range mgr.connected {
			jc.Rumble(nil)
		}
		return
	case "test": // make JoyCon viberate for 32 frames.
		for jc := range mgr.connected {
			jc.Test()
		}
		return
	}
	color.HiRed("wrong command")
}

func main() {
	// need 1 thread per blocked cgo call
	runtime.GOMAXPROCS(8 + runtime.NumCPU())

	log.SetFormatter(&log.TextFormatter{
		ForceColors:      true,
		DisableTimestamp: true,
		TimestampFormat:  "04:05.000",
	})
	log.SetOutput(colorable.NewColorableStdout())

	// load config
	if e := loadConfig(); e != nil {
		log.Errorf("load config error: %s", e.Error())
		return
	}

	// watch config file change
	stopWatch := watchConfig()
	defer func() { stopWatch <- struct{}{} }()

	// command prompt
	color.HiBlue("press Ctrl-D to exit")
	chExit_Ctrl_d := make(chan struct{}, 1)

	go func() {
		p := prompt.New(
			executor,
			completer,
			prompt.OptionPrefix(">>> "),
		)
		p.Run()

		// It goes here after Ctrl-D is pressed.
		// Send an exit signal to the manager.
		chExit_Ctrl_d <- struct{}{}
	}()

	mgr = NewManager()
	mgr.monitorNewDevice(chExit_Ctrl_d)
}

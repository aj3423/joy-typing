package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/aj3423/joy-typing/joycon"
	"github.com/aj3423/joy-typing/mode"
	"github.com/gen2brain/beeep"
	log "github.com/sirupsen/logrus"
	"github.com/sstallion/go-hid"
)

type Manager struct {
	mu sync.Mutex

	connected map[joycon.Controller]joycon.RemoveListenerFn
}

func NewManager() *Manager {
	return &Manager{
		connected: make(map[joycon.Controller]joycon.RemoveListenerFn),
	}
}

func (m *Manager) monitorNewDevice(chExit_Ctrl_d chan struct{}) {

	// scan for newly connected device every second
	tickNewDevice := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-tickNewDevice.C:
			m.CheckNewDevice()

		case <-chExit_Ctrl_d: // ctrl-d in console, exit app
			m.removeAll()
			return
		}
	}
}

func (m *Manager) OnReadWriteError(
	jc joycon.Controller, err error,
) {
	log.Errorf("%s: %s", jc.Side().String(), err.Error())
	go beeep.Notify("R/W error", jc.Side().String(), "")

	m.mu.Lock()
	defer m.mu.Unlock()

	// If any r/w error occurs, the connection must be broken,
	// but the zombie connection still stays in system BT manager for a while,
	// so forcely disconnect it to avoid that period of time.
	go m.remove(jc, true)
}
func (m *Manager) OnButton(
	jc joycon.Controller,
	down, up, curr *joycon.ButtonState,
) {
	log.Tracef("onButton, down: %v, up: %v", down, up)

	mode.Manager.Handle(
		&mode.Input{
			Type: mode.InputType_Button,

			Jc: jc,
			ButtonInput: &mode.ButtonInput{
				Up: up, Down: down, Curr: curr,
			},
		},
	)
}

func isGoingNeutral(curr, prev *joycon.Ratio) bool {
	return !prev.AtNeutral() && curr.AtNeutral()
}
func isLeavingNeutral(curr, prev *joycon.Ratio) bool {
	return prev.AtNeutral() && !curr.AtNeutral()
}
func isGoingEdge(curr, prev *joycon.Ratio, dir joycon.SpinDirection) bool {
	return !prev.AtEdge(dir) && curr.AtEdge(dir)
}
func isLeavingEdge(curr, prev *joycon.Ratio, dir joycon.SpinDirection) bool {
	return prev.AtEdge(dir) && !curr.AtEdge(dir)
}

func (m *Manager) OnStick(
	jc joycon.Controller, side joycon.JoyConSide, curr, prev *joycon.Ratio,
) {
	go log.Tracef("onStick <%s>:  %s", jc.Side().String(), curr.String())

	// fire 2 events:
	// 1. a movement event
	// 2. a edge-movement event, e.g. Up, Upleave

	// 1.
	mode.Manager.Handle(
		&mode.Input{
			Type: mode.InputType_Stick,

			Jc: jc,

			StickInput: &mode.StickInput{
				Side:      side,
				Ratio:     curr,
				Direction: joycon.SpinDirection_None,
			},
		},
	)

	// 2.
	dir := joycon.SpinDirection_None

	if isGoingEdge(curr, prev, joycon.Spin_Up) { // spin Up
		dir = joycon.Spin_Up
	} else if isLeavingEdge(curr, prev, joycon.Spin_Up) { // leave Up area
		dir = joycon.Spin_Up_Leave
	} else if isGoingEdge(curr, prev, joycon.Spin_Right) {
		dir = joycon.Spin_Right
	} else if isLeavingEdge(curr, prev, joycon.Spin_Right) {
		dir = joycon.Spin_Right_Leave
	} else if isGoingEdge(curr, prev, joycon.Spin_Down) {
		dir = joycon.Spin_Down
	} else if isLeavingEdge(curr, prev, joycon.Spin_Down) {
		dir = joycon.Spin_Down_Leave
	} else if isGoingEdge(curr, prev, joycon.Spin_Left) {
		dir = joycon.Spin_Left
	} else if isLeavingEdge(curr, prev, joycon.Spin_Left) {
		dir = joycon.Spin_Left_Leave
	} else if isGoingNeutral(curr, prev) {
		dir = joycon.Spin_Neutral
	} else if isLeavingNeutral(curr, prev) {
		dir = joycon.Spin_Neutral_Leave
	}

	if dir != joycon.SpinDirection_None {
		mode.Manager.Handle(
			&mode.Input{
				Type: mode.InputType_Stick,

				Jc: jc,

				StickInput: &mode.StickInput{
					Side:      side,
					Ratio:     curr,
					Direction: dir,
				},
			},
		)
	}

}
func (m *Manager) OnStickCalib(
	jc joycon.Controller, calib *[2]joycon.CalibrationData,
) {
	log.Infof("🔧 <%s> Calibrated: %v", jc.Side().String(), calib)
	go beeep.Notify("🔧 Calibrated", jc.Side().String(), "")
}
func (m *Manager) OnGyro(jc joycon.Controller, gyro *joycon.GyroFrame) {
	log.Tracef("onGyro, %v", *gyro)
	mode.Manager.Handle(
		&mode.Input{
			Type: mode.InputType_Gyro,
			Gyro: &mode.Gyro{
				Frame: gyro,
			},
		},
	)
}
func renderBattery(level int8, charging bool) string {
	if charging {
		return fmt.Sprintf("🔋%d, ⚡ Charging", level)
	} else {
		return fmt.Sprintf("🔋%d", level)
	}
}
func (m *Manager) OnBattery(
	jc joycon.Controller, level int8, charging bool,
) {
	name := jc.Side().String()
	batt := renderBattery(level, charging)

	log.Infof("Battery: %s %s", name, batt)
	go beeep.Notify(name, batt, "")
}

func (m *Manager) remove(
	jc joycon.Controller, disconnectBT bool,
) {
	log.Warningf("Removing %s (%s) ...", jc.Side().String(), jc.Mac())

	// unbind events
	unbind := m.connected[jc]
	unbind()

	if disconnectBT {
		jc.ShutdownBT()
	}
	jc.Disconnect()
	delete(m.connected, jc)
}

func (m *Manager) removeAll() {
	log.Warningf("Removing all controllers...")
	m.mu.Lock()
	defer m.mu.Unlock()

	for jc := range m.connected {
		m.remove(jc, false)
	}
}

// find which side is not connected yet
func (m *Manager) findBySide(side joycon.JoyConSide) joycon.Controller {
	for jc := range m.connected {
		if jc.Side() == side {
			return jc
		}
	}
	return nil
}

// corresponding side <-> products
var sides = []joycon.JoyConSide{joycon.SideLeft, joycon.SideRight}
var products = []uint16{joycon.JOYCON_PRODUCT_L, joycon.JOYCON_PRODUCT_R}

func (m *Manager) CheckNewDevice() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, side := range sides {
		if m.findBySide(side) != nil { // already exists
			continue
		}

		dev, e := hid.OpenFirst(joycon.VENDOR_NINTENDO, products[i])
		if e != nil {
			continue
		}

		mac, e := dev.GetSerialNbr()
		if e != nil {
			dev.Close()
			continue
		}

		jc := joycon.NewJoycon(dev, side, mac)
		m.addNewDevice(jc)

	}

}
func (m *Manager) addNewDevice(jc joycon.Controller) {
	unbindFn := jc.SetListener(m)
	m.connected[jc] = unbindFn
	log.Infof("Connected to: <%s> %s", jc.Side(), jc.Mac())
	go beeep.Notify("Connected", jc.Side().String(), "")

	jc.Rumble(&joycon.RumbleFrequencySample)
}

package joycon

import (
	"fmt"
)

type ButtonState [3]byte

type ButtonID uint16

// First byte of ButtonState.
const (
	Button_R_Y ButtonID = 0x000 + (1 << iota)
	Button_R_X
	Button_R_B
	Button_R_A
	Button_R_SR
	Button_R_SL
	Button_R_R
	Button_R_ZR
)

// Middle byte of ButtonState.
const (
	Button_Minus ButtonID = 0x100 + (1 << iota)
	Button_Plus
	Button_R_Stick
	Button_L_Stick
	Button_Home
	Button_Capture
	Button_Unused1
	Button_IsChargeGrip
)

// Last byte of ButtonState.
const (
	Button_L_Down ButtonID = 0x200 + (1 << iota)
	Button_L_Up
	Button_L_Right
	Button_L_Left
	Button_L_SR
	Button_L_SL
	Button_L_L
	Button_L_ZL
)

var buttonNameMap = map[ButtonID]string{
	Button_R_Y:          "Y",
	Button_R_X:          "X",
	Button_R_B:          "B",
	Button_R_A:          "A",
	Button_R_SR:         "R-SR",
	Button_R_SL:         "R-SL",
	Button_R_R:          "R",
	Button_R_ZR:         "ZR",
	Button_Minus:        "-",
	Button_Plus:         "+",
	Button_R_Stick:      "RStick",
	Button_L_Stick:      "LStick",
	Button_Home:         "Home",
	Button_Capture:      "Capture",
	Button_Unused1:      "Unused1",
	Button_IsChargeGrip: "ChargingGrip",
	Button_L_Down:       "Down",
	Button_L_Up:         "Up",
	Button_L_Right:      "Right",
	Button_L_Left:       "Left",
	Button_L_SR:         "L-SR",
	Button_L_SL:         "L-SL",
	Button_L_L:          "L",
	Button_L_ZL:         "ZL",
}
var buttonNameVMap = map[string]ButtonID{
	"Y":            Button_R_Y,
	"X":            Button_R_X,
	"B":            Button_R_B,
	"A":            Button_R_A,
	"R-SR":         Button_R_SR,
	"R-SL":         Button_R_SL,
	"R":            Button_R_R,
	"ZR":           Button_R_ZR,
	"-":            Button_Minus,
	"+":            Button_Plus,
	"RStick":       Button_R_Stick,
	"LStick":       Button_L_Stick,
	"Home":         Button_Home,
	"Capture":      Button_Capture,
	"Unused1":      Button_Unused1,
	"ChargingGrip": Button_IsChargeGrip,
	"Down":         Button_L_Down,
	"Up":           Button_L_Up,
	"Right":        Button_L_Right,
	"Left":         Button_L_Left,
	"L-SR":         Button_L_SR,
	"L-SL":         Button_L_SL,
	"L":            Button_L_L,
	"ZL":           Button_L_ZL,
}

func (b ButtonID) String() string {
	name, ok := buttonNameMap[b]
	if !ok {
		panic(fmt.Sprintf("no button with id: %d", b))
	}
	return name
}

func ButtonFromString(s string) (ButtonID, bool) {
	id, ok := buttonNameVMap[s]
	return id, ok
}

// ButtonsFromSlice copies the provided slice from a standard input report into
// a ButtonState.
func ButtonsFromSlice(b []byte) ButtonState {
	var result ButtonState
	result[0] = b[0]
	result[1] = b[1]
	result[2] = b[2]
	return result
}

// Has the state of a single ButtonID.
func (b ButtonState) Has(i ButtonID) bool {
	return b[(i&0x0300)>>8]&byte(i&0xFF) != 0
}

// DownMask returns buttons that being pressed down
func (b ButtonState) DownMask(other ButtonState) ButtonState {
	var result ButtonState
	result[0] = ^b[0] & other[0]
	result[1] = ^b[1] & other[1]
	result[2] = ^b[2] & other[2]
	return result
}

// DownMask returns buttons that being released up
func (b ButtonState) UpMask(other ButtonState) ButtonState {
	var result ButtonState
	result[0] = b[0] & ^other[0]
	result[1] = b[1] & ^other[1]
	result[2] = b[2] & ^other[2]
	return result
}

// IsZero returns true if all key is zero
func (b ButtonState) IsZero() bool {
	return b[0] == 0 && b[1] == 0 && b[2] == 0
}

package mode

import (
	"github.com/aj3423/joy-typing/joycon"
)

type InputType int

const (
	InputType_None = iota // for always-on-trigger
	InputType_Button
	InputType_Stick
	InputType_Gyro
	InputType_Speech
)

type ButtonInput struct {
	Up, Down, Curr *joycon.ButtonState
}
type StickInput struct {
	Side      joycon.JoyConSide
	Ratio     *joycon.Ratio
	Direction joycon.SpinDirection
}
type SpeechInput struct {
	Text string
}
type Gyro struct {
	Frame *joycon.GyroFrame
}

type Input struct {
	Type InputType

	Jc joycon.Controller

	// button
	*ButtonInput

	// stick
	*StickInput

	// gyro
	*Gyro

	// text
	*SpeechInput
}

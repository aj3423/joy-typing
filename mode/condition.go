package mode

import (
	"github.com/aj3423/joy_typing/joycon"
)

// For checking if an *Input matches the requirement
type condition interface {
	Satisfy(*Input) bool
}

// If the *Input matches Up/Down event of specified key
type ButtonCondition struct {
	// if 'true', it responds to button-down event
	// and 'false' for button-up event
	whenDown bool

	// the target button
	btnId joycon.ButtonID
}

func (bc *ButtonCondition) Satisfy(in *Input) bool {
	if in.Type != InputType_Button {
		return false
	}
	if bc.whenDown && in.Down.Has(bc.btnId) {
		return true
	}
	if !bc.whenDown && in.Up.Has(bc.btnId) {
		return true
	}
	return false
}

// check if there is specified stick movement event
type StickMoveCondition struct {
	side joycon.JoyConSide
}

func (sc *StickMoveCondition) Satisfy(in *Input) bool {
	return in.Type == InputType_Stick &&
		in.Side == sc.side &&
		in.Direction == joycon.SpinDirection_None
}

// 5 enter and 5 leave events for: U D L R Center
type StickDirectionCondition struct {
	side joycon.JoyConSide
	dir  joycon.SpinDirection
}

func (sc *StickDirectionCondition) Satisfy(in *Input) bool {
	return in.Type == InputType_Stick &&
		in.Side == sc.side &&
		in.Direction == sc.dir
}

// return true if there is speech signal
type SpeechCondition struct{}

func (sc *SpeechCondition) Satisfy(in *Input) bool {
	return in.Type == InputType_Speech && len(in.Text) > 0
}

// return true if there is gyro signal
type GyroCondition struct{}

func (sc *GyroCondition) Satisfy(in *Input) bool {
	return in.Type == InputType_Gyro
}

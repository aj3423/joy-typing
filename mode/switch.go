package mode

import (
	"github.com/aj3423/joy-typing/joycon"
)

type SwitchResult int

const (
	SwitchedOn SwitchResult = iota
	SwitchedOff
	SwitchNotChange
)

// A switch keeps tracking of On/Off state,
// On/Off actions will be triggered when state changes.
type switch_ interface {
	IsOn() bool
	Reset()

	GetOnTrigger() trigger
	SetOnTrigger(trigger)
	GetOffTrigger() trigger
	SetOffTrigger(trigger)

	// If the `Input` is expected then
	// the switch is turned on/off.
	// And corresponding action triggered
	Handle(*Input) SwitchResult
}

// A convenient class for default members
type Switch struct {
	isOn bool

	onTrigger  trigger
	offTrigger trigger
}

func (s *Switch) IsOn() bool              { return s.isOn }
func (s *Switch) Reset()                  { s.isOn = false }
func (s *Switch) GetOnTrigger() trigger   { return s.onTrigger }
func (s *Switch) SetOnTrigger(t trigger)  { s.onTrigger = t }
func (s *Switch) GetOffTrigger() trigger  { return s.offTrigger }
func (s *Switch) SetOffTrigger(t trigger) { s.offTrigger = t }

func (s *Switch) Handle(
	in *Input,
) SwitchResult {
	if s.onTrigger.Handle(in) == Triggered {
		s.isOn = true
		return SwitchedOn
	}
	if s.offTrigger.Handle(in) == Triggered {
		s.isOn = false
		return SwitchedOff
	}
	return SwitchNotChange
}

// A switch that is turned on On/Off by Button down/up
type ButtonSwitch struct {
	Switch
}

func NewButtonSwitch(btnId joycon.ButtonID) *ButtonSwitch {
	bs := &ButtonSwitch{}
	bs.SetOnTrigger(NewButtonTrigger(btnId, true, nil))
	bs.SetOffTrigger(NewButtonTrigger(btnId, false, nil))
	return bs
}

// A switch that is turned on On/Off by spinning stick to the specified direction
type StickDirectionSwitch struct {
	Switch

	side joycon.JoyConSide
	dir  joycon.SpinDirection
}

func NewStickDirectionSwitch(side joycon.JoyConSide, dir joycon.SpinDirection) *StickDirectionSwitch {
	ss := &StickDirectionSwitch{}
	ss.SetOnTrigger(NewStickDirectionTrigger(side, dir, nil))
	// if On trigger is "Spin Up", reversed trigger is "Spin Up Leave"
	reversed := joycon.ReverseDirectionMap[dir]
	ss.SetOffTrigger(NewStickDirectionTrigger(side, reversed, nil))
	return ss
}

// Use two words to turn the switch on/off
// type VoiceSwitch struct {}

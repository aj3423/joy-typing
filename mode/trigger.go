package mode

import (
	"github.com/aj3423/joy_typing/joycon"
)

type TriggerResult int

const (
	Triggered TriggerResult = iota
	NotTriggered
)

// a trigger is made of a condition and an action
// if the *Input event satisfies the condition, the action will be called
type trigger interface {
	GetCondition() condition
	SetCondition(condition)

	GetAction() action
	SetAction(action)

	// If the `Input` satisfies the `condition` then the trigger is pulled
	// Returns whether it's pulled.
	Handle(*Input) TriggerResult
}

// A convenient class for default members
type Trigger struct {
	condition
	action
}

func (t *Trigger) GetCondition() condition  { return t.condition }
func (t *Trigger) SetCondition(c condition) { t.condition = c }
func (t *Trigger) GetAction() action        { return t.action }
func (t *Trigger) SetAction(a action)       { t.action = a }

func (t *Trigger) Handle(in *Input) TriggerResult {
	if t.Satisfy(in) {
		if t.action != nil {
			t.Do(in)
		}
		return Triggered
	}
	return NotTriggered
}

type ButtonTrigger struct {
	Trigger
}

func NewButtonTrigger(
	btnId joycon.ButtonID, whenDown bool, a action,
) *ButtonTrigger {
	b := &ButtonTrigger{}

	b.condition = &ButtonCondition{
		whenDown: whenDown, btnId: btnId,
	}
	b.action = a
	return b
}

type StickMoveTrigger struct {
	Trigger
}

func NewStickMoveTrigger(
	side joycon.JoyConSide, a action,
) *StickMoveTrigger {
	t := &StickMoveTrigger{}

	t.condition = &StickMoveCondition{
		side: side,
	}
	t.action = a
	return t
}

type StickDirectionTrigger struct {
	Trigger
}

func NewStickDirectionTrigger(
	side joycon.JoyConSide, dir joycon.SpinDirection, a action,
) *StickDirectionTrigger {
	t := &StickDirectionTrigger{}

	t.condition = &StickDirectionCondition{
		side: side,
		dir:  dir,
	}
	t.action = a
	return t
}

type SpeechTrigger struct {
	Trigger
}

func NewSpeechTrigger(
	a action,
) *SpeechTrigger {
	t := &SpeechTrigger{}

	t.condition = &SpeechCondition{}
	t.action = a
	return t
}

type GyroTrigger struct {
	Trigger
}

func NewGyroTrigger(
	a action,
) *GyroTrigger {
	t := &GyroTrigger{}

	t.condition = &GyroCondition{}
	t.action = a
	return t
}

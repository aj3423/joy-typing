package mode

import (
	"fmt"
	"sync"
)

// global variable
var Manager = &ModeManager{}

type ModeManager struct {
	mu sync.RWMutex

	// all modes indexed by mode.Id
	map_ map[string]mode

	// switches to handle the mode switching,
	// these switches are called before `mode.Handle()`
	modeSwitches []switch_
	// A trigger for exit current mode
	// It is set to the exit trigger when entering a new mode
	// e.g. switching from A->B by button-ZR-down, this is set to button-ZR-up
	trigExit trigger

	// the default mode, it's the first one in config
	defaultMode mode

	currentMode mode
}

// Set modes from configuration file,
// the first is set as default mode
func (l *ModeManager) SetModes(list []mode, modeSwitches []switch_) error {
	l.mu.RLock()
	defer l.mu.RUnlock()

	l.modeSwitches = modeSwitches

	l.map_ = make(map[string]mode)
	for _, m := range list {
		if _, exist := l.map_[m.Id()]; exist {
			return fmt.Errorf("duplicated mode id: %s", m.Id())
		}
		l.map_[m.Id()] = m
	}
	// first as default
	l.defaultMode = list[0]
	return l.switchTo(l.defaultMode.Id(), nil)
}

func (l *ModeManager) Handle(in *Input) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// check if it's mode switching
	if l.currentMode == l.defaultMode {
		// check if it's mode entering
		for _, swch := range l.modeSwitches {
			if swch.Handle(in) == SwitchedOn {
				// mode changed in the above Handle(), save the exit trigger
				l.trigExit = swch.GetOffTrigger()
				return // stop if mode switched
			}
		}
	} else { // mode has been switched, check if the Input triggers exit
		if l.trigExit.Handle(in) == Triggered {
			l.trigExit = nil
			return
		}
	}

	// The *Input isn't related to the mode switching,
	// pass it to the concrete mode
	l.currentMode.Handle(in)
}

func (l *ModeManager) switchTo(id string, in *Input) error {
	m, ok := l.map_[id]
	if !ok {
		return fmt.Errorf("mode '%s' not exists", id)
	}

	if l.currentMode != nil {
		if e := l.currentMode.OnExit(in); e != nil {
			return e
		}
	}
	l.currentMode = m
	return l.currentMode.OnEnter(in)
}

func (l *ModeManager) DefaultMode() mode {
	return l.defaultMode
}
func (l *ModeManager) CurrentMode() mode {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.currentMode
}

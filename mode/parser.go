package mode

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aj3423/joy-typing/joycon"
	"github.com/alexflint/go-arg"
	"github.com/mattn/go-shellwords"
)

func parseArg(grammar interface{}, args []string) error {
	a, _ := arg.NewParser(arg.Config{}, grammar)
	return a.Parse(args)
}

var ModeList []ModeConfig
var PhraseList map[string][]string
var WordMapping map[string][]string

type ModeConfig struct {
	Mode  string   ``
	Rules []string `toml:"Rules,multiline"`
}

func splitArrowLine(line string) (left, right []string, e error) {
	lr := strings.Split(line, "->")
	if len(lr) != 2 {
		return nil, nil, fmt.Errorf("missing '->' in rule line: %s", line)
	}

	left, e = shellwords.Parse(lr[0])
	if e != nil {
		return nil, nil, fmt.Errorf("left part error: %s", lr[0])
	}
	right, e = shellwords.Parse(lr[1])
	if e != nil {
		return nil, nil, fmt.Errorf("right part error: %s", lr[1])
	}

	return
}

func Parse() ([]mode, []switch_, error) {
	if len(ModeList) == 0 {
		return nil, nil, errors.New("no mode rules configured")
	}

	// all modes
	retModes := []mode{}
	// all hotkeys for switching mode
	retModeSwitches := []switch_{}

	for modeIndex, modeBlock := range ModeList {
		var m mode
		var actions []trigger                     // actions of above mode
		var switches = make(map[switch_]modifier) // modifiers of above mode

		// 1. parse mode
		Types, e := shellwords.Parse(modeBlock.Mode)
		if e != nil {
			return nil, nil, fmt.Errorf("wrong mode '%s': %s", modeBlock.Mode, e.Error())
		}

		if len(Types) < 1 {
			return nil, nil, fmt.Errorf("missing type for mode[%d]", modeIndex)
		}
		m, e = parseMode(Types[0], Types[1:])
		if e != nil {
			return nil, nil, fmt.Errorf("wrong mode: %s", e.Error())
		}

		// 2. parse rules
		for _, line := range modeBlock.Rules {
			lefts, rights, e := splitArrowLine(line)
			if e != nil {
				return nil, nil, fmt.Errorf("wrong rule format: %s: %s", line, e.Error())
			}
			if len(lefts) < 2 || len(rights) == 0 {
				return nil, nil, fmt.Errorf("wrong rule: %s", line)
			}

			switch lefts[0] {

			case `[trigger]`: // trigger -> action
				t, e := parseTrigger(lefts[1], lefts[2:])
				if e != nil {
					return nil, nil, fmt.Errorf("wrong trigger: %s, %s", line, e.Error())
				}
				a, e := parseAction(rights[0], rights[1:])
				if e != nil {
					return nil, nil, fmt.Errorf("wrong action: %s, %s", line, e.Error())
				}
				t.SetAction(a) // bind action to trigger
				actions = append(actions, t)

			case `[switch]`: // switch -> modifier/mode
				s, e := parseSwitch(lefts[1], lefts[2:])
				if e != nil {
					return nil, nil, fmt.Errorf("wrong switch: %s, %s", line, e.Error())
				}
				// the Value part of json can be either a modifier or an action
				// try modifier first
				modi, e := parseModifier(rights[0], rights[1:])
				if e == nil {
					switches[s] = modi
				} else { // not modifier, try action
					switch rights[0] {
					case `[mode]`:

						if modeIndex != 0 {
							return nil, nil, fmt.Errorf("mode switch can only be defined in default mode(the first in the list):  %s", line)
						}
						a, e := parseAction(rights[0], rights[1:])
						if e != nil {
							return nil, nil, fmt.Errorf("wrong mode action: %s, %s", line, e.Error())
						}
						s.GetOnTrigger().SetAction(a)
						s.GetOffTrigger().SetAction(&RestoreMode{})
						retModeSwitches = append(retModeSwitches, s)
					case `[gyro]`:

						s.GetOnTrigger().SetAction(&EnableGyro{true})
						s.GetOffTrigger().SetAction(&EnableGyro{false})
						switches[s] = nil // no need modifier for this switch

					case `[mouse_toggle]`:
						grammar := &struct {
							Button string
						}{Button: "left"}
						e := parseArg(grammar, rights[1:])

						if e != nil {
							return nil, nil, e
						}

						s.GetOnTrigger().SetAction(NewMouseDown(grammar.Button))
						s.GetOffTrigger().SetAction(NewMouseUp(grammar.Button))
						switches[s] = nil // no need modifier for this switch

					default:
						return nil, nil, fmt.Errorf("unknown switch: %s", rights[0])
					}
				}
			default:
				return nil, nil, fmt.Errorf("unknown key '%s', should begin with either 'trigger'/'switch'", lefts[0])
			}
		}

		m.SetSwitches(switches)
		m.SetActions(actions)

		retModes = append(retModes, m)
	}

	return retModes, retModeSwitches, nil
}
func parseTrigger(name string, args []string) (trigger, error) {
	lname := strings.ToLower(name)

	switch lname {
	case `button`:
		grammar := &struct {
			Id       string `arg:"required"`
			WhenDown bool
		}{WhenDown: true}

		e := parseArg(grammar, args)
		if e != nil {
			return nil, e
		}

		btnId, ok := joycon.ButtonFromString(grammar.Id)

		if !ok {
			return nil, fmt.Errorf("no button named: %s", grammar.Id)
		}
		return NewButtonTrigger(btnId, grammar.WhenDown, nil), nil

	case `stick`:
		grammar := &struct {
			Side string `arg:"required"`
			Dir  string
		}{}
		e := parseArg(grammar, args)

		if e != nil {
			return nil, fmt.Errorf("wrong 'stick' args: %s", e.Error())
		}
		side, valid := joycon.SideMap[grammar.Side]
		if !valid {
			return nil, fmt.Errorf("unsupported JoyCon side: %s", grammar.Side)
		}

		direction, exist := joycon.SpinDirectionMap[grammar.Dir]
		if exist {
			return NewStickDirectionTrigger(side, direction, nil), nil
		} else {
			return NewStickMoveTrigger(side, nil), nil
		}
	case `gyro`:
		return NewGyroTrigger(nil), nil
	case `speech`:
		return NewSpeechTrigger(nil), nil
	default:
		return nil, fmt.Errorf("no trigger named: %s", name)
	}
}
func parseSwitch(
	name string, args []string,
) (switch_, error) {
	lname := strings.ToLower(name)

	switch lname {
	case `button`:
		grammar := &struct {
			Id string `arg:"required"`
		}{}

		e := parseArg(grammar, args)
		if e != nil {
			return nil, e
		}

		btnId, ok := joycon.ButtonFromString(grammar.Id)

		if !ok {
			return nil, fmt.Errorf("no button named: %s", grammar.Id)
		}
		return NewButtonSwitch(btnId), nil
	case `stick`:
		grammar := &struct {
			Side string `arg:"required"`
			Dir  string `arg:"required"`
		}{}

		e := parseArg(grammar, args)
		if e != nil {
			return nil, e
		}

		side, valid := joycon.SideMap[grammar.Side]
		if !valid {
			return nil, fmt.Errorf("invalid side: %s", grammar.Side)
		}

		direction, valid := joycon.SpinDirectionMap[grammar.Dir]
		if !valid {
			return nil, fmt.Errorf("invalid direction: %s", grammar.Dir)
		}
		return NewStickDirectionSwitch(side, direction), nil
	default:
		return nil, fmt.Errorf("no switch named: %s", name)
	}
}
func parseModifier(
	name string, args []string,
) (modifier, error) {
	lname := strings.ToLower(name)

	switch lname {

	case `[upper]`:
		return &TextPrefix{prefix: Tag_CaseUpper, space: true}, nil
	case `[camel]`:
		return &TextPrefix{prefix: Tag_CaseCamel, space: true}, nil
	case `[title]`:
		return &TextPrefix{prefix: Tag_CaseTitle, space: true}, nil
	case `[snake]`:
		return &TextPrefix{prefix: Tag_CaseSnake, space: true}, nil
	case `[prefix]`:
		grammar := &struct {
			Prefix string `arg:"required"`
			Space  bool
		}{Space: true}
		e := parseArg(grammar, args)
		return &TextPrefix{prefix: grammar.Prefix, space: grammar.Space}, e

	case `[boost]`:
		grammar := &struct {
			Multiplier float64 `arg:"required"`
		}{}
		e := parseArg(grammar, args)
		return NewCursorBoost(grammar.Multiplier), e

	default:
		return nil, fmt.Errorf("no modifier named: %s", name)
	}
}

func parseAction(name string, args []string) (action, error) {
	lname := strings.ToLower(name)

	switch lname {

	case `[cursor]`:
		grammar := &struct {
			Speed float64 ``
		}{Speed: 0.01}
		e := parseArg(grammar, args)
		return NewMoveCursor(grammar.Speed), e
	case `[click]`:
		grammar := &struct {
			Button string
			Double bool
		}{Button: "left"}
		e := parseArg(grammar, args)
		return NewMouseClick(grammar.Button, grammar.Double), e
	case `[hotkey]`:
		grammar := &struct {
			Keys []string `arg:"required"`
		}{}
		e := parseArg(grammar, args)
		return NewHotkey(grammar.Keys), e

	case `[notify]`:
		grammar := &struct {
			Title, Text, Icon string
		}{}
		e := parseArg(grammar, args)
		return NewSysNotify(grammar.Title, grammar.Text, grammar.Icon), e

	case `[speak]`:
		grammar := &struct {
			Text string
		}{}
		e := parseArg(grammar, args)
		return &Speak{grammar.Text}, e
	case `[speech]`:
		grammar := &struct {
			Number, NoSpace, Typing bool
			Map                     []string
		}{Number: true, NoSpace: true, Typing: true}
		e := parseArg(grammar, args)
		if e != nil {
			return nil, e
		}
		return NewExecSpeech(grammar.Number, grammar.NoSpace, grammar.Typing, grammar.Map)

	case `[flush]`:
		return &FlushVoice{}, nil
	case `[repeat]`:
		return &Repeat{}, nil

	case `[mode]`:
		grammar := &struct {
			Id string `arg:"required"`
		}{}
		e := parseArg(grammar, args)
		return NewSwitchMode(grammar.Id), e

	default:
		return nil, fmt.Errorf("unknown action: %s", name)
	}
}
func parseMode(name string, args []string) (mode, error) {
	lname := strings.ToLower(name)

	switch lname {

	case `[idle]`:
		grammar := &struct {
			Id string `arg:"required"`
		}{}
		e := parseArg(grammar, args)
		return NewIdleMode(grammar.Id), e

	case `[gyro]`:
		grammar := &struct {
			Id string `arg:"required"`
		}{}
		e := parseArg(grammar, args)
		return NewGyroMode(grammar.Id), e

	case `[speech]`:
		grammar := &struct {
			Id          string `arg:"required"`
			Host        string
			Phrase      []string
			Engine      string
			FlushOnExit bool
		}{Engine: "vosk", Host: "localhost:2701"}
		e := parseArg(grammar, args)
		if e != nil {
			return nil, e
		}
		return NewSpeechMode(grammar.Id, grammar.Engine, grammar.Host, grammar.Phrase, grammar.FlushOnExit)

	default:
		return nil, fmt.Errorf("no mode named: '%s'", name)
	}
}

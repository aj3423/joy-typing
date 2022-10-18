package mode

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/go-vgo/robotgo"
	log "github.com/sirupsen/logrus"
)

// An `action` does something according to the Input
type action interface {
	// Take action agaist input,
	//   result decorated by modifiers
	Do(*Input)
}

// ---- actions ----

var muMouse sync.Mutex

// Moving mouse cursor takes time, should be don in goroutine,
// but calling `robotgo.MoveRelative` simutainously causes crash, so use a lock.
func robotgo_MoveRelative(x, y int) {
	muMouse.Lock()
	robotgo.MoveRelative(x, y)
	muMouse.Unlock()
}
func robotgo_Click(type_ string, isDouble bool) {
	muMouse.Lock()
	robotgo.Click(type_, isDouble)
	muMouse.Unlock()
}
func robotgo_Toggle(type_, downUp string) {
	muMouse.Lock()
	robotgo.Toggle(type_, downUp)
	muMouse.Unlock()
}

// Cursor movement
// JoyCon's packet push inerval is 15ms
// `robotgo.MoveRelative` takes < 2ms
type MoveCursor struct {

	// a factor that can increase/decrease the cursor speed
	speed float64
}

func NewMoveCursor(speed float64) *MoveCursor {
	return &MoveCursor{speed: speed}
}
func (mc *MoveCursor) Do(in *Input) {
	switch in.Type {
	case InputType_Stick:
		go robotgo_MoveRelative(
			int(in.Ratio.X*mc.speed),
			int(-in.Ratio.Y*mc.speed), // it's vertical inverted without the '-'
		)

	case InputType_Gyro:
		go robotgo_MoveRelative(
			int(float64(in.Frame.Yaw)*mc.speed),
			int(-float64(in.Frame.Pitch)*mc.speed),
		)
	}
}

type MouseClick struct {
	// available button type:
	// "left", "center", "right", "wheelDown", "wheelUp", "wheelLeft", "wheelRight"
	button   string
	isDouble bool // is double click
}

func NewMouseClick(button string, isDouble bool) *MouseClick {
	return &MouseClick{button: button, isDouble: isDouble}
}
func (mc *MouseClick) Do(*Input) {
	robotgo_Click(mc.button, mc.isDouble)
}

// Mode Switcher
type SwitchMode struct {
	modeId string
}

func NewSwitchMode(modeId string) *SwitchMode {
	return &SwitchMode{modeId: modeId}
}

func (sm *SwitchMode) Do(in *Input) {
	e := Manager.switchTo(sm.modeId, in)
	if e != nil {
		go beeep.Alert("failed to switch to mode "+sm.modeId, e.Error(), "")
	}
}

// Restore to default mode
type RestoreMode struct{}

func (rm *RestoreMode) Do(in *Input) {
	Manager.switchTo(Manager.defaultMode.Id(), in)
}

// Popup a system notification with specified Title/Text/Icon
type SysNotify struct {
	title, text, icon string
}

func NewSysNotify(title, text, icon string) *SysNotify {
	return &SysNotify{title, text, icon}
}

func (sn *SysNotify) Do(in *Input) {
	switch in.Type {
	case InputType_Speech: // just for testing
		go beeep.Notify("speech", in.Text, sn.icon)
	default:
		go beeep.Notify(sn.title, sn.text, sn.icon)
	}
}

// `FlushVoice` forces speech engine stop waiting for
// further data and return the result immediately
type FlushVoice struct{}

func (fv *FlushVoice) Do(*Input) {
	switch sp := Manager.currentMode.(type) {
	case *SpeechMode:
		go sp.recEngine.Flush()
	}
}

// `Speak` simulate a speech, throw it to the word executor,
// normally used for complex macro that cannot be done with a simple action.
// For example, if we want to:
// win+R -> delay-1-second -> type"calc.exe" -> enter -> delay-1-second -> type"1+1="
// this can only be done with word executors, hence a [speech] trigger must be set in that mode.
type Speak struct {
	text string
}

func (s *Speak) Do(*Input) {
	go Manager.Handle(&Input{
		Type: InputType_Speech,
		SpeechInput: &SpeechInput{
			Text: s.text,
		},
	})

}

type EnableGyro struct {
	enable bool
}

func (eg *EnableGyro) Do(in *Input) {
	go in.Jc.EnableGyro(eg.enable)
}

type MouseToggle struct {
	button string
	downUp string
}

func NewMouseToggle(button, downUp string) *MouseToggle {
	return &MouseToggle{button, downUp}
}
func (md *MouseToggle) Do(*Input) {
	go robotgo_Toggle(md.button, md.downUp)
}
func NewMouseDown(button string) *MouseToggle {
	return NewMouseToggle(button, "down")
}
func NewMouseUp(button string) *MouseToggle {
	return NewMouseToggle(button, "up")
}

// just use the `hotkey` of `word executor`
type Hotkey struct {
	hotkey
}

func (h *Hotkey) Do(*Input) {
	h.exec(nil)
}
func NewHotkey(keys []string) *Hotkey {
	h := &Hotkey{}
	h.keys = keys
	return h
}

// used for Repeat
var lastSpeech string

type Repeat struct {
}

func (r *Repeat) Do(*Input) {
	if len(lastSpeech) > 0 {
		go Manager.Handle(&Input{
			Type: InputType_Speech,
			SpeechInput: &SpeechInput{
				Text: lastSpeech,
			},
		})
	}
}

type ExecSpeech struct {
	// whether cast number words to digital format, e.g. "twenty twenty two" -> "2022"
	castNumber bool
	// whether remove space between words, for programming mode
	noSpace bool
	// whether simulate keyboard stroke to type the words
	typing bool

	// words replacement, configured in the section `WordMapping`,
	// if a replacement are configured without any keyword prefix,
	// it is considered as a replacement
	// e.g. `space -> " "`,
	mappingTree *node[*replace]

	// for text executing, in the section `WordMapping`,
	// these are configured with a function keyword prefix like `[shell]`
	// e.g. `launch browser -> [shell] "brave-browser" "www.test.com"`
	execTree *fallbackNode[executorFactory]
}

func NewExecSpeech(
	castNumber, noSpace, typing bool,
	mappings []string,
) (*ExecSpeech, error) {
	es := &ExecSpeech{
		castNumber:  castNumber,
		noSpace:     noSpace,
		mappingTree: newNode[*replace](),
		execTree:    newFallbackNode[executorFactory](),
	}
	if typing {
		es.execTree.SetFallback(&typingFactory{noSpace: noSpace})
	}

	for _, mapId := range mappings {

		mapp, ok := WordMapping[mapId]
		if !ok {
			return nil, fmt.Errorf("mapping id not exist in 'WordMapping' section: %s", mapId)
		}
		for _, line := range mapp {

			lefts, rights, e := splitArrowLine(line)

			if e != nil || len(lefts) == 0 || len(rights) == 0 {
				return nil, fmt.Errorf("wrong format: %s", line)
			}

			switch rights[0] {
			case `[shell]`:
				if len(rights) < 2 {
					return nil, fmt.Errorf("wrong '[shell]': %s, missing command list", line)
				}
				es.execTree.Set(lefts, &shellFactory{cmd: rights[1:]})
			case `[hotkey]`:
				if len(rights) == 1 {
					es.execTree.Set(lefts, &hotkeyDynFactory{})
				} else {
					es.execTree.Set(lefts, &hotkeyFixFactory{keys: rights[1:]})
				}
			case `[delay]`:
				switch len(rights) {
				case 1:
					es.execTree.Set(lefts, &delayDynGenerator{})
				case 2:
					dur, e := time.ParseDuration(rights[1])
					if e != nil {
						return nil, fmt.Errorf("wrong '[delay] duration': %s", line)
					}
					es.execTree.Set(lefts, &delayFixGenerator{duration: dur})
				default:
					return nil, fmt.Errorf("wrong '[delay]': %s", line)
				}
			default: // it's word replace
				es.mappingTree.Set(lefts, &replace{to: rights})
			}
		}
	}
	// always set these case mapping
	// if ignorecase
	es.execTree.Set([]string{Tag_CaseCamel}, &wordDecoratorFactory{fn: toCamel})
	es.execTree.Set([]string{Tag_CaseUpper}, &wordDecoratorFactory{fn: toUpper})
	es.execTree.Set([]string{Tag_CaseSnake}, &wordDecoratorFactory{fn: toSnake})
	es.execTree.Set([]string{Tag_CaseTitle}, &wordDecoratorFactory{fn: toTitle})
	es.execTree.Set([]string{"[repeat]"}, &repeatFactory{})
	// more to add

	return es, nil
}
func containsRepeat(words wordArray) bool {
	for _, w := range words {
		if w == Tag_Repeat {
			return true
		}
	}
	return false
}

func (es *ExecSpeech) Do(in *Input) {

	log.Info("💬 ", in.Text)
	var words wordArray = strings.Split(in.Text, ` `)

	// 1. cast numbers, e.g.
	// "hello two thousand n three hundred and twelve world eleven"
	// -> "hello 2312 world 11"
	if es.castNumber {
		words = replaceNumbers(words)
	}

	// 2. replace words that configured in section `WordMapping`
	// only do text replacement
	// `newWords` holds the new mapped words
	newWords := wordArray{}
	for !words.empty() {
		scanCost, replacer, isMapped := es.mappingTree.Scan(words)
		if isMapped { // replace this word
			replacer.exec(&newWords) // this appends replaced word to `newWords`
			words.skip(scanCost)     // skip the words being scanned
		} else { // don't replace this word if it's not mapped
			newWords.add1(words[0])
			words.skip(1)
		}
	}

	// save text for [repeat]
	if !containsRepeat(newWords) { // not being with `[repeat]`
		lastSpeech = in.Text
	}

	// 3. generate executor array from word array
	// e.g. ["camel", "hello", "world", "hotkey", "control", "a"]
	// -> [textModify(camel), typing("hello"), typing("world"), hotkey(ctrl+a)]
	var executors []executor

	words = newWords

	for !words.empty() {
		scanCost, factory, isMapped := es.execTree.Scan(words)
		if isMapped { // executor found for current word(s)
			words.skip(scanCost)

			cost, ex, e := factory.parse(words)
			if e == nil {

				// executor `typing` need to handle words together
				// for example: "[camel] hello world" -> "helloWorld"
				// not possible to get the result "helloWorld" if these words are executed separatly
				// so group the `typing`s together, if the word is not a space char
				{
					if len(executors) > 0 {
						last := executors[len(executors)-1]
						// if previous one is `typing`
						if exPrev, prevIsTyping := last.(*typing); prevIsTyping && !exPrev.isSpace() {
							// if current is also `typing`
							if exCurr, currIsTyping := ex.(*typing); currIsTyping && !exCurr.isSpace() {
								exPrev.group(exCurr)
								words.skip(1)
								// not append to `executors` because it's combined to the previous one
								continue
							}
						}
					}
				}

				words.skip(cost)
				executors = append(executors, ex)
			}
		} else {
			// if no executor handles this word, skip this word
			words.skip(1)
		}
	}

	// 4. run all executors
	modifiers := []*wordDecorator{}

	for _, g := range executors {
		switch ex := g.(type) {
		case *wordDecorator: // hardcoded...
			modifiers = append(modifiers, ex)
		case *typing:
			ex.decorators = modifiers
			modifiers = nil
			g.exec(nil)
		default:
			g.exec(nil)
		}
	}
}

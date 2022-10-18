package mode

import (
	"fmt"

	"github.com/aj3423/joy-typing/voice"
)

// mode it self does nothing,
// it's a container of modifier switches and action triggers
type mode interface {
	Id() string

	// initialize/reset self and underlying triggers/switches
	// param *Input indicates which event triggers this switching
	OnEnter(*Input) error
	OnExit(*Input) error

	Handle(*Input)

	SetSwitches(map[switch_]modifier)
	SetActions([]trigger)
}

// `Mode` is a container of modifier switches and action triggers,
// it passes all Input events to them
type Mode struct {
	id string

	switches map[switch_]modifier

	actions []trigger // handlers to *Input event
}

func (m *Mode) OnEnter(*Input) error {
	// `` and `modifier`s have on/off state
	// should be reset to off
	for swch, modi := range m.switches {
		swch.Reset()
		if modi != nil {
			modi.Reset()
		}
	}
	return nil
}
func (m *Mode) OnExit(*Input) error { return nil }

func (m *Mode) Id() string { return m.id }

func (m *Mode) SetActions(t []trigger)              { m.actions = t }
func (m *Mode) SetSwitches(sw map[switch_]modifier) { m.switches = sw }

func (m *Mode) Handle(in *Input) {
	// Turn modifier switches on/off
	for swch, modi := range m.switches {
		// Some switches do action when triggered,
		// some only change the state `isOn`.
		swch.Handle(in)

		// If the modifier has state 'On',
		// apply it to *Input
		// for example: modifier uppercase changes `input.Text` to UPPER CASE
		if swch.IsOn() && modi != nil {
			modi.Modify(in)
		}
	}

	// trigger actions
	for _, trig := range m.actions {
		trig.Handle(in)
	}
}

// A mode that does nothing, used as default mode
type IdleMode struct {
	Mode
}

func NewIdleMode(modeId string) *IdleMode {
	r := &IdleMode{}
	r.id = modeId
	return r
}

type GyroMode struct {
	Mode
}

func NewGyroMode(modeId string) *GyroMode {
	g := &GyroMode{}
	g.id = modeId
	return g
}

func (g *GyroMode) OnEnter(in *Input) error {
	go in.Jc.EnableGyro(true)
	return g.Mode.OnEnter(in)
}
func (g *GyroMode) OnExit(in *Input) error {
	go in.Jc.EnableGyro(false)
	return g.Mode.OnExit(nil)
}

type SpeechMode struct {
	Mode

	host string

	// flush voice on exit, works well with limited phrase_list,
	// but not work with full phrase_list
	flushOnExit bool

	recEngine   voice.RecognitionEngine
	phrase_list []string // vosk config 'phrase_list'

	paused bool // only send audio data to engine when not paused
}

func NewSpeechMode(
	modeId, engine, host string,
	phraseIds []string, // e.g. ["common", "go", "lua"]
	flushOnExit bool,
) (*SpeechMode, error) {

	sp := &SpeechMode{}
	sp.id = modeId
	sp.host = host
	sp.paused = true
	sp.flushOnExit = flushOnExit

	// 1. recognition engine
	switch engine {
	case "vosk":
		sp.recEngine = &voice.Vosk{}
	default:
		return nil, fmt.Errorf("unknown speech engine: %s", engine)
	}

	// 2. parse phrase files
	sp.phrase_list = []string{}
	for _, phId := range phraseIds { // parse all word files
		ph, exist := PhraseList[phId]
		if !exist {
			return nil, fmt.Errorf("phrase '%s' not exist", phId)
		}

		sp.phrase_list = append(sp.phrase_list, ph...)
	}

	return sp, nil
}

// Start monitoring microphone input and send captured voice data to recognition engine.
func (sp *SpeechMode) OnEnter(*Input) error {
	if !sp.recEngine.IsAlive() {
		sp.recEngine.Close() // cleanup

		// handle speech recognition result from engine
		sp.recEngine.SetCallback(func(result string) {
			if len(result) > 0 {
				Manager.Handle(&Input{
					Type: InputType_Speech,
					SpeechInput: &SpeechInput{
						Text: result,
					},
				})
			}
		})

		// make websocket connection
		if e := sp.recEngine.Dial(sp.host); e != nil {
			return e
		}
		// make this websocket only recognize these words
		if e := sp.recEngine.SetPhraseList(sp.phrase_list); e != nil {
			return e
		}
	}

	// start capturing from audio device, the `StartCapture` is singleton
	// it'll never be shut down once it's started, because the startup is expensive
	// on my machine it takes 300+ ms, can't do that frequently
	if e := voice.AudioDevice.StartCapture(); e != nil {
		return e
	}
	// redirect all miniaudio voice data to recognition engine through the websocket
	voice.AudioDevice.SetCallback(func(data []byte) {
		if !sp.paused {
			sp.recEngine.SendBinary(data)
		}
	})
	sp.paused = false

	return sp.Mode.OnEnter(nil)
}
func (sp *SpeechMode) OnExit(*Input) error {
	if sp.flushOnExit {
		sp.recEngine.Flush()
	}
	sp.paused = true
	return sp.Mode.OnExit(nil)
}

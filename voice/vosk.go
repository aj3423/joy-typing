package voice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
)

type Vosk struct {
	mu sync.Mutex
	ws *websocket.Conn

	isAlive  bool
	callback func(string)
}

type voskResult struct {
	Text string `json:"text"`
}

func (v *Vosk) SetCallback(cb func(string)) {
	v.callback = cb
}
func (v *Vosk) Dial(host string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	var e error
	u := url.URL{Scheme: "ws", Host: host, Path: ""}
	v.ws, _, e = websocket.DefaultDialer.DialContext(
		context.Background(), u.String(), nil)
	if e != nil {
		return e
	}

	go v.read()

	return nil
}
func (v *Vosk) read() {
	v.isAlive = true
	defer func() { v.isAlive = false }()

	for {
		_, msg, e := v.ws.ReadMessage()
		if e != nil {
			v.Close()
			break
		}
		r := &voskResult{}
		if e := json.Unmarshal(msg, r); e == nil {
			if v.callback != nil {
				v.callback(r.Text)
			}
		}
	}
}
func (v *Vosk) IsAlive() bool { return v.isAlive }

func (v *Vosk) Close() error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.ws == nil {
		return nil
	}

	return v.ws.Close()
}

// The difference between SpeechModes is 'phrase_list'
// Set custom 'phrase_list' when entering new SpeechMode
func (v *Vosk) SetPhraseList(phrase []string) error {
	s, e := json.Marshal(phrase) // convert to string: `["a","b",..."z"]`
	if e != nil {
		return e
	}
	return v.SendText(fmt.Sprintf(`
		{
			"config" : {
				"sample_rate" : 16000,
				"phrase_list" : %s,
				"words": 0
			}
		}`, string(s)))
}

func (v *Vosk) send(type_ int, data []byte) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.ws == nil {
		return errors.New("not connected")
	}

	return v.ws.WriteMessage(type_, data)
}
func (v *Vosk) SendText(text string) error {
	return v.send(websocket.TextMessage, []byte(text))
}
func (v *Vosk) SendBinary(data []byte) error {
	return v.send(websocket.BinaryMessage, data)
}

// Speech server like VOSK waits for further data after stopped speaking
// Normall it takes 500ms ~ several seconds to return the result.
// I don't find any 'official' way of telling it not wait and return result immediately,
// So force it stop waiting by sending a chunk of 40k zeroes
var FlushData = make([]byte, 40000)

func (v *Vosk) Flush() error {
	return v.send(websocket.BinaryMessage, FlushData)
}

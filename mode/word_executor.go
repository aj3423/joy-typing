package mode

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/iancoleman/strcase"
)

type wordArray []string

// add a word array
func (w *wordArray) add(words []string) {
	*w = append(*w, words...)
}

// add 1 single word
func (w *wordArray) add1(word string) {
	w.add([]string{word})
}
func (w *wordArray) skip(n int) {
	*w = (*w)[n:]
}
func (w *wordArray) size() int {
	return len(*w)
}
func (w *wordArray) empty() bool {
	return w.size() == 0
}
func (w *wordArray) clear() *wordArray {
	*w = nil
	return w
}

// word executor unit
type executor interface {
	// the param `*wordContext` is only used for data output, such as: `replace`
	exec(out *wordArray) error
}

type executorFactory interface {
	// dynamic parse input words and generate `executor`
	// eg 1: if factory is dynamic `delay`
	//  words == ["1", "second", "haha"]
	// `parse` comsumes 2 words ["1", "second"] and returns (2, *delay, nil)
	// eg 2: factory == typing
	//  words == ["aa", "bb"]
	// `parse` comsumes the first 1 word and returns (1, *typing, nil)
	parse(words []string) (wordCost int, exec executor, e error)
}

// simply sleep, format: "delay number second(s)|millisecond(s)"
type delay struct {
	duration time.Duration
}

func (d *delay) exec(_ *wordArray) error {
	time.Sleep(d.duration)
	return nil
}

type delayFixGenerator struct {
	duration time.Duration
}

func (dg *delayFixGenerator) parse(_ []string) (cost int, exec executor, e error) {
	return 0, &delay{duration: dg.duration}, nil
}

// simply sleep, format: "delayDynamic number second(s)|millisecond(s)"
type delayDynGenerator struct{}

func (d *delayDynGenerator) parse(words []string) (cost int, exec executor, e error) {
	words = replaceNumbers(words)

	if len(words) < 3 {
		return 0, nil, errors.New(`'delay' requires at least 3 words, like: "delay three second|millisecond"`)
	}
	number, e := strconv.Atoi(words[1])
	if e != nil {
		return 0, nil, e
	}
	switch words[2] {
	case `second`, `seconds`:
		return 3, &delay{duration: time.Duration(number) * time.Second}, nil
	case `millisecond`, `milliseconds`:
		return 3, &delay{duration: time.Duration(number) * time.Millisecond}, nil
	default:
		return 0, nil, fmt.Errorf("unknown sleep unit: %s", words[2])
	}
}

// e.g. config "launch chrome" -> ["brave-browser", "--no-sandbox", "www.test.com"]
// and say "launch chrome"
type shell struct {
	cmd []string
}

func (s *shell) exec(_ *wordArray) error {
	cm := exec.Command(s.cmd[0], s.cmd[1:]...)
	_, e := cm.CombinedOutput()
	return e
}

// factory for generting fixed shell, which are configured in file
type shellFactory struct {
	cmd []string
}

func (fac *shellFactory) parse(_ []string) (cost int, exec executor, e error) {
	return 0, &shell{cmd: fac.cmd}, nil
}

// config "launch terminal" -> ["t", "alt", "control"]
// and say "launch terminal"
type hotkey struct {
	keys []string // e.g. ["t", "alt", "control"]
}

func (h *hotkey) fixKeys() {
	for i, key := range h.keys {
		if key == "control" {
			h.keys[i] = "ctrl"
		}
	}
}
func (h *hotkey) exec(_ *wordArray) error {
	h.fixKeys() // change "control" -> "ctrl", which is used by robotgo
	// lastPos := len(h.keys) - 1
	// last := h.keys[lastPos]
	// switch last {
	// case `control`, `alt`, `meta`, `shift`:
	return robotgo.KeyTap(h.keys[0], h.keys[1:])
	// default:
	// 	return robotgo.KeyTap(last, h.keys[0:lastPos])
	// }
}

// factory for generting fixed hotkey, which are configured in file
type hotkeyFixFactory struct {
	keys []string
}

func (fac *hotkeyFixFactory) parse(_ []string) (cost int, exec executor, e error) {
	return 0, &hotkey{keys: fac.keys}, nil
}

// factory for generting dynamic hotkey action by
// speech starting with word "hotkey" and followed by key combinations
// like: "hotkey" "control" "alt" "t"
type hotkeyDynFactory struct{}

func (fac *hotkeyDynFactory) parse(words []string) (cost int, exec executor, e error) {
	words = replaceNumbers(words)

	if len(words) == 0 {
		return 0, nil, errors.New(`'hotkey' requires at least 1 words, like: "hotkey A"`)
	}

	// TODO fix this, implement multiple keys
	keys := []string{}

	switch words[0] {
	case "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
		"delete":
		return 1, &hotkey{keys: keys}, nil
	}
	return 0, nil, fmt.Errorf("unknown hotkey: %s", words[0])
}

type replace struct {
	to []string // target words are saved in context
}

// append target words to the context `out`
func (r *replace) exec(out *wordArray) error {
	out.add(r.to)
	return nil
}

// typing out text directly
// used as a fallback action when words not being recognized as hotkey/shell_command
type typing struct {
	noSpace bool

	decorators []*wordDecorator
	words      wordArray // target words to type
}

// the member `words` should be groupped when `exec()`
func (t *typing) exec(_ *wordArray) error {
	if len(t.words) > 0 {
		for _, dec := range t.decorators {
			dec.exec(&t.words)
		}
		if t.noSpace {
			robotgo.TypeStr(strings.Join(t.words, ``))
		} else {
			robotgo.TypeStr(strings.Join(t.words, ` `))
		}
	}
	return nil
}

func (t *typing) group(other *typing) {
	t.words.add(other.words)
}
func (t *typing) isSpace() bool {
	return t.words.size() == 1 && t.words[0] == " "
}

type typingFactory struct {
	noSpace bool
}

func (fac *typingFactory) parse(words []string) (cost int, exec executor, e error) {
	t := &typing{
		noSpace: fac.noSpace,
	}
	t.words.add1(words[0]) // add 1 word
	return 1, t, nil       // return 1
}

type repeat struct{}

// the member `words` should be groupped when `exec()`
func (r *repeat) exec(_ *wordArray) error {
	// todo, dynamic parse like `repeat 5`
	go Manager.Handle(&Input{
		Type: InputType_Speech,
		SpeechInput: &SpeechInput{
			Text: lastSpeech,
		},
	})
	return nil
}

type repeatFactory struct{}

func (fac *repeatFactory) parse(words []string) (cost int, exec executor, e error) {
	t := &repeat{}
	return 0, t, nil // return 1
}

const (
	Tag_Repeat    = "[repeat]"
	Tag_CaseCamel = "[camel]"
	Tag_CaseTitle = "[title]"
	Tag_CaseUpper = "[upper]"
	Tag_CaseSnake = "[snake]"
)

// decorate words to other case
type wordDecorator struct {
	fn func(*wordArray) error
}

func (t *wordDecorator) exec(outWords *wordArray) error {
	return t.fn(outWords)
}

type wordDecoratorFactory struct {
	fn func(*wordArray) error
}

func (fac *wordDecoratorFactory) parse(_ []string) (cost int, exec executor, e error) {
	t := &wordDecorator{fn: fac.fn}
	return 0, t, nil
}

func toCamel(wa *wordArray) error {
	s := strcase.ToLowerCamel(strings.Join(*wa, " "))
	wa.clear().add1(s)
	return nil
}
func toUpper(wa *wordArray) error {
	for i, w := range *wa {
		(*wa)[i] = strings.ToUpper(w)
	}
	return nil
}
func toTitle(wa *wordArray) error {
	if wa.size() > 0 {
		(*wa)[0] = strings.Title((*wa)[0])
	}
	return nil
}
func toSnake(wa *wordArray) error {
	s := strcase.ToSnake(strings.Join(*wa, " "))
	wa.clear().add1(s)
	return nil
}

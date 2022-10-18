package voice

type RecognitionEngine interface {
	Dial(host string) error
	Close() error

	SetCallback(func(string)) // callback will be invoked when got result from engine

	IsAlive() bool // websocket connection is alive

	SendText(text string) error
	SendBinary(data []byte) error
	SetPhraseList(phraseList []string) error
	Flush() error
}

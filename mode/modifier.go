package mode

type modifier interface {
	Modify(*Input)
	Reset()
}

// A convenient class for default members
type Modifier struct {
}

func (m *Modifier) Reset() {}

// ---- all modifiers ----

type TextPrefix struct {
	Modifier
	prefix string // a string prepended to speech text
	space  bool   // add a " " after "prefix", default: true
}

func (tp *TextPrefix) Modify(in *Input) {
	if in.Type == InputType_Speech {
		pre := tp.prefix
		if tp.space {
			pre += " "
		}
		in.Text = pre + in.Text
	}
}

// Boost cursor move speed or slow it down
type CursorBoost struct {
	Modifier

	multiplier float64
}

func NewCursorBoost(multiplier float64) *CursorBoost {
	return &CursorBoost{multiplier: multiplier}
}

func (cb *CursorBoost) Modify(in *Input) {
	switch in.Type {

	case InputType_Stick:
		in.Ratio.X *= cb.multiplier
		in.Ratio.Y *= cb.multiplier
	case InputType_Gyro:
		in.Frame.X = int16(float64(in.Frame.X) * cb.multiplier)
		in.Frame.Y = int16(float64(in.Frame.Y) * cb.multiplier)
		in.Frame.Z = int16(float64(in.Frame.Z) * cb.multiplier)
		in.Frame.Roll = int16(float64(in.Frame.Roll) * cb.multiplier)
		in.Frame.Yaw = int16(float64(in.Frame.Yaw) * cb.multiplier)
		in.Frame.Pitch = int16(float64(in.Frame.Pitch) * cb.multiplier)
	}
}

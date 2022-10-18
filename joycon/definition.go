package joycon

const (
	VENDOR_NINTENDO           uint16 = 0x057e
	JOYCON_PRODUCT_L          uint16 = 0x2006
	JOYCON_PRODUCT_R          uint16 = 0x2007
	JOYCON_PRODUCT_FAKE       uint16 = 0x2008
	JOYCON_PRODUCT_PRO        uint16 = 0x2009
	JOYCON_PRODUCT_CHARGEGRIP uint16 = 0x200e
)

type JoyConSide int

const (
	SideInvalid JoyConSide = iota
	SideLeft               = 1
	SideRight              = 2
	SideBoth               = 3
)

func (s JoyConSide) IsLeft() bool {
	return s == SideLeft || s == SideBoth
}

func (s JoyConSide) IsRight() bool {
	return s == SideRight || s == SideBoth
}

func (s JoyConSide) String() string {
	switch s {
	case SideLeft:
		return "Joy-Con L"
	case SideRight:
		return "Joy-Con R"
	case SideBoth:
		return "Switch Pro Controller"
	}
	return "Unknown Device"
}

var SideMap = map[string]JoyConSide{
	"Left":  SideLeft,
	"Right": SideRight,
}

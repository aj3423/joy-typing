package joycon

import (
	"fmt"
	"math"
)

var (
	// Movement event only fired if > this tolerance
	SpinNeutralThreshold float64 = 0.08
	// Rotation event only fired if > this tolerance
	SpinEdgeThreshhold float64 = 0.70
)

type SpinDirection int

const (
	SpinDirection_None SpinDirection = iota
	Spin_Neutral
	Spin_Neutral_Leave
	Spin_Up
	Spin_Up_Leave
	Spin_Right
	Spin_Right_Leave
	Spin_Down
	Spin_Down_Leave
	Spin_Left
	Spin_Left_Leave
)

var SpinDirectionMap = map[string]SpinDirection{
	"Up":           Spin_Up,
	"UpLeave":      Spin_Up_Leave,
	"Right":        Spin_Right,
	"RightLeave":   Spin_Right_Leave,
	"Down":         Spin_Down,
	"DownLeave":    Spin_Down_Leave,
	"Left":         Spin_Left,
	"LeftLeave":    Spin_Left_Leave,
	"Neutral":      Spin_Neutral,
	"NeutralLeave": Spin_Neutral_Leave,
}
var ReverseDirectionMap = map[SpinDirection]SpinDirection{
	Spin_Up:    Spin_Up_Leave,
	Spin_Right: Spin_Right_Leave,
	Spin_Down:  Spin_Down_Leave,
	Spin_Left:  Spin_Left_Leave,
}

type Axis2D[T uint16 | float64] struct {
	X, Y T
}
type Point Axis2D[uint16]
type Ratio Axis2D[float64]

var NeutralRatio = Ratio{}

func (r *Ratio) String() string {
	return fmt.Sprintf("[%f  %f]", r.X, r.Y)
}
func (r *Ratio) AtNeutral() bool {
	return math.Abs(r.X) <= SpinNeutralThreshold && math.Abs(r.Y) <= SpinNeutralThreshold
}

// if the Ratio is at edge
// meaning it exceeds the `SpinEdgeThreshhold` in that direction
func (r *Ratio) AtEdge(dir SpinDirection) bool {
	switch dir {
	case Spin_Up:
		return r.Y >= SpinEdgeThreshhold
	case Spin_Down:
		return -r.Y >= SpinEdgeThreshhold
	case Spin_Left:
		return -r.X >= SpinEdgeThreshhold
	case Spin_Right:
		return r.X >= SpinEdgeThreshhold
	}
	return false
}

func decodeUint12(b []byte) (uint16, uint16) {
	d1 := uint16(b[0]) | (uint16(b[1]&0xF) << 8)
	d2 := uint16(b[1]>>4) | (uint16(b[2]) << 4)
	return d1, d2
}

type CalibrationData struct {
	xMinOff, xCenter, xMaxOff uint16 // x
	yMinOff, yCenter, yMaxOff uint16 // y
}

var EmptyCalibrationData = CalibrationData{}

// side must be TypeLeft or TypeRight; TypeBoth controllers should call this twice
func (c *CalibrationData) Parse(b []byte, side JoyConSide) {
	if side == SideLeft {
		c.xMaxOff, c.yMaxOff = decodeUint12(b[0:3])
		c.xCenter, c.yCenter = decodeUint12(b[3:6])
		c.xMinOff, c.yMinOff = decodeUint12(b[6:9])
	} else {
		c.xCenter, c.yCenter = decodeUint12(b[0:3])
		c.xMinOff, c.yMinOff = decodeUint12(b[3:6])
		c.xMaxOff, c.yMaxOff = decodeUint12(b[6:9])
	}
}

// Transform raw stick values into -1.0~1.0 float64.
// 0 == center, -1.0 == most left, 1.0 == most right
func (c *CalibrationData) Adjust(rawXY *Point) (ret Ratio) {
	// cast to int() to avoid overflow

	xOffset := int(rawXY.X) - int(c.xCenter)
	if xOffset > 0 {
		ret.X = float64(xOffset) / float64(c.xMaxOff)
	} else {
		ret.X = float64(xOffset) / float64(c.xMinOff)
	}

	yOffset := int(rawXY.Y) - int(c.yCenter)
	if yOffset > 0 {
		ret.Y = float64(yOffset) / float64(c.yMaxOff)
	} else {
		ret.Y = float64(yOffset) / float64(c.yMinOff)
	}
	return
}

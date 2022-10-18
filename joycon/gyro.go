package joycon

type Gyro3D struct {
	X, Y, Z int16
}

func (g *Gyro3D) Adjust(endPos *Gyro3D) Gyro3D {
	return Gyro3D{
		X: endPos.X - g.X,
		Y: endPos.Y - g.Y,
		Z: endPos.Z - g.Z,
	}
}

type Acceleration struct {
	Roll, Pitch, Yaw int16
}

// from: https://github.com/dekuNukem/Nintendo_Switch_Reverse_Engineering/blob/master/imu_sensor_notes.md
// The 6-Axis data is repeated 3 times. On Joy-con with a 15ms packet push,
// this is translated to 5ms difference sampling.
// E.g. 1st sample 0ms, 2nd 5ms, 3rd 10ms.
// Using all 3 samples let you have a 5ms precision instead of 15ms.
type GyroFrame struct {
	Gyro3D       // absolute value
	Acceleration // reletive value
}

var GyroFrame_Nil GyroFrame

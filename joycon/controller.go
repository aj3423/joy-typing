package joycon

type Controller interface {
	Mac() string
	Side() JoyConSide

	Disconnect()
	ShutdownBT() error

	// Bind events to listener,
	// Call the returned function to unbind.
	SetListener(EventListener) RemoveListenerFn

	Battery() (level int8, charging bool) // 4=full, 3, 2, 1=critical, 0=empty
	EnableGyro(isOn bool) error
	Rumble(*RumbleFrequency) error
	SetLights(pattern byte)

	CalibrateStick() error

	Test()
}

type RemoveListenerFn func()

type EventListener interface {
	// fail read/write to hid, BT connection broken
	OnReadWriteError(Controller, error)
	// button down/up
	OnButton(jc Controller, down, up, curr *ButtonState)
	// stick spinning around
	OnStick(jc Controller, t JoyConSide, curr, prev *Ratio)
	// stick calibrated successfylly
	OnStickCalib(Controller, *[2]CalibrationData)
	// gyro motion, todo
	OnGyro(Controller, *GyroFrame)
	// battery level change
	OnBattery(jc Controller, level int8, charging bool)
}

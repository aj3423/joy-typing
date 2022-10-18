package joycon

// NOTE
// ---- FBI WARNING ----
// Don't use real maximum values for Amplitude. Otherwise, they can damage the linear actuators.
// Ref: https://github.com/dekuNukem/Nintendo_Switch_Reverse_Engineering/blob/master/bluetooth_hid_notes.md#rumble-data
type RumbleFrequency [8]byte

// (320Hz 0.0f 160Hz 0.0f) is neutral, it doesn't vibrate.
var RumbleFrequencyNeutral RumbleFrequency = [8]byte{0, 1, 0x40, 0x40, 0, 1, 0x40, 0x40}

// a frequency sample that vibrates
var RumbleFrequencySample RumbleFrequency = [8]byte{0, 4, 0x1, 0xfc, 0, 4, 0x01, 0xfc}

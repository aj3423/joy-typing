package joycon

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/sstallion/go-hid"
)

const (
	// When Joycon is in this mode, packets are pushed from Joycon at 60Hz
	// for all button/stick/gyro events.
	// Linux enters this mode automatically when BT connected, Windows doesn't, see below.
	StandardFull byte = 0x30

	// Default mode on Windows, need to switch to StandardFull manually
	// it only push packet on button press, no stick/gyro events.
	SimpleHid byte = 0x3F
)

type joycon struct {
	hidDev *hid.Device

	packetId byte // // Increment by 1 for each packet sent. It loops in 0x0 - 0xF range.

	side JoyConSide

	listener EventListener

	mac string

	mu sync.RWMutex

	battery     uint8       // 1 byte contains battery and charging state
	prevButtons ButtonState // button state for previous frame
	currButtons ButtonState // button state for current frame

	prevStick  [2]Ratio           // [left, right][x, y]
	currStick  [2]Ratio           // [left, right][x, y]
	stickCalib [2]CalibrationData // [left, right]

	gyroOn    bool
	gyroBegin GyroFrame // the initial state when start rotating
	gyro      [3]GyroFrame
}

func NewJoycon(
	hidDev *hid.Device,
	side JoyConSide,
	mac string,
) Controller {
	jc := &joycon{
		hidDev: hidDev,
		side:   side,
		mac:    mac,
	}

	go jc.readLoop()

	go func() {
		// switch runtime.GOOS {
		// case "linux": // do nothing, linux auto enters StandardFullMode, no idea why.
		// }
		// Sometimes it never get the response if the calibrating packet is sent too quick(right after attached)
		// It may even cause the CPU goes to 100% and system freezes to death.
		// Don't call `SPIRead()` too early, even a 500ms delay is too short.
		// 2 seconds seems ok so far, never freezes yet and always calibrates successfully.
		time.Sleep(2 * time.Second)

		jc.enterMode(StandardFull)

		time.Sleep(200 * time.Millisecond)
		for try := 0; try < 3; try++ {
			if jc.isCalibrated() {
				break
			}
			jc.CalibrateStick()
			time.Sleep(time.Second)
		}
	}()

	return jc
}
func (jc *joycon) CalibrateStick() error {
	return jc.SPIRead(factoryStickCalibStart, factoryStickCalibLen)
	// time.Sleep(100 * time.Millisecond)
	// jc.SPIRead(userStickCalibStart, userStickCalibLen)
}

func (jc *joycon) SetListener(lsn EventListener) RemoveListenerFn {
	jc.mu.Lock()
	defer jc.mu.Unlock()

	jc.listener = lsn

	return func() {
		jc.mu.Lock()
		defer jc.mu.Unlock()

		jc.listener = nil
	}
}
func (jc *joycon) Mac() string {
	return jc.mac
}

func (jc *joycon) Side() JoyConSide {
	return jc.side
}

// Battery level and charging state.
func (jc *joycon) Battery() (int8, bool) {
	return int8(jc.battery >> 5), jc.battery&0x10 != 0
}

func (jc *joycon) enterMode(mode byte) error {
	sub := []byte{0x03, mode}

	return jc.sendSubcommand(sub, nil)
}
func (jc *joycon) EnableGyro(enable bool) error {
	sub := []byte{0x40, 0}
	if enable {
		sub[1] = 1
	}

	jc.gyroBegin = GyroFrame_Nil
	jc.gyroOn = enable

	return jc.sendSubcommand(sub, nil)
}

/*
* see: https://github.com/dekuNukem/Nintendo_Switch_Reverse_Engineering/blob/master/bluetooth_hid_subcommands_notes.md#subcommand-0x30-set-player-lights
*   bits of the `pattern` byte:
*   aaaa bbbb
*        3210 - keep light on
*   3210 - flash light
* it works fine on Windows, but it only work for 1 frame on Linux even with "keep light on"
 */
func (jc *joycon) SetLights(pattern byte) {
	sub := []byte{0x30, byte(pattern)}
	jc.sendSubcommand(sub, nil)
}

// Perform an SPI read
func (jc *joycon) SPIRead(addr uint32, len byte) error {
	sub := []byte{0x10, 0, 0, 0, 0, len}

	binary.LittleEndian.PutUint32(sub[1:], addr)

	return jc.sendSubcommand(sub, nil)
}

func (jc *joycon) Disconnect() {
	jc.mu.Lock()
	defer jc.mu.Unlock()

	if jc.hidDev != nil {
		jc.hidDev.Close()
		jc.hidDev = nil
	}
}
func (jc *joycon) Test() {
	jc.enterMode(StandardFull)
}
func (jc *joycon) ShutdownBT() error {
	sub := []byte{
		6, // sets HCI state
	}
	return jc.sendSubcommand(sub, nil)
}

// send output report
func (jc *joycon) sendSubcommand(
	sub []byte,
	rumble *RumbleFrequency,
) error {
	jc.mu.Lock()
	defer jc.mu.Unlock()

	hidHandle := jc.hidDev

	if hidHandle == nil {
		return errors.New("hid handle closed")
	}
	jc.packetId++
	if jc.packetId == 16 { // loop between 0~15
		jc.packetId = 0
	}

	packet := make([]byte, 10)
	packet[0] = 0x1 // outputReportID = 0x01
	packet[1] = jc.packetId
	if rumble != nil {
		copy(packet[2:], (*rumble)[:]) // 2~10 rumble data, zeroes by default
	}
	packet = append(packet, sub...)

	_, e := hidHandle.Write(packet)
	return e
}

// Rumble seems not work on Windows, but works fine on Linux
func (jc *joycon) Rumble(freq *RumbleFrequency) error {
	sub := []byte{
		0x48, 0x01,
	}
	return jc.sendSubcommand(sub, freq)

}
func (jc *joycon) decodeBattery(packet []byte) {
	prevBattery := jc.battery
	jc.battery = packet[2] & 0xF0 // battery is high nibble of this byte
	if prevBattery != jc.battery {
		lvl, charging := jc.Battery()
		jc.listener.OnBattery(jc, lvl, charging)
	}
}

func (jc *joycon) decodeButton(packet []byte) {
	jc.prevButtons = jc.currButtons
	jc.currButtons = ButtonsFromSlice(packet[3:6])

	down := jc.prevButtons.DownMask(jc.currButtons) // all key down
	up := jc.prevButtons.UpMask(jc.currButtons)     // all key up

	// only trigger event if there is change
	if !down.IsZero() || !up.IsZero() {
		jc.listener.OnButton(jc, &down, &up, &jc.currButtons)
	}
}

func (jc *joycon) decodeStick(packet []byte) {
	if jc.isCalibrated() { // stick
		jc.prevStick = jc.currStick
		p := &Point{}

		if jc.side.IsLeft() {
			p.X, p.Y = decodeUint12(packet[6:9])

			jc.currStick[0] = jc.stickCalib[0].Adjust(p)

			// don't fire event if it stays at neutral position
			if !jc.currStick[0].AtNeutral() || !jc.prevStick[0].AtNeutral() {
				jc.listener.OnStick(jc, SideLeft, &jc.currStick[0], &jc.prevStick[0])
			}
		}
		if jc.side.IsRight() {
			p.X, p.Y = decodeUint12(packet[9:12])
			jc.currStick[1] = jc.stickCalib[1].Adjust(p)

			if !jc.currStick[1].AtNeutral() || !jc.prevStick[1].AtNeutral() {
				jc.listener.OnStick(jc, SideRight, &jc.currStick[1], &jc.prevStick[1])
			}
		}
	}
}

func (jc *joycon) decodePacket(packet []byte) {
	jc.mu.Lock()
	defer jc.mu.Unlock()

	jc.decodeBattery(packet)
	jc.decodeButton(packet)
	jc.decodeStick(packet)
}

// the stick data is only useful when it's calibrated.
func (jc *joycon) isCalibrated() bool {
	if jc.side.IsLeft() {
		return jc.stickCalib[0] != EmptyCalibrationData
	} else if jc.side.IsRight() {
		return jc.stickCalib[1] != EmptyCalibrationData
	}
	return false
}

func gyroPrint(f *GyroFrame) {
	fmt.Printf("  %7d %7d %7d %7d %7d %7d\n",
		f.X, f.Y, f.Z, f.Roll, f.Pitch, f.Yaw)
}

func (jc *joycon) decodeGyroData(packet []byte) {
	jc.mu.Lock()
	defer jc.mu.Unlock()

	if !jc.gyroOn {
		return
	}

	// from: https://github.com/dekuNukem/Nintendo_Switch_Reverse_Engineering/blob/master/imu_sensor_notes.md
	// The 6-Axis data is repeated 3 times. On Joy-con with a 15ms packet push,
	// this is translated to 5ms difference sampling.
	// E.g. 1st sample 0ms, 2nd 5ms, 3rd 10ms.
	// Using all 3 samples let you have a 5ms precision instead of 15ms.
	for i := 0; i < 3; i++ {
		jc.gyro[i].X = int16(binary.LittleEndian.Uint16(packet[13+2*(i*6+0):]))
		jc.gyro[i].Y = int16(binary.LittleEndian.Uint16(packet[13+2*(i*6+1):]))
		jc.gyro[i].Z = int16(binary.LittleEndian.Uint16(packet[13+2*(i*6+2):]))
		jc.gyro[i].Roll = int16(binary.LittleEndian.Uint16(packet[13+2*(i*6+3):]))
		jc.gyro[i].Pitch = int16(binary.LittleEndian.Uint16(packet[13+2*(i*6+4):]))
		jc.gyro[i].Yaw = int16(binary.LittleEndian.Uint16(packet[13+2*(i*6+5):]))
	}

	// fmt.Println("====data====")
	// gyroPrint(&jc.gyro[0])
	// gyroPrint(&jc.gyro[1])
	// gyroPrint(&jc.gyro[2])
	/*
		sample data
		  X		  Y		Z			roll  pitch     yaw
		  338     -16   -4050       0     -10       3
		  337     -17   -4041      -2     -11       4
		  333     -20   -4039      -7     -10       4
	*/

	// early packets(<200ms) may contain all zeroes, skip them
	if jc.gyro[0] == GyroFrame_Nil {
		return
	}

	// save the start value
	if jc.gyroBegin == GyroFrame_Nil {
		jc.gyroBegin = jc.gyro[0]
	}

	// use the first frame, adjust it with `gyroBegin` to calculate the offset
	// and fire event
	adj := jc.gyro[0]
	adj.Gyro3D = adj.Adjust(&jc.gyroBegin.Gyro3D)

	jc.listener.OnGyro(jc, &adj)
}

func (jc *joycon) handleSubcommandReply(packet []byte) {
	packetID := packet[0]

	replyPacketID := byte(0)
	if packetID == 0x21 {
		replyPacketID = packet[13] - 0x80
	} else /* 0x31-0x33 */ {
		replyPacketID = packet[13]
	}

	switch replyPacketID {
	case 0: // If it is a simple ACK, the byte13 is x80 and thus the type of data is x00
	case 2: // only once, after the BT connection is made.
	case 0x10: // SPI Flash Read
		jc.handleSPIRead(packet[13:])
	default:
		log.Warningf("got subcommand reply packet: %d\n%s", replyPacketID, hex.Dump(packet[:]))
	}
}

func (jc *joycon) readLoop() {
	var buffer [0x200]byte // windows max packet size: 0x16a

	for {
		n, e := jc.hidDev.Read(buffer[:]) // blocking
		if e != nil {
			jc.listener.OnReadWriteError(jc, e)
			return
		}
		if n >= len(buffer) {
			panic("need larger buffer size")
		}

		packet := buffer[:n]
		if len(packet) == 0 {
			continue
		}

		// fmt.Printf("packet0: %x, n: %x\n", packet[0], n)
		// if len(packet) > 0x100 {
		// 	fmt.Println(hex.Dump(packet))
		// }

		switch packet[0] {

		// Standard input reports used for subcommand replies.
		// most packets of 0x21 are light blinking on/off, 8 packets per second == 4 lights * 2(on/off state change).
		case 0x21:
			jc.decodePacket(packet)
			jc.handleSubcommandReply(packet) // handle SPIRead response

		// all input reports with IMU data instead of subcommand replies.
		case StandardFull: // 0x30
			jc.decodePacket(packet)
			jc.decodeGyroData(packet)

		// button events are pushed as 0x3f packet if not switched to StandardFullMode,
		// size == 0x16a with lots of trailing zeroes, seems bug of hidapi, one issue of it claims this was fixed but actually it's not
		// the buffer must be at least 0x16a on Windows
		case SimpleHid: //0x3F
			// do nothing, should be switched to StandardFullMode when attached
			// maybe still have a few packets, just ignore

		default:
			log.Warningf("Packet %02X:\n%s", packet[0], hex.Dump(packet))
		}
	}
}

const (
	factoryStickCalibStart = 0x603D
	factoryStickCalibLen   = 25
	userStickCalibStart    = 0x8010
	userStickCalibLen      = 22
)

func (jc *joycon) handleSPIRead(packet []byte) {
	if len(packet) <= 7 {
		log.Warningf("%s: SPI data too short\n%s", jc.Mac(), hex.Dump(packet))
		return
	}
	addr := binary.LittleEndian.Uint32(packet[2:])
	length := packet[6]
	data := packet[7:]

	if int(length)+7 <= len(packet) {
		data = packet[7 : 7+length]
	}

	if addr == factoryStickCalibStart && length == factoryStickCalibLen {
		jc.mu.Lock()
		jc.stickCalib[0].Parse(data[0:9], SideLeft)
		jc.stickCalib[1].Parse(data[9:18], SideRight)
		jc.mu.Unlock()

		jc.listener.OnStickCalib(jc, &jc.stickCalib)
	} else if addr == userStickCalibStart && length == userStickCalibLen {
		had := false
		const magicHaveCalibration = 0xA1B2

		jc.mu.Lock()
		if binary.LittleEndian.Uint16(data[0:2]) == magicHaveCalibration {
			jc.stickCalib[0].Parse(data[2:2+9], SideLeft)
			had = true
		}
		if binary.LittleEndian.Uint16(data[11:13]) == magicHaveCalibration {
			jc.stickCalib[1].Parse(data[13:13+9], SideRight)
			had = true
		}
		jc.mu.Unlock()

		if had {
			jc.listener.OnStickCalib(jc, &jc.stickCalib)
		}
	} else {
		log.Infof("%s: SPI read returned [%x+%d]\n%s", jc.Mac(), addr, length, hex.Dump(data))
	}
}

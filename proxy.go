package main

import (
	"bytes"
	"emu/fanatec"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/sstallion/go-hid"
	"go.bug.st/serial"
)

// VID: 30b7, PID: 1001, Serial: SP248E36055FEDC30D, Product: Heusinkveld Sim Pedals Sprint, Interface: 0
var pedalInfo = hid.DeviceInfo{
	VendorID:  0x30b7,
	ProductID: 0x1001,
}

var mutex sync.Mutex

// HID report from HE pedals
type HEPedalReport struct {
	Id       uint8
	Throttle uint16
	Brake    uint16
	Clutch   uint16
}

type Step struct {
	BaudRate int
	Rx       []byte
	Tx       []byte
}

type ProxyPedals struct {
	fanatec.Pedals
	mutex sync.Mutex
}

func dumpBytes(b []byte) string {
	var sb strings.Builder
	for i, b := range b {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(fmt.Sprintf("%02x", b))
	}
	return sb.String()
}

func dumpStep(s Step) string {
	// create a string buffer and append the Tx and Tx bytes to it
	return fmt.Sprintf("Rx: [%s], Tx: [%s]", dumpBytes(s.Rx), dumpBytes(s.Tx))
}

func fanatec_handler(port serial.Port, pp *ProxyPedals) {

	steps := []Step{
		{BaudRate: 250000, Rx: []uint8{0x0a}, Tx: []uint8{0x1a}},
		{BaudRate: 250000, Rx: []uint8{0x05}, Tx: []uint8{0x15}},
		{
			BaudRate: 115200,
			Rx: []uint8{
				0x7B, 0x02, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x26, 0x7D,
				0x7B, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xAA, 0x7D,
				0x7B, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x5F, 0x7D,
			},
			Tx: []uint8{
				0x7B, 0x05, 0x06, 0x62, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x6D, 0x7D,
				0x7B, 0x07, 0x0B, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x12, 0x7D,
				0x7B, 0x08, 0x01, 0x06, 0x07, 0x00, 0x00, 0x00, 0x00, 0x00, 0xBF, 0x7D,
			},
		},
	}

	buff := make([]byte, 10)

NEXT_STEP:
	for stepNum := 0; stepNum < len(steps); {
		step := steps[stepNum]
		log.Printf("In step[%d]: %s\n", stepNum, dumpStep(step))
		port.SetMode(&serial.Mode{BaudRate: step.BaudRate})
		rxIdx := 0
	NEXT_RX:
		for {
			n, err := port.Read(buff)
			if err != nil {
				log.Fatal("Failed to read serial port:", err)
			}
			// try and find the first rx byte or the next match
			for i := 0; i < n && rxIdx < len(step.Rx); i++ {

				if buff[i] != step.Rx[rxIdx] {
					log.Printf("in step %d, expected [%s], got [%s]\n", stepNum, dumpBytes(step.Rx[rxIdx:]), dumpBytes(buff[i:n]))
					if stepNum > 0 {
						// reset the steps
						stepNum = 0
					}
					rxIdx = 0
					continue NEXT_STEP

				} else {
					log.Printf("in step %d, rxIdx: %d, found: [%s]", stepNum, rxIdx, dumpBytes(buff[i:n]))
					rxIdx++
					if rxIdx >= len(step.Rx) {
						// got a complete match
						break NEXT_RX
					}
				}

			}
		}
		log.Printf("Step [%d], Got [%s], Sending: [%s]\n", stepNum, dumpBytes(step.Rx), dumpBytes(step.Tx))
		port.Write(step.Tx)
		stepNum++
	}

	// must have finished the steps, just keep sending the pedals now
	for {
		pp.mutex.Lock()
		p := pp.CreatePacket()
		pp.mutex.Unlock()
		port.Write(p)
	}

}

func he_handler(pedals *hid.Device, pp *ProxyPedals) {
	last_he := HEPedalReport{}

	for {

		rbuf := make([]byte, 64)
		_, err := pedals.Read(rbuf)
		if err != nil {
			log.Fatal("Failed to read HID device: ", err)
		}

		he := HEPedalReport{}
		err = binary.Read(bytes.NewBuffer(rbuf), binary.LittleEndian, &he)
		if err != nil {
			log.Fatal("binary.Read failed:", err)
		}

		if last_he == he {
			continue
		}
		last_he = he

		pp.mutex.Lock()
		// convert from HE values to 0 - 65535
		pp.Throttle = (he.Throttle << 4) + (he.Throttle >> 8)
		pp.Brake = (he.Brake << 4) + (he.Brake >> 8)
		pp.Clutch = (he.Clutch << 4) + (he.Clutch >> 8)
		pp.mutex.Unlock()
		log.Printf("HE HID Report: %4d / %4d / %4d, Proxy: %5d / %5d / %5d",
			he.Throttle, he.Brake, he.Clutch, pp.Throttle, pp.Brake, pp.Clutch)

	}
}

func main() {

	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("No serial port defined")
	}

	log.Println("starting proxy")
	log.Printf("Opening Pedal Device: %x:%x\n", pedalInfo.VendorID, pedalInfo.ProductID)
	pedals, err := hid.OpenFirst(pedalInfo.VendorID, pedalInfo.ProductID)
	if err != nil {
		log.Fatal("Failed to open HE Pedals HID:", err)
	}
	defer pedals.Close()

	mode := &serial.Mode{
		BaudRate: 250000,
	}

	comport := flag.Arg(0)
	port, err := serial.Open(comport, mode)
	if err != nil {
		log.Fatal("Failed to open serial port: ", comport, " ", err)
	}
	defer port.Close()

	// create the shared pedal report
	pp := ProxyPedals{}

	// start the wheelbase handler
	go fanatec_handler(port, &pp)
	// start the HE handler
	he_handler(pedals, &pp)

}

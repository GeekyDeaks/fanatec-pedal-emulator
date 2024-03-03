package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"log"
	"sync"

	"github.com/sstallion/go-hid"
	"go.bug.st/serial"
)

// VID: 30b7, PID: 1001, Serial: SP248E36055FEDC30D, Product: Heusinkveld Sim Pedals Sprint, Interface: 0
var pedalInfo = hid.DeviceInfo{
	VendorID:  0x30b7,
	ProductID: 0x1001,
}

type ProxyPedal struct {
	mutex    sync.Mutex
	Throttle uint8
	Brake    uint8
	Clutch   uint8
	Combined uint8
}

// HID report from HE pedals
type HEPedalReport struct {
	Id       uint8
	Throttle uint16
	Brake    uint16
	Clutch   uint16
}

func fanatec_handler(port serial.Port, pp *ProxyPedal) {

	buff := make([]byte, 10)
	out := make([]byte, 1)
	for {
		n, err := port.Read(buff)
		if err != nil {
			log.Fatal("Failed to read serial port:", err)
			break
		}
		cmd := buff[n-1] // get the last command asked for
		out[0] = 0x80
		pp.mutex.Lock()
		switch cmd {
		case 0x80:
			out[0] = pp.Clutch | 0x80
		case 0xA0:
			out[0] = pp.Throttle | 0x80
		case 0xC0:
			out[0] = pp.Brake | 0x80
		case 0xE0:
			out[0] = pp.Combined | 0x80

		}
		pp.mutex.Unlock()

		n, err = port.Write(out)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func he_handler(pedals *hid.Device, pp *ProxyPedal) {
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
		log.Printf("HE HID Report: %v", he)

		pp.mutex.Lock()
		// convert from HE values to 0 - 127
		pp.Throttle = 127 - uint8(he.Throttle>>5)
		pp.Brake = 127 - uint8(he.Brake>>5)
		pp.Combined = 127 - uint8(he.Clutch>>5)
		pp.mutex.Unlock()

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
		BaudRate: 230400,
	}

	comport := flag.Arg(0)
	port, err := serial.Open(comport, mode)
	if err != nil {
		log.Fatal("Failed to open serial port: ", comport, " ", err)
	}
	defer port.Close()

	// create the shared pedal report
	pp := ProxyPedal{}

	// start the wheelbase handler
	go fanatec_handler(port, &pp)
	// start the HE handler
	he_handler(pedals, &pp)

}

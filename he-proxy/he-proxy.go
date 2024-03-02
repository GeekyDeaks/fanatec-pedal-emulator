package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/sstallion/go-hid"
)

// VID: 1b4f, PID: 9208, Serial: C, Product: LilyPad USB, Interface: 2
// VID: 2341, PID: 8036, Serial: C, Product: Arduino Leonardo, Interface: 2
var proxyInfo = hid.DeviceInfo{
	VendorID:  0x2341,
	ProductID: 0x8036,
}

// VID: 30b7, PID: 1001, Serial: SP248E36055FEDC30D, Product: Heusinkveld Sim Pedals Sprint, Interface: 0
var pedalInfo = hid.DeviceInfo{
	VendorID:  0x30b7,
	ProductID: 0x1001,
}

type ProxyPedalReport struct {
	Id        uint8
	Throttle  uint8
	Brake     uint8
	Clutch    uint8
	Handbrake uint8
}

type HEPedalReport struct {
	Id       uint8
	Throttle uint16
	Brake    uint16
	Clutch   uint16
}

func main() {
	fmt.Println("starting proxy")
	fmt.Printf("Opening Pedal Device: %x:%x\n", pedalInfo.VendorID, pedalInfo.ProductID)
	pedals, err := hid.OpenFirst(pedalInfo.VendorID, pedalInfo.ProductID)
	if err != nil {
		log.Fatal(err)
	}
	defer pedals.Close()

	fmt.Printf("Opening Proxy Device: %x:%x\n", proxyInfo.VendorID, proxyInfo.ProductID)
	proxy, err := hid.OpenFirst(proxyInfo.VendorID, proxyInfo.ProductID)
	if err != nil {
		log.Fatal(err)
	}
	defer proxy.Close()

	last_he := HEPedalReport{}

	for {

		rbuf := make([]byte, 64)
		_, err := pedals.Read(rbuf)
		if err != nil {
			log.Fatal(err)
		}

		he := HEPedalReport{}
		err = binary.Read(bytes.NewBuffer(rbuf), binary.LittleEndian, &he)
		if err != nil {
			fmt.Println("binary.Read failed:", err)
		}

		if last_he == he {
			continue
		}
		last_he = he
		fmt.Printf("%v", he)

		// convert from HE values to 0 - 127
		p_pedal := ProxyPedalReport{
			Id:       0,
			Throttle: 127 - uint8(he.Throttle>>5),
			Brake:    127 - uint8(he.Brake>>5),
			//Clutch:    uint8(he.Clutch >> 5),
			// Need to check this one...
			Handbrake: 127 - uint8(he.Clutch>>5),
		}

		fmt.Printf(" %v", p_pedal)

		wbuf := bytes.Buffer{}
		err = binary.Write(&wbuf, binary.LittleEndian, p_pedal)
		if err != nil {
			fmt.Println("binary.Write failed:", err)
		}

		wb := wbuf.Bytes()
		for _, b := range wb {
			fmt.Printf(" %02x", b)
		}
		fmt.Println()
		_, err = proxy.Write(wb)
		if err != nil {
			log.Fatal(err)
		}

	}

}

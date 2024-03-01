package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"

	"encoding/binary"

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

	enumerate := flag.Bool("enumerate", false, "enumerate devices")
	flag.Parse()

	if *enumerate {
		ls()
		return
	}

	fmt.Println("starting proxy")
	// Initialize the hid package.
	if err := hid.Init(); err != nil {
		log.Fatal(err)
	}
	defer hid.Exit()

	fmt.Printf("Opening Proxy Device: %x:%x\n", proxyInfo.VendorID, proxyInfo.ProductID)
	proxy, err := hid.OpenFirst(proxyInfo.VendorID, proxyInfo.ProductID)
	if err != nil {
		log.Fatal(err)
	}
	defer proxy.Close()

	fmt.Printf("Opening Pedal Device: %x:%x\n", pedalInfo.VendorID, pedalInfo.ProductID)
	pedals, err := hid.OpenFirst(pedalInfo.VendorID, pedalInfo.ProductID)
	if err != nil {
		log.Fatal(err)
	}
	defer pedals.Close()

	buf := make([]byte, 256)
	last_he := HEPedalReport{}

	for {
		_, err := pedals.Read(buf)
		if err != nil {
			log.Fatal(err)
		}

		he := HEPedalReport{}

		rbufr := bytes.NewReader(buf)
		err = binary.Read(rbufr, binary.LittleEndian, &he)
		if err != nil {
			fmt.Println("binary.Read failed:", err)
		}

		if last_he == he {
			continue
		}
		fmt.Printf("%v", he)
		fmt.Println()
	}

}

func ls() {
	hid.Enumerate(hid.VendorIDAny, hid.ProductIDAny, func(info *hid.DeviceInfo) error {
		fmt.Printf("%s: ID %04x:%04x %s %s\n",
			info.Path,
			info.VendorID,
			info.ProductID,
			info.MfrStr,
			info.ProductStr)
		return nil
	})
}

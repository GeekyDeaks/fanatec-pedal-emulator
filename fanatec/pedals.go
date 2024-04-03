package fanatec

import (
	"encoding/binary"
)

type Pedals struct {
	Throttle  uint16
	Brake     uint16
	Clutch    uint16
	Handbrake uint16
}

func (p *Pedals) CreatePacket() []byte {

	buf := make([]byte, 0, 11)

	buf = append(buf, 0x7b)
	buf = append(buf, 0x01) // send the pedals
	buf = binary.LittleEndian.AppendUint16(buf, p.Throttle)
	buf = binary.LittleEndian.AppendUint16(buf, p.Brake)
	buf = binary.LittleEndian.AppendUint16(buf, p.Clutch)
	buf = binary.LittleEndian.AppendUint16(buf, p.Handbrake)

	crc := GenerateCRC(buf[1:])

	buf = append(buf, crc)
	buf = append(buf, 0x7d)

	return buf

}

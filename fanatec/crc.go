package fanatec

func make_le_table(poly byte) []byte {
	table := make([]byte, 256)
	for i := 0; i < 256; i++ {
		crc := byte(i)
		for j := 0; j < 8; j++ {
			bit := (crc & 0x01) != 0
			crc >>= 1
			if bit {
				crc ^= poly
			}
		}
		table[i] = crc
	}
	return table
}

var crc_table = make_le_table(0x8c)

func GenerateCRC(input []byte) byte {
	var crc byte = 0xff

	for _, b := range input {
		crc = crc_table[b^crc]
	}

	return crc
}

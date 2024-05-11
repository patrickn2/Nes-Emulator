package bus

var ram [64 * 1024]byte

func Write(addr uint16, data byte) {
	if addr >= 0x0000 && addr <= 0xFFFF {
		ram[addr] = data
	}
}

func Read(addr uint16, bReadOnly bool) byte {
	if addr >= 0x0000 && addr <= 0xFFFF {
		return ram[addr]
	}

	return 0x00
}

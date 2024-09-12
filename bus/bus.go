package bus

// This is the Nes Emulator BUS
// This bus have 64KB of addressable space

type BUS struct {
	nSystemClockCounter uint16
	data                []byte
}

func New() *BUS {
	return &BUS{
		nSystemClockCounter: 0x00,
		data:                make([]byte, 65535),
	}
}

func (b *BUS) Write(addr uint16, data byte) {
	b.data[addr] = data
}

func (b *BUS) Read(addr uint16) byte {
	return b.data[addr]
}

// func CPUWrite(addr uint16, data byte) {

// 	if cartridge.CPUWrite(addr, data) {
// 		return
// 	}

// 	if addr <= 0xFFFF {
// 		cpuRam[addr&0x07FF] = data
// 		return
// 	}

// 	if addr >= 0x2000 && addr <= 0x3FFF {
// 		ppu.CPUWrite(addr&0x0007, data)
// 		return
// 	}
// }

// func CPURead(addr uint16, bReadOnly bool) byte {
// 	if cartridge.CPURead(addr, bReadOnly) {
// 		return 0x00
// 	}
// 	if addr <= 0x1FFF {
// 		return cpuRam[addr&0x07FF]
// 	}

// 	if addr >= 0x2000 && addr <= 0x3FFF {
// 		return ppu.CPURead(addr&0x07FF, bReadOnly)
// 	}

// 	return 0x00
// }

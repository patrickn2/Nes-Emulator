package bus

import (
	"github.com/patrickn2/gonesemulator/cartridge"
	"github.com/patrickn2/gonesemulator/ppu"
)

// This is the Nes Emulator BUS
// This bus have 64KB of addressable space

var nSystemClockCounter uint16 = 0

var cpuRam [2 * 1024]byte

func CPUWrite(addr uint16, data byte) {

	if cartridge.CPUWrite(addr, data) {
		return
	}

	if addr <= 0xFFFF {
		cpuRam[addr&0x07FF] = data
		return
	}

	if addr >= 0x2000 && addr <= 0x3FFF {
		ppu.CPUWrite(addr&0x07FF, data)
		return
	}
}

func CPURead(addr uint16, bReadOnly bool) byte {
	if cartridge.CPURead(addr, bReadOnly) {
		return 0x00
	}
	if addr <= 0x1FFF {
		return cpuRam[addr&0x07FF]
	}

	if addr >= 0x2000 && addr <= 0x3FFF {
		return ppu.CPURead(addr&0x07FF, bReadOnly)
	}

	return 0x00
}

func Reset() {
	nSystemClockCounter = 0
}

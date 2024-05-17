package ppu

var tblName = [2][1024]uint8{}
var tblPallete = [32]uint8{}

func CPUWrite(addr uint16, data byte) {
	switch addr {
	case 0x0000: // Control
		break
	case 0x0001: // Mask
		break
	case 0x0002: // Status
		break
	case 0x0003: // OAM Address
		break
	case 0x0004: // OAM Data
		break
	case 0x0005: // Scroll
		break
	case 0x0006: // PPU Address
		break
	case 0x0007: // PPU Data
		break
	}
}

func CPURead(addr uint16, bReadOnly bool) byte {

	switch addr {
	case 0x0000: // Control
		break
	case 0x0001: // Mask
		break
	case 0x0002: // Status
		break
	case 0x0003: // OAM Address
		break
	case 0x0004: // OAM Data
		break
	case 0x0005: // Scroll
		break
	case 0x0006: // PPU Address
		break
	case 0x0007: // PPU Data
		break
	}
	return 0x00
}

func PPURead(addr uint16, bReadOnly bool) byte {

	addr &= 0x3FFF
	return 0x00
}

func PPUWrite(addr uint16, data byte) {
	addr &= 0x3FFF
}

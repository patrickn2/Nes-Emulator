package cpu

import (
	"github.com/patrickn2/gonesemulator/bus"
)

// Public variables
var Status byte = 0x00           // Status Register
var Accumulator byte = 0x00      // Accumulator Register
var XRegister byte = 0x00        // X Register
var YRegister byte = 0x00        // Y Register
var StackPointer byte = 0x00     // Stack Pointer
var ProgramCounter uint16 = 0x00 // Program Counter

type instruction struct {
	name     string
	operate  func() bool
	addrMode func() bool
	cycles   byte
}

var lookup = [256]instruction{
	{"BRK", brk, imp, 7}, {"BRK", brk, imp, 7}, {"BRK", brk, imp, 7}, {"BRK", brk, imp, 7},
}

// Private variables
var fetched byte = 0x00
var addrAbs uint16 = 0x00
var addrRel uint16 = 0x00
var opcode byte = 0x00
var cycles byte = 0x00

var FLAGS6502 = struct {
	C uint8
	Z uint8
	I uint8
	D uint8
	B uint8
	U uint8
	V uint8
	N uint8
}{
	C: 1 << 0, // Carry Bit
	Z: 1 << 1, // Zero
	I: 1 << 2, // Disable Interrupts
	D: 1 << 3, // Decimal Mode (unused in this implementation)
	B: 1 << 4, // Break
	U: 1 << 5, // Unused
	V: 1 << 6, // Overflow
	N: 1 << 7, // Negative
}

// Public functions
func Clock() {
	if cycles == 0 {
		opcode = read(ProgramCounter)
		ProgramCounter++
		cycles = lookup[opcode].cycles
		addrMode := lookup[opcode].addrMode()
		operate := lookup[opcode].operate()
		if addrMode && operate {
			cycles++
		}
	}
	cycles--
}

// Private Functions
func setFlag(flag uint8, value bool) {
	if value {
		Status |= flag
	} else {
		Status &= ^flag
	}
}

func getFlag(flag uint8) uint8 {
	if (Status & flag) > 0 {
		return 1
	}
	return 0
}

func read(addr uint16) byte {
	return bus.Read(addr, false)
}

func write(addr uint16, data byte) {
	bus.Write(addr, data)
}

func imp() bool {
	fetched = Accumulator
	return false
}

func imm() bool {
	addrAbs = ProgramCounter
	ProgramCounter++
	return false
}

func zp0() bool {
	addrAbs = uint16(read(ProgramCounter))
	ProgramCounter++
	addrAbs &= 0x00FF
	return false
}

func zpx() bool {
	addrAbs = uint16(read(ProgramCounter) + XRegister)
	ProgramCounter++
	addrAbs &= 0x00FF
	return false
}

func zpy() bool {
	addrAbs = uint16(read(ProgramCounter) + YRegister)
	ProgramCounter++
	addrAbs &= 0x00FF
	return false
}

func abs() bool {
	lo := read(ProgramCounter)
	ProgramCounter++
	hi := read(ProgramCounter)
	ProgramCounter++
	addrAbs = uint16((hi << 8) | lo)
	return false
}

func abx() bool {
	lo := read(ProgramCounter)
	ProgramCounter++
	hi := read(ProgramCounter)
	ProgramCounter++
	addrAbs = uint16((hi << 8) | lo)
	addrAbs += uint16(XRegister)
	if (addrAbs & 0xFF00) != uint16(hi<<8) {
		return true
	}
	return false
}

func aby() bool {
	lo := read(ProgramCounter)
	ProgramCounter++
	hi := read(ProgramCounter)
	ProgramCounter++
	addrAbs = uint16((hi << 8) | lo)
	addrAbs += uint16(YRegister)
	if (addrAbs & 0xFF00) != uint16(hi<<8) {
		return true
	}
	return false
}

func ind() bool {
	lo := read(ProgramCounter)
	ProgramCounter++
	hi := read(ProgramCounter)
	ProgramCounter++

	ptr := uint16((hi << 8) | lo)

	if lo == 0x00FF {
		addrAbs = uint16(read(ptr&0xFF00))<<8 | uint16(read(ptr))
	} else {
		addrAbs = uint16(read(ptr+1))<<8 | uint16(read(ptr))
	}

	return false
}

func idx() bool {
	t := read(ProgramCounter)
	ProgramCounter++

	lo := read(uint16(t+XRegister) & 0x00FF)
	hi := read(uint16(t+XRegister+1) & 0x00FF)
	addrAbs = uint16((hi << 8) | lo)
	return false
}

func idy() bool {
	t := read(ProgramCounter)
	ProgramCounter++

	lo := read(uint16(t) & 0x00FF)
	hi := read(uint16(t+1) & 0x00FF)
	addrAbs = uint16((hi << 8) | lo)
	addrAbs += uint16(YRegister)
	if (addrAbs & 0xFF00) != uint16(hi<<8) {
		return true
	}
	return false
}

func rel() bool {
	addrRel = uint16((ProgramCounter))
	ProgramCounter++
	if (addrRel & 0x80) != 0 {
		addrRel |= 0xFF00
	}
	return false
}

// Instructions

func fetch() byte {
	if lookup[opcode].name == "IMP" {
		fetched = read(addrAbs)
	}
	return fetched
}

func and() bool {
	fetch()
	Accumulator = Accumulator & fetched
	setFlag(FLAGS6502.Z, Accumulator == 0x00)
	setFlag(FLAGS6502.N, (Accumulator&0x00) != 0)
	return true
}

func asl() bool {
	fetch()
	temp := fetched << 1
	setFlag(FLAGS6502.C, (temp&0xFF00) > 0)
	setFlag(FLAGS6502.Z, (temp&0x00FF) == 0)
	setFlag(FLAGS6502.N, (temp&0x80) != 0)
	if lookup[opcode].name == "IMP" {
		Accumulator = temp & 0x00FF
	} else {
		write(addrAbs, temp&0x00FF)
	}
	return false
}

func bcs() bool {
	if getFlag(FLAGS6502.C) == 1 {
		cycles++
		addrAbs = ProgramCounter + addrRel
		if (addrAbs & 0xFF00) != (ProgramCounter & 0xFF00) {
			cycles++
		}
		ProgramCounter = addrAbs
	}
	return false
}

func bcc() bool {
	if getFlag(FLAGS6502.C) == 0 {
		cycles++
		addrAbs = ProgramCounter + addrRel
		if (addrAbs & 0xFF00) != (ProgramCounter & 0xFF00) {
			cycles++
		}
		ProgramCounter = addrAbs
	}
	return false
}

func beq() bool {
	if getFlag(FLAGS6502.Z) == 1 {
		cycles++
		addrAbs = ProgramCounter + addrRel
		if (addrAbs & 0xFF00) != (ProgramCounter & 0xFF00) {
			cycles++
		}
		ProgramCounter = addrAbs
	}
	return false
}

func bit() bool {
	fetch()
	temp := Accumulator & fetched
	setFlag(FLAGS6502.Z, (temp&0x00FF) == 0)
	setFlag(FLAGS6502.N, fetched&(1<<7) > 0)
	setFlag(FLAGS6502.V, fetched&(1<<6) > 0)
	return false
}

func bmi() bool {
	if getFlag(FLAGS6502.N) == 1 {
		cycles++
		addrAbs = ProgramCounter + addrRel
		if (addrAbs & 0xFF00) != (ProgramCounter & 0xFF00) {
			cycles++
		}
		ProgramCounter = addrAbs
	}
	return false
}

func bne() bool {
	if getFlag(FLAGS6502.Z) == 0 {
		cycles++
		addrAbs = ProgramCounter + addrRel
		if (addrAbs & 0xFF00) != (ProgramCounter & 0xFF00) {
			cycles++
		}
		ProgramCounter = addrAbs
	}
	return false
}

func bpl() bool {
	if getFlag(FLAGS6502.N) == 0 {
		cycles++
		addrAbs = ProgramCounter + addrRel
		if (addrAbs & 0xFF00) != (ProgramCounter & 0xFF00) {
			cycles++
		}
		ProgramCounter = addrAbs
	}
	return false
}

func bvc() bool {
	if getFlag(FLAGS6502.Z) == 0 {
		cycles++
		addrAbs = ProgramCounter + addrRel
		if (addrAbs & 0xFF00) != (ProgramCounter & 0xFF00) {
			cycles++
		}
		ProgramCounter = addrAbs
	}
	return false
}

func bvs() bool {
	if getFlag(FLAGS6502.V) == 1 {
		cycles++
		addrAbs = ProgramCounter + addrRel
		if (addrAbs & 0xFF00) != (ProgramCounter & 0xFF00) {
			cycles++
		}
		ProgramCounter = addrAbs
	}
	return false
}

func clc() bool {
	setFlag(FLAGS6502.C, false)
	return false
}

func cld() bool {
	setFlag(FLAGS6502.D, false)
	return false
}

func cli() bool {
	setFlag(FLAGS6502.I, false)
	return false
}

func clv() bool {
	setFlag(FLAGS6502.V, false)
	return false
}

func cmp() bool {
	fetch()
	temp := uint16(Accumulator - fetched)
	setFlag(FLAGS6502.C, temp < uint16(Accumulator))
	setFlag(FLAGS6502.Z, (temp&0x00FF) == 0)
	setFlag(FLAGS6502.N, (temp&0x0080) != 0)
	return false
}

func cpx() bool {
	fetch()
	temp := uint16(XRegister - fetched)
	setFlag(FLAGS6502.C, temp < uint16(XRegister))
	setFlag(FLAGS6502.Z, (temp&0x00FF) == 0)
	setFlag(FLAGS6502.N, (temp&0x0080) != 0)
	return false
}

func cpy() bool {
	fetch()
	temp := uint16(YRegister - fetched)
	setFlag(FLAGS6502.C, temp < uint16(YRegister))
	setFlag(FLAGS6502.Z, (temp&0x00FF) == 0)
	setFlag(FLAGS6502.N, (temp&0x0080) != 0)
	return false
}

func dec() bool {
	fetch()
	temp := fetched - 1
	write(addrAbs, temp&0x00FF)
	setFlag(FLAGS6502.Z, (temp&0x00FF) == 0)
	setFlag(FLAGS6502.N, (temp&0x80) != 0)
	return false
}

func dex() bool {
	XRegister--
	setFlag(FLAGS6502.Z, (XRegister&0x00FF) == 0)
	setFlag(FLAGS6502.N, (XRegister&0x80) != 0)
	return false
}

func dey() bool {
	YRegister--
	setFlag(FLAGS6502.Z, (YRegister&0x00FF) == 0)
	setFlag(FLAGS6502.N, (YRegister&0x80) != 0)
	return false
}

func eor() bool {
	fetch()
	Accumulator = Accumulator ^ fetched
	setFlag(FLAGS6502.Z, Accumulator == 0x00)
	setFlag(FLAGS6502.N, (Accumulator&0x80) != 0)
	return true
}

func inc() bool {
	fetch()
	temp := fetched + 1
	write(addrAbs, temp&0x00FF)
	setFlag(FLAGS6502.Z, (temp&0x00FF) == 0)
	setFlag(FLAGS6502.N, (temp&0x80) != 0)
	return false
}

func inx() bool {
	XRegister++
	setFlag(FLAGS6502.Z, (XRegister&0x00FF) == 0)
	setFlag(FLAGS6502.N, (XRegister&0x80) != 0)
	return false
}

func iny() bool {
	YRegister++
	setFlag(FLAGS6502.Z, (YRegister&0x00FF) == 0)
	setFlag(FLAGS6502.N, (YRegister&0x80) != 0)
	return false
}

func jmp() bool {
	ProgramCounter = addrAbs
	return false
}

func jsr() bool {
	ProgramCounter--

	write(0x0100+uint16(StackPointer), byte((ProgramCounter>>8)&0x00FF))
	StackPointer--
	write(0x0100+uint16(StackPointer), byte(ProgramCounter&0x00FF))
	StackPointer--
	ProgramCounter = addrAbs
	return false
}

func lda() bool {
	fetch()
	Accumulator = fetched
	setFlag(FLAGS6502.Z, Accumulator == 0x00)
	setFlag(FLAGS6502.N, (Accumulator&0x80) != 0)
	return false
}

func ldx() bool {
	fetch()
	XRegister = fetched
	setFlag(FLAGS6502.Z, XRegister == 0x00)
	setFlag(FLAGS6502.N, (XRegister&0x80) != 0)
	return false
}

func ldy() bool {
	fetch()
	YRegister = fetched
	setFlag(FLAGS6502.Z, YRegister == 0x00)
	setFlag(FLAGS6502.N, (YRegister&0x80) != 0)
	return false
}

func lsr() bool {
	fetch()
	setFlag(FLAGS6502.C, (fetched&0x01) > 0)
	temp := fetched >> 1
	setFlag(FLAGS6502.Z, (temp&0x00FF) == 0)
	setFlag(FLAGS6502.N, (temp&0x80) != 0)
	if lookup[opcode].name == "IMP" {
		Accumulator = temp & 0x00FF
	} else {
		write(addrAbs, temp&0x00FF)
	}
	return false
}

func nop() bool {
	switch opcode {
	case 0x1C:
	case 0x3C:
	case 0x5C:
	case 0x7C:
	case 0xDC:
	case 0xFC:
		return true
	}
	return false
}

func adc() bool {
	fetch()
	temp := uint16(Accumulator) + uint16(fetched) + uint16(getFlag(FLAGS6502.C))
	setFlag(FLAGS6502.C, temp > 255)
	setFlag(FLAGS6502.Z, (temp&0x00FF) == 0)
	setFlag(FLAGS6502.N, (temp&0x80) != 0)
	setFlag(FLAGS6502.V, ((Accumulator^fetched)&(Accumulator^byte(temp))&0x80) != 0)
	Accumulator = byte(temp & 0x00FF)
	return true
}

func sbc() bool {
	fetch()
	value := uint16(fetched) ^ 0x00FF
	temp := uint16(Accumulator) + value + uint16(getFlag(FLAGS6502.C))
	setFlag(FLAGS6502.C, temp&0xFF00 != 0)
	setFlag(FLAGS6502.Z, (temp&0x00FF) == 0)
	setFlag(FLAGS6502.N, (temp&0x80) != 0)
	setFlag(FLAGS6502.V, (temp^(uint16(Accumulator)^value))&0x80 != 0)
	Accumulator = byte(temp & 0x00FF)
	return true
}

func pha() bool {
	write(0x0100+uint16(StackPointer), Accumulator)
	StackPointer--
	return false
}

func php() bool {
	write(0x0100+uint16(StackPointer), Status|FLAGS6502.B|FLAGS6502.U)
	setFlag(FLAGS6502.B, false)
	setFlag(FLAGS6502.U, false)
	StackPointer--
	return false
}

func pla() bool {
	StackPointer++
	Accumulator = read(0x0100 + uint16(StackPointer))
	setFlag(FLAGS6502.Z, Accumulator == 0x00)
	setFlag(FLAGS6502.N, (Accumulator&0x80) != 0)
	return false
}

func plp() bool {
	StackPointer++
	Status = read(0x0100 + uint16(StackPointer))
	setFlag(FLAGS6502.U, true)
	return false
}

func rol() bool {
	fetch()
	temp := (fetched << 1) | getFlag(FLAGS6502.C)
	setFlag(FLAGS6502.C, (uint16(temp)&0xFF00) > 0)
	setFlag(FLAGS6502.Z, (temp&0x00FF) == 0)
	setFlag(FLAGS6502.N, (temp&0x0080) != 0)
	if lookup[opcode].name == "IMP" {
		Accumulator = temp & 0x00FF
	} else {
		write(addrAbs, temp&0x00FF)
	}
	return false
}

func ror() bool {
	fetch()
	temp := (getFlag(FLAGS6502.C) << 7) | (fetched >> 1)
	setFlag(FLAGS6502.C, (uint16(temp)&0x01) != 0)
	setFlag(FLAGS6502.Z, (temp&0x00FF) == 0)
	setFlag(FLAGS6502.N, (temp&0x80) != 0)
	if lookup[opcode].name == "IMP" {
		Accumulator = temp & 0x00FF
	} else {
		write(addrAbs, temp&0x00FF)
	}
	return false
}

func reset() {
	addrAbs = 0xFFFC
	lo := read(addrAbs + 0)
	hi := read(addrAbs + 1)
	ProgramCounter = (uint16(hi) << 8) | uint16(lo)
	Status = 0x00 | FLAGS6502.U
	StackPointer = 0xFD
	addrAbs = 0x0000
	addrRel = 0x0000
	fetched = 0x00
	cycles = 8
}

func irq() {
	if getFlag(FLAGS6502.I) == 0 {
		write(0x0100+uint16(StackPointer), byte((ProgramCounter>>8)&0x00FF))
		StackPointer--
		write(0x0100+uint16(StackPointer), byte(ProgramCounter&0x00FF))
		StackPointer--

		setFlag(FLAGS6502.B, false)
		setFlag(FLAGS6502.U, true)
		setFlag(FLAGS6502.I, true)
		write(0x0100+uint16(StackPointer), Status)
		StackPointer--
		addrAbs = 0xFFFE
		lo := read(addrAbs + 0)
		hi := read(addrAbs + 1)
		ProgramCounter = (uint16(hi) << 8) | uint16(lo)
		cycles = 7
	}
}

func nmi() {
	write(0x0100+uint16(StackPointer), byte((ProgramCounter>>8)&0x00FF))
	StackPointer--
	write(0x0100+uint16(StackPointer), byte(ProgramCounter&0x00FF))
	StackPointer--

	setFlag(FLAGS6502.B, false)
	setFlag(FLAGS6502.U, true)
	setFlag(FLAGS6502.I, true)
	write(0x0100+uint16(StackPointer), Status)
	StackPointer--
	addrAbs = 0xFFFA
	lo := read(addrAbs + 0)
	hi := read(addrAbs + 1)
	ProgramCounter = (uint16(hi) << 8) | uint16(lo)
	cycles = 8
}

func rti() bool {
	StackPointer++
	Status = read(0x0100 + uint16(StackPointer))
	Status &= ^FLAGS6502.B
	Status &= ^FLAGS6502.U
	StackPointer++
	ProgramCounter = uint16(read(0x0100 + uint16(StackPointer)))
	StackPointer++
	ProgramCounter |= uint16(read(0x0100+uint16(StackPointer))) << 8
	return false
}

func rts() bool {
	StackPointer++
	ProgramCounter = uint16(read(0x0100 + uint16(StackPointer)))
	StackPointer++
	ProgramCounter |= uint16(read(0x0100+uint16(StackPointer))) << 8
	ProgramCounter++
	return false
}

func sec() bool {
	setFlag(FLAGS6502.C, true)
	return false
}

func sed() bool {
	setFlag(FLAGS6502.D, true)
	return false
}

func sei() bool {
	setFlag(FLAGS6502.I, true)
	return false
}

func sta() bool {
	write(addrAbs, Accumulator)
	return false
}

func stx() bool {
	write(addrAbs, XRegister)
	return false
}

func sty() bool {
	write(addrAbs, YRegister)
	return false
}

func tax() bool {
	XRegister = Accumulator
	setFlag(FLAGS6502.Z, XRegister == 0x00)
	setFlag(FLAGS6502.N, (XRegister&0x80) != 0)
	return false
}

func tay() bool {
	YRegister = Accumulator
	setFlag(FLAGS6502.Z, YRegister == 0x00)
	setFlag(FLAGS6502.N, (YRegister&0x80) != 0)
	return false
}

func tsx() bool {
	XRegister = StackPointer
	setFlag(FLAGS6502.Z, XRegister == 0x00)
	setFlag(FLAGS6502.N, (XRegister&0x80) != 0)
	return false
}

func txa() bool {
	Accumulator = XRegister
	setFlag(FLAGS6502.Z, Accumulator == 0x00)
	setFlag(FLAGS6502.N, (Accumulator&0x80) != 0)
	return false
}

func txs() bool {
	StackPointer = XRegister
	return false
}

func tya() bool {
	Accumulator = YRegister
	setFlag(FLAGS6502.Z, Accumulator == 0x00)
	setFlag(FLAGS6502.N, (Accumulator&0x80) != 0)
	return false
}

func xxx() bool {
	return false
}

func brk() bool {
	ProgramCounter++
	setFlag(FLAGS6502.I, true)
	write(0x0100+uint16(StackPointer), byte((ProgramCounter>>8)&0x00FF))
	StackPointer--
	write(0x0100+uint16(StackPointer), byte(ProgramCounter&0x00FF))
	StackPointer--

	setFlag(FLAGS6502.B, true)
	write(0x0100+uint16(StackPointer), Status)
	StackPointer--
	setFlag(FLAGS6502.B, false)

	lo := read(0xFFFE)
	hi := read(0xFFFF)
	ProgramCounter = uint16((hi << 8) | lo)
	return false
}

func ora() bool {
	fetch()
	Accumulator = Accumulator | fetched
	setFlag(FLAGS6502.Z, Accumulator == 0x00)
	setFlag(FLAGS6502.N, (Accumulator&0x80) != 0)
	return true
}

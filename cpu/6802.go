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

var instruction = struct {
	name     string
	operate  func() bool
	addrMode func() bool
	cycles   byte
}{}

var lookup = [256]instruction{
	{"BRK", brk, imp, 7}, {"ORA", ora, izx, 6}, {"KIL", KIL, IMP, 2}, {"SLO", SLO, IZX, 8}, {"NOP", NOP, ZP0, 3}, {"ORA", ORA, ZP0, 3}, {"ASL", ASL, ZP0, 5}, {"SLO", SLO, ZP0, 5}, {"PHP", PHP, IMP, 3}, {"ORA", ORA, IMM, 2}, {"ASL", ASL, IMP, 2}, {"ANC", ANC, IMM, 2}, {"NOP", NOP, ABS, 4}, {"ORA", ORA, ABS, 4}, {"ASL", ASL, abs, 6}, {"SLO", SLO, ABS, 6},
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
		if lookup[opcode].addrMode() && lookup[opcode].operate() {
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

func pla() bool {
	StackPointer++
	Accumulator = read(0x0100 + uint16(StackPointer))
	setFlag(FLAGS6502.Z, Accumulator == 0x00)
	setFlag(FLAGS6502.N, (Accumulator&0x80) != 0)
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

func rti() {
	StackPointer++
	Status = read(0x0100 + uint16(StackPointer))
	Status &= ^FLAGS6502.B
	Status &= ^FLAGS6502.U
	StackPointer++
	lo := uint16(read(0x0100 + uint16(StackPointer)))
	StackPointer++
	hi := uint16(read(0x0100 + uint16(StackPointer)))
	ProgramCounter = (hi << 8) | lo
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

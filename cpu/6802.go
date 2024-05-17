package cpu

import (
	"reflect"

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

var lookup = [256]instruction{}

// Private variables
var fetched byte = 0x00
var addrAbs uint16 = 0x00
var addrRel uint16 = 0x00
var opcode byte = 0x00
var cycles byte = 0x00

var Flags = struct {
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
	return bus.CPURead(addr, false)
}

func write(addr uint16, data byte) {
	bus.CPUWrite(addr, data)
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
	addrAbs = uint16(hi)<<8 | uint16(lo)
	return false
}

func abx() bool {
	lo := read(ProgramCounter)
	ProgramCounter++
	hi := read(ProgramCounter)
	ProgramCounter++
	addrAbs = uint16(hi)<<8 | uint16(lo)
	addrAbs += uint16(XRegister)
	return uint16(addrAbs&0xFF00) != uint16(hi)<<8
}

func aby() bool {
	lo := read(ProgramCounter)
	ProgramCounter++
	hi := read(ProgramCounter)
	ProgramCounter++
	addrAbs = uint16(hi)<<8 | uint16(lo)
	addrAbs += uint16(YRegister)
	return uint16(addrAbs&0xFF00) != uint16(hi)<<8
}

func ind() bool {
	lo := read(ProgramCounter)
	ProgramCounter++
	hi := read(ProgramCounter)
	ProgramCounter++

	ptr := uint16(hi)<<8 | uint16(lo)

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
	addrAbs = uint16(hi)<<8 | uint16(lo)
	return false
}

func idy() bool {
	t := read(ProgramCounter)
	ProgramCounter++

	lo := read(uint16(t) & 0x00FF)
	hi := read(uint16(t+1) & 0x00FF)
	addrAbs = uint16(hi)<<8 | uint16(lo)
	addrAbs += uint16(YRegister)
	return (addrAbs & 0xFF00) != uint16(hi)<<8
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
	if reflect.ValueOf(lookup[opcode].addrMode).Pointer() == reflect.ValueOf(imp).Pointer() {
		fetched = read(addrAbs)
	}
	return fetched
}

func and() bool {
	fetch()
	Accumulator = Accumulator & fetched
	setFlag(Flags.Z, Accumulator == 0x00)
	setFlag(Flags.N, (Accumulator&0x80) != 0)
	return true
}

func asl() bool {
	fetch()
	temp := fetched << 1
	setFlag(Flags.C, (uint16(temp)&0xFF00) > 0)
	setFlag(Flags.Z, (temp&0x00FF) == 0)
	setFlag(Flags.N, (temp&0x80) != 0)
	if reflect.ValueOf(lookup[opcode].addrMode).Pointer() == reflect.ValueOf(imp).Pointer() {
		Accumulator = temp & 0x00FF
	} else {
		write(addrAbs, temp&0x00FF)
	}
	return false
}

func bcs() bool {
	if getFlag(Flags.C) == 1 {
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
	if getFlag(Flags.C) == 0 {
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
	if getFlag(Flags.Z) == 1 {
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
	setFlag(Flags.Z, (temp&0x00FF) == 0)
	setFlag(Flags.N, fetched&(1<<7) > 0)
	setFlag(Flags.V, fetched&(1<<6) > 0)
	return false
}

func bmi() bool {
	if getFlag(Flags.N) == 1 {
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
	if getFlag(Flags.Z) == 0 {
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
	if getFlag(Flags.N) == 0 {
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
	if getFlag(Flags.Z) == 0 {
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
	if getFlag(Flags.V) == 1 {
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
	setFlag(Flags.C, false)
	return false
}

func cld() bool {
	setFlag(Flags.D, false)
	return false
}

func cli() bool {
	setFlag(Flags.I, false)
	return false
}

func clv() bool {
	setFlag(Flags.V, false)
	return false
}

func cmp() bool {
	fetch()
	temp := uint16(Accumulator - fetched)
	setFlag(Flags.C, temp < uint16(Accumulator))
	setFlag(Flags.Z, (temp&0x00FF) == 0)
	setFlag(Flags.N, (temp&0x0080) != 0)
	return false
}

func cpx() bool {
	fetch()
	temp := uint16(XRegister - fetched)
	setFlag(Flags.C, temp < uint16(XRegister))
	setFlag(Flags.Z, (temp&0x00FF) == 0)
	setFlag(Flags.N, (temp&0x0080) != 0)
	return false
}

func cpy() bool {
	fetch()
	temp := uint16(YRegister - fetched)
	setFlag(Flags.C, temp < uint16(YRegister))
	setFlag(Flags.Z, (temp&0x00FF) == 0)
	setFlag(Flags.N, (temp&0x0080) != 0)
	return false
}

func dec() bool {
	fetch()
	temp := fetched - 1
	write(addrAbs, temp&0x00FF)
	setFlag(Flags.Z, (temp&0x00FF) == 0)
	setFlag(Flags.N, (temp&0x80) != 0)
	return false
}

func dex() bool {
	XRegister--
	setFlag(Flags.Z, (XRegister&0x00FF) == 0)
	setFlag(Flags.N, (XRegister&0x80) != 0)
	return false
}

func dey() bool {
	YRegister--
	setFlag(Flags.Z, (YRegister&0x00FF) == 0)
	setFlag(Flags.N, (YRegister&0x80) != 0)
	return false
}

func eor() bool {
	fetch()
	Accumulator = Accumulator ^ fetched
	setFlag(Flags.Z, Accumulator == 0x00)
	setFlag(Flags.N, (Accumulator&0x80) != 0)
	return true
}

func inc() bool {
	fetch()
	temp := fetched + 1
	write(addrAbs, temp&0x00FF)
	setFlag(Flags.Z, (temp&0x00FF) == 0)
	setFlag(Flags.N, (temp&0x80) != 0)
	return false
}

func inx() bool {
	XRegister++
	setFlag(Flags.Z, (XRegister&0x00FF) == 0)
	setFlag(Flags.N, (XRegister&0x80) != 0)
	return false
}

func iny() bool {
	YRegister++
	setFlag(Flags.Z, (YRegister&0x00FF) == 0)
	setFlag(Flags.N, (YRegister&0x80) != 0)
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
	setFlag(Flags.Z, Accumulator == 0x00)
	setFlag(Flags.N, (Accumulator&0x80) != 0)
	return false
}

func ldx() bool {
	fetch()
	XRegister = fetched
	setFlag(Flags.Z, XRegister == 0x00)
	setFlag(Flags.N, (XRegister&0x80) != 0)
	return false
}

func ldy() bool {
	fetch()
	YRegister = fetched
	setFlag(Flags.Z, YRegister == 0x00)
	setFlag(Flags.N, (YRegister&0x80) != 0)
	return false
}

func lsr() bool {
	fetch()
	setFlag(Flags.C, (fetched&0x01) > 0)
	temp := fetched >> 1
	setFlag(Flags.Z, (temp&0x00FF) == 0)
	setFlag(Flags.N, (temp&0x80) != 0)
	if reflect.ValueOf(lookup[opcode].addrMode).Pointer() == reflect.ValueOf(imp).Pointer() {
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
	temp := uint16(Accumulator) + uint16(fetched) + uint16(getFlag(Flags.C))
	setFlag(Flags.C, temp > 255)
	setFlag(Flags.Z, (temp&0x00FF) == 0)
	setFlag(Flags.N, (temp&0x80) != 0)
	setFlag(Flags.V, ((Accumulator^fetched)&(Accumulator^byte(temp))&0x80) != 0)
	Accumulator = byte(temp & 0x00FF)
	return true
}

func sbc() bool {
	fetch()
	value := uint16(fetched) ^ 0x00FF
	temp := uint16(Accumulator) + value + uint16(getFlag(Flags.C))
	setFlag(Flags.C, temp&0xFF00 != 0)
	setFlag(Flags.Z, (temp&0x00FF) == 0)
	setFlag(Flags.N, (temp&0x80) != 0)
	setFlag(Flags.V, (temp^(uint16(Accumulator)^value))&0x80 != 0)
	Accumulator = byte(temp & 0x00FF)
	return true
}

func pha() bool {
	write(0x0100+uint16(StackPointer), Accumulator)
	StackPointer--
	return false
}

func php() bool {
	write(0x0100+uint16(StackPointer), Status|Flags.B|Flags.U)
	setFlag(Flags.B, false)
	setFlag(Flags.U, false)
	StackPointer--
	return false
}

func pla() bool {
	StackPointer++
	Accumulator = read(0x0100 + uint16(StackPointer))
	setFlag(Flags.Z, Accumulator == 0x00)
	setFlag(Flags.N, (Accumulator&0x80) != 0)
	return false
}

func plp() bool {
	StackPointer++
	Status = read(0x0100 + uint16(StackPointer))
	setFlag(Flags.U, true)
	return false
}

func rol() bool {
	fetch()
	temp := (fetched << 1) | getFlag(Flags.C)
	setFlag(Flags.C, (uint16(temp)&0xFF00) > 0)
	setFlag(Flags.Z, (temp&0x00FF) == 0)
	setFlag(Flags.N, (temp&0x0080) != 0)
	if reflect.ValueOf(lookup[opcode].addrMode).Pointer() == reflect.ValueOf(imp).Pointer() {
		Accumulator = temp & 0x00FF
	} else {
		write(addrAbs, temp&0x00FF)
	}
	return false
}

func ror() bool {
	fetch()
	temp := (getFlag(Flags.C) << 7) | (fetched >> 1)
	setFlag(Flags.C, (uint16(temp)&0x01) != 0)
	setFlag(Flags.Z, (temp&0x00FF) == 0)
	setFlag(Flags.N, (temp&0x80) != 0)
	if reflect.ValueOf(lookup[opcode].addrMode).Pointer() == reflect.ValueOf(imp).Pointer() {
		Accumulator = temp & 0x00FF
	} else {
		write(addrAbs, temp&0x00FF)
	}
	return false
}

func Reset() {
	addrAbs = 0xFFFC
	lo := read(addrAbs + 0)
	hi := read(addrAbs + 1)
	ProgramCounter = (uint16(hi) << 8) | uint16(lo)
	Status = 0x00 | Flags.U
	StackPointer = 0xFD
	addrAbs = 0x0000
	addrRel = 0x0000
	fetched = 0x00
	cycles = 8
	bus.Reset()
}

func irq() {
	if getFlag(Flags.I) == 0 {
		write(0x0100+uint16(StackPointer), byte((ProgramCounter>>8)&0x00FF))
		StackPointer--
		write(0x0100+uint16(StackPointer), byte(ProgramCounter&0x00FF))
		StackPointer--

		setFlag(Flags.B, false)
		setFlag(Flags.U, true)
		setFlag(Flags.I, true)
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

	setFlag(Flags.B, false)
	setFlag(Flags.U, true)
	setFlag(Flags.I, true)
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
	Status &= ^Flags.B
	Status &= ^Flags.U
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
	setFlag(Flags.C, true)
	return false
}

func sed() bool {
	setFlag(Flags.D, true)
	return false
}

func sei() bool {
	setFlag(Flags.I, true)
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
	setFlag(Flags.Z, XRegister == 0x00)
	setFlag(Flags.N, (XRegister&0x80) != 0)
	return false
}

func tay() bool {
	YRegister = Accumulator
	setFlag(Flags.Z, YRegister == 0x00)
	setFlag(Flags.N, (YRegister&0x80) != 0)
	return false
}

func tsx() bool {
	XRegister = StackPointer
	setFlag(Flags.Z, XRegister == 0x00)
	setFlag(Flags.N, (XRegister&0x80) != 0)
	return false
}

func txa() bool {
	Accumulator = XRegister
	setFlag(Flags.Z, Accumulator == 0x00)
	setFlag(Flags.N, (Accumulator&0x80) != 0)
	return false
}

func txs() bool {
	StackPointer = XRegister
	return false
}

func tya() bool {
	Accumulator = YRegister
	setFlag(Flags.Z, Accumulator == 0x00)
	setFlag(Flags.N, (Accumulator&0x80) != 0)
	return false
}

func xxx() bool {
	return false
}

func brk() bool {
	ProgramCounter++
	setFlag(Flags.I, true)
	write(0x0100+uint16(StackPointer), byte((ProgramCounter>>8)&0x00FF))
	StackPointer--
	write(0x0100+uint16(StackPointer), byte(ProgramCounter&0x00FF))
	StackPointer--

	setFlag(Flags.B, true)
	write(0x0100+uint16(StackPointer), Status)
	StackPointer--
	setFlag(Flags.B, false)

	lo := read(0xFFFE)
	hi := read(0xFFFF)
	ProgramCounter = uint16(hi)<<8 | uint16(lo)
	return false
}

func ora() bool {
	fetch()
	Accumulator = Accumulator | fetched
	setFlag(Flags.Z, Accumulator == 0x00)
	setFlag(Flags.N, (Accumulator&0x80) != 0)
	return true
}

func izx() bool {
	t := uint16(read(ProgramCounter))
	ProgramCounter++
	lo := read((t + uint16(XRegister)&0x00FF))
	hi := read(((t + uint16(XRegister) + 1) & 0x00FF))

	addrAbs = uint16(hi)<<8 | uint16(lo)
	return false
}

func izy() bool {
	t := uint16(read(ProgramCounter))
	ProgramCounter++
	lo := read(t & 0x00FF)
	hi := read((t + 1) & 0x00FF)

	addrAbs = uint16(hi)<<8 | uint16(lo)
	addrAbs += uint16(YRegister)
	return (addrAbs & 0xFF00) != uint16(hi)<<8
}

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

func init() {
	lookup = [256]instruction{
		{"BRK", brk, imm, 7}, {"ORA", ora, izx, 6}, {"???", xxx, imp, 2}, {"???", xxx, imp, 8}, {"???", nop, imp, 3}, {"ORA", ora, zp0, 3}, {"ASL", asl, zp0, 5}, {"???", xxx, imp, 5}, {"PHP", php, imp, 3}, {"ORA", ora, imm, 2}, {"ASL", asl, imp, 2}, {"???", xxx, imp, 2}, {"???", nop, imp, 4}, {"ORA", ora, abs, 4}, {"ASL", asl, abs, 6}, {"???", xxx, imp, 6}, {"BPL", bpl, rel, 2}, {"ORA", ora, izy, 5}, {"???", xxx, imp, 2}, {"???", xxx, imp, 8}, {"???", nop, imp, 4}, {"ORA", ora, zpx, 4}, {"ASL", asl, zpx, 6}, {"???", xxx, imp, 6}, {"CLC", clc, imp, 2}, {"ORA", ora, aby, 4}, {"???", nop, imp, 2}, {"???", xxx, imp, 7}, {"???", nop, imp, 4}, {"ORA", ora, abx, 4}, {"ASL", asl, abx, 7}, {"???", xxx, imp, 7}, {"JSR", jsr, abs, 6}, {"AND", and, izx, 6}, {"???", xxx, imp, 2}, {"???", xxx, imp, 8}, {"BIT", bit, zp0, 3}, {"AND", and, zp0, 3}, {"ROL", rol, zp0, 5}, {"???", xxx, imp, 5}, {"PLP", plp, imp, 4}, {"AND", and, imm, 2}, {"ROL", rol, imp, 2}, {"???", xxx, imp, 2}, {"BIT", bit, abs, 4}, {"AND", and, abs, 4}, {"ROL", rol, abs, 6}, {"???", xxx, imp, 6}, {"BMI", bmi, rel, 2}, {"AND", and, izy, 5}, {"???", xxx, imp, 2}, {"???", xxx, imp, 8}, {"???", nop, imp, 4}, {"AND", and, zpx, 4}, {"ROL", rol, zpx, 6}, {"???", xxx, imp, 6}, {"SEC", sec, imp, 2}, {"AND", and, aby, 4}, {"???", nop, imp, 2}, {"???", xxx, imp, 7}, {"???", nop, imp, 4}, {"AND", and, abx, 4}, {"ROL", rol, abx, 7}, {"???", xxx, imp, 7}, {"RTI", rti, imp, 6}, {"EOR", eor, izx, 6}, {"???", xxx, imp, 2}, {"???", xxx, imp, 8}, {"???", nop, imp, 3}, {"EOR", eor, zp0, 3}, {"LSR", lsr, zp0, 5}, {"???", xxx, imp, 5}, {"PHA", pha, imp, 3}, {"EOR", eor, imm, 2}, {"LSR", lsr, imp, 2}, {"???", xxx, imp, 2}, {"JMP", jmp, abs, 3}, {"EOR", eor, abs, 4}, {"LSR", lsr, abs, 6}, {"???", xxx, imp, 6}, {"BVC", bvc, rel, 2}, {"EOR", eor, izy, 5}, {"???", xxx, imp, 2}, {"???", xxx, imp, 8}, {"???", nop, imp, 4}, {"EOR", eor, zpx, 4}, {"LSR", lsr, zpx, 6}, {"???", xxx, imp, 6}, {"CLI", cli, imp, 2}, {"EOR", eor, aby, 4}, {"???", nop, imp, 2}, {"???", xxx, imp, 7}, {"???", nop, imp, 4}, {"EOR", eor, abx, 4}, {"LSR", lsr, abx, 7}, {"???", xxx, imp, 7}, {"RTS", rts, imp, 6}, {"ADC", adc, izx, 6}, {"???", xxx, imp, 2}, {"???", xxx, imp, 8}, {"???", nop, imp, 3}, {"ADC", adc, zp0, 3}, {"ROR", ror, zp0, 5}, {"???", xxx, imp, 5}, {"PLA", pla, imp, 4}, {"ADC", adc, imm, 2}, {"ROR", ror, imp, 2}, {"???", xxx, imp, 2}, {"JMP", jmp, ind, 5}, {"ADC", adc, abs, 4}, {"ROR", ror, abs, 6}, {"???", xxx, imp, 6}, {"BVS", bvs, rel, 2}, {"ADC", adc, izy, 5}, {"???", xxx, imp, 2}, {"???", xxx, imp, 8}, {"???", nop, imp, 4}, {"ADC", adc, zpx, 4}, {"ROR", ror, zpx, 6}, {"???", xxx, imp, 6}, {"SEI", sei, imp, 2}, {"ADC", adc, aby, 4}, {"???", nop, imp, 2}, {"???", xxx, imp, 7}, {"???", nop, imp, 4}, {"ADC", adc, abx, 4}, {"ROR", ror, abx, 7}, {"???", xxx, imp, 7}, {"???", nop, imp, 2}, {"STA", sta, izx, 6}, {"???", nop, imp, 2}, {"???", xxx, imp, 6}, {"STY", sty, zp0, 3}, {"STA", sta, zp0, 3}, {"STX", stx, zp0, 3}, {"???", xxx, imp, 3}, {"DEY", dey, imp, 2}, {"???", nop, imp, 2}, {"TXA", txa, imp, 2}, {"???", xxx, imp, 2}, {"STY", sty, abs, 4}, {"STA", sta, abs, 4}, {"STX", stx, abs, 4}, {"???", xxx, imp, 4}, {"BCC", bcc, rel, 2}, {"STA", sta, izy, 6}, {"???", xxx, imp, 2}, {"???", xxx, imp, 6}, {"STY", sty, zpx, 4}, {"STA", sta, zpx, 4}, {"STX", stx, zpy, 4}, {"???", xxx, imp, 4}, {"TYA", tya, imp, 2}, {"STA", sta, aby, 5}, {"TXS", txs, imp, 2}, {"???", xxx, imp, 5}, {"???", nop, imp, 5}, {"STA", sta, abx, 5}, {"???", xxx, imp, 5}, {"???", xxx, imp, 5}, {"LDY", ldy, imm, 2}, {"LDA", lda, izx, 6}, {"LDX", ldx, imm, 2}, {"???", xxx, imp, 6}, {"LDY", ldy, zp0, 3}, {"LDA", lda, zp0, 3}, {"LDX", ldx, zp0, 3}, {"???", xxx, imp, 3}, {"TAY", tay, imp, 2}, {"LDA", lda, imm, 2}, {"TAX", tax, imp, 2}, {"???", xxx, imp, 2}, {"LDY", ldy, abs, 4}, {"LDA", lda, abs, 4}, {"LDX", ldx, abs, 4}, {"???", xxx, imp, 4}, {"BCS", bcs, rel, 2}, {"LDA", lda, izy, 5}, {"???", xxx, imp, 2}, {"???", xxx, imp, 5}, {"LDY", ldy, zpx, 4}, {"LDA", lda, zpx, 4}, {"LDX", ldx, zpy, 4}, {"???", xxx, imp, 4}, {"CLV", clv, imp, 2}, {"LDA", lda, aby, 4}, {"TSX", tsx, imp, 2}, {"???", xxx, imp, 4}, {"LDY", ldy, abx, 4}, {"LDA", lda, abx, 4}, {"LDX", ldx, aby, 4}, {"???", xxx, imp, 4}, {"CPY", cpy, imm, 2}, {"CMP", cmp, izx, 6}, {"???", nop, imp, 2}, {"???", xxx, imp, 8}, {"CPY", cpy, zp0, 3}, {"CMP", cmp, zp0, 3}, {"DEC", dec, zp0, 5}, {"???", xxx, imp, 5}, {"INY", iny, imp, 2}, {"CMP", cmp, imm, 2}, {"DEX", dex, imp, 2}, {"???", xxx, imp, 2}, {"CPY", cpy, abs, 4}, {"CMP", cmp, abs, 4}, {"DEC", dec, abs, 6}, {"???", xxx, imp, 6}, {"BNE", bne, rel, 2}, {"CMP", cmp, izy, 5}, {"???", xxx, imp, 2}, {"???", xxx, imp, 8}, {"???", nop, imp, 4}, {"CMP", cmp, zpx, 4}, {"DEC", dec, zpx, 6}, {"???", xxx, imp, 6}, {"CLD", cld, imp, 2}, {"CMP", cmp, aby, 4}, {"NOP", nop, imp, 2}, {"???", xxx, imp, 7}, {"???", nop, imp, 4}, {"CMP", cmp, abx, 4}, {"DEC", dec, abx, 7}, {"???", xxx, imp, 7}, {"CPX", cpx, imm, 2}, {"SBC", sbc, izx, 6}, {"???", nop, imp, 2}, {"???", xxx, imp, 8}, {"CPX", cpx, zp0, 3}, {"SBC", sbc, zp0, 3}, {"INC", inc, zp0, 5}, {"???", xxx, imp, 5}, {"INX", inx, imp, 2}, {"SBC", sbc, imm, 2}, {"NOP", nop, imp, 2}, {"???", sbc, imp, 2}, {"CPX", cpx, abs, 4}, {"SBC", sbc, abs, 4}, {"INC", inc, abs, 6}, {"???", xxx, imp, 6}, {"BEQ", beq, rel, 2}, {"SBC", sbc, izy, 5}, {"???", xxx, imp, 2}, {"???", xxx, imp, 8}, {"???", nop, imp, 4}, {"SBC", sbc, zpx, 4}, {"INC", inc, zpx, 6}, {"???", xxx, imp, 6}, {"SED", sed, imp, 2}, {"SBC", sbc, aby, 4}, {"NOP", nop, imp, 2}, {"???", xxx, imp, 7}, {"???", nop, imp, 4}, {"SBC", sbc, abx, 4}, {"INC", inc, abx, 7}, {"???", xxx, imp, 7},
	}
}

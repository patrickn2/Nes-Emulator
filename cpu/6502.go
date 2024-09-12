package cpu

import (
	"reflect"

	"github.com/patrickn2/gonesemulator/bus"
)

type instruction struct {
	name     string
	operate  func() bool
	addrMode func() bool
	cycles   byte
}

type flags struct {
	c uint8
	z uint8
	i uint8
	d uint8
	b uint8
	u uint8
	v uint8
	n uint8
}

type cpu struct {
	bus            *bus.BUS
	status         byte
	accumulator    byte
	xRegister      byte
	yRegister      byte
	stackPointer   byte
	programCounter uint16
	lookup         [256]instruction
	fetched        byte
	addrAbs        uint16
	addrRel        uint16
	opcode         byte
	cycles         byte
	flags          flags
}

func New(bus *bus.BUS) *cpu {
	c := &cpu{
		bus:            bus,
		status:         0x00,
		accumulator:    0x00,
		xRegister:      0x00,
		yRegister:      0x00,
		stackPointer:   0x00,
		programCounter: 0x00,
		fetched:        0x00,
		addrAbs:        0x00,
		addrRel:        0x00,
		opcode:         0x00,
		cycles:         0x00,
		flags: flags{
			c: 1 << 0, // Carry Bit
			z: 1 << 1, // Zero
			i: 1 << 2, // Disable Interrupts
			d: 1 << 3, // Decimal Mode (unused in this implementation)
			b: 1 << 4, // Break
			u: 1 << 5, // Unused
			v: 1 << 6, // Overflow
			n: 1 << 7, // Negative
		},
	}
	c.lookup = [256]instruction{
		{"BRK", c.brk, c.imm, 7},
		{"ORA", c.ora, c.izx, 6},
		{"???", c.xxx, c.imp, 2},
		{"???", c.xxx, c.imp, 8},
		{"???", c.nop, c.imp, 3},
		{"ORA", c.ora, c.zp0, 3},
		{"ASL", c.asl, c.zp0, 5},
		{"???", c.xxx, c.imp, 5},
		{"PHP", c.php, c.imp, 3},
		{"ORA", c.ora, c.imm, 2},
		{"ASL", c.asl, c.imp, 2},
		{"???", c.xxx, c.imp, 2},
		{"???", c.nop, c.imp, 4},
		{"ORA", c.ora, c.abs, 4},
		{"ASL", c.asl, c.abs, 6},
		{"???", c.xxx, c.imp, 6},
		{"BPL", c.bpl, c.rel, 2},
		{"ORA", c.ora, c.izy, 5},
		{"???", c.xxx, c.imp, 2},
		{"???", c.xxx, c.imp, 8},
		{"???", c.nop, c.imp, 4},
		{"ORA", c.ora, c.zpx, 4},
		{"ASL", c.asl, c.zpx, 6},
		{"???", c.xxx, c.imp, 6},
		{"CLC", c.clc, c.imp, 2},
		{"ORA", c.ora, c.aby, 4},
		{"???", c.nop, c.imp, 2},
		{"???", c.xxx, c.imp, 7},
		{"???", c.nop, c.imp, 4},
		{"ORA", c.ora, c.abx, 4},
		{"ASL", c.asl, c.abx, 7},
		{"???", c.xxx, c.imp, 7},
		{"JSR", c.jsr, c.abs, 6},
		{"AND", c.and, c.izx, 6},
		{"???", c.xxx, c.imp, 2},
		{"???", c.xxx, c.imp, 8},
		{"BIT", c.bit, c.zp0, 3},
		{"AND", c.and, c.zp0, 3},
		{"ROL", c.rol, c.zp0, 5},
		{"???", c.xxx, c.imp, 5},
		{"PLP", c.plp, c.imp, 4},
		{"AND", c.and, c.imm, 2},
		{"ROL", c.rol, c.imp, 2},
		{"???", c.xxx, c.imp, 2},
		{"BIT", c.bit, c.abs, 4},
		{"AND", c.and, c.abs, 4},
		{"ROL", c.rol, c.abs, 6},
		{"???", c.xxx, c.imp, 6},
		{"BMI", c.bmi, c.rel, 2},
		{"AND", c.and, c.izy, 5},
		{"???", c.xxx, c.imp, 2},
		{"???", c.xxx, c.imp, 8},
		{"???", c.nop, c.imp, 4},
		{"AND", c.and, c.zpx, 4},
		{"ROL", c.rol, c.zpx, 6},
		{"???", c.xxx, c.imp, 6},
		{"SEC", c.sec, c.imp, 2},
		{"AND", c.and, c.aby, 4},
		{"???", c.nop, c.imp, 2},
		{"???", c.xxx, c.imp, 7},
		{"???", c.nop, c.imp, 4},
		{"AND", c.and, c.abx, 4},
		{"ROL", c.rol, c.abx, 7},
		{"???", c.xxx, c.imp, 7},
		{"RTI", c.rti, c.imp, 6},
		{"EOR", c.eor, c.izx, 6},
		{"???", c.xxx, c.imp, 2},
		{"???", c.xxx, c.imp, 8},
		{"???", c.nop, c.imp, 3},
		{"EOR", c.eor, c.zp0, 3},
		{"LSR", c.lsr, c.zp0, 5},
		{"???", c.xxx, c.imp, 5},
		{"PHA", c.pha, c.imp, 3},
		{"EOR", c.eor, c.imm, 2},
		{"LSR", c.lsr, c.imp, 2},
		{"???", c.xxx, c.imp, 2},
		{"JMP", c.jmp, c.abs, 3},
		{"EOR", c.eor, c.abs, 4},
		{"LSR", c.lsr, c.abs, 6},
		{"???", c.xxx, c.imp, 6},
		{"BVC", c.bvc, c.rel, 2},
		{"EOR", c.eor, c.izy, 5},
		{"???", c.xxx, c.imp, 2},
		{"???", c.xxx, c.imp, 8},
		{"???", c.nop, c.imp, 4},
		{"EOR", c.eor, c.zpx, 4},
		{"LSR", c.lsr, c.zpx, 6},
		{"???", c.xxx, c.imp, 6},
		{"CLI", c.cli, c.imp, 2},
		{"EOR", c.eor, c.aby, 4},
		{"???", c.nop, c.imp, 2},
		{"???", c.xxx, c.imp, 7},
		{"???", c.nop, c.imp, 4},
		{"EOR", c.eor, c.abx, 4},
		{"LSR", c.lsr, c.abx, 7},
		{"???", c.xxx, c.imp, 7},
		{"RTS", c.rts, c.imp, 6},
		{"ADC", c.adc, c.izx, 6},
		{"???", c.xxx, c.imp, 2},
		{"???", c.xxx, c.imp, 8},
		{"???", c.nop, c.imp, 3},
		{"ADC", c.adc, c.zp0, 3},
		{"ROR", c.ror, c.zp0, 5},
		{"???", c.xxx, c.imp, 5},
		{"PLA", c.pla, c.imp, 4},
		{"ADC", c.adc, c.imm, 2},
		{"ROR", c.ror, c.imp, 2},
		{"???", c.xxx, c.imp, 2},
		{"JMP", c.jmp, c.ind, 5},
		{"ADC", c.adc, c.abs, 4},
		{"ROR", c.ror, c.abs, 6},
		{"???", c.xxx, c.imp, 6},
		{"BVS", c.bvs, c.rel, 2},
		{"ADC", c.adc, c.izy, 5},
		{"???", c.xxx, c.imp, 2},
		{"???", c.xxx, c.imp, 8},
		{"???", c.nop, c.imp, 4},
		{"ADC", c.adc, c.zpx, 4},
		{"ROR", c.ror, c.zpx, 6},
		{"???", c.xxx, c.imp, 6},
		{"SEI", c.sei, c.imp, 2},
		{"ADC", c.adc, c.aby, 4},
		{"???", c.nop, c.imp, 2},
		{"???", c.xxx, c.imp, 7},
		{"???", c.nop, c.imp, 4},
		{"ADC", c.adc, c.abx, 4},
		{"ROR", c.ror, c.abx, 7},
		{"???", c.xxx, c.imp, 7},
		{"???", c.nop, c.imp, 2},
		{"STA", c.sta, c.izx, 6},
		{"???", c.nop, c.imp, 2},
		{"???", c.xxx, c.imp, 6},
		{"STY", c.sty, c.zp0, 3},
		{"STA", c.sta, c.zp0, 3},
		{"STX", c.stx, c.zp0, 3},
		{"???", c.xxx, c.imp, 3},
		{"DEY", c.dey, c.imp, 2},
		{"???", c.nop, c.imp, 2},
		{"TXA", c.txa, c.imp, 2},
		{"???", c.xxx, c.imp, 2},
		{"STY", c.sty, c.abs, 4},
		{"STA", c.sta, c.abs, 4},
		{"STX", c.stx, c.abs, 4},
		{"???", c.xxx, c.imp, 4},
		{"BCC", c.bcc, c.rel, 2},
		{"STA", c.sta, c.izy, 6},
		{"???", c.xxx, c.imp, 2},
		{"???", c.xxx, c.imp, 6},
		{"STY", c.sty, c.zpx, 4},
		{"STA", c.sta, c.zpx, 4},
		{"STX", c.stx, c.zpy, 4},
		{"???", c.xxx, c.imp, 4},
		{"TYA", c.tya, c.imp, 2},
		{"STA", c.sta, c.aby, 5},
		{"TXS", c.txs, c.imp, 2},
		{"???", c.xxx, c.imp, 5},
		{"???", c.nop, c.imp, 5},
		{"STA", c.sta, c.abx, 5},
		{"???", c.xxx, c.imp, 5},
		{"???", c.xxx, c.imp, 5},
		{"LDY", c.ldy, c.imm, 2},
		{"LDA", c.lda, c.izx, 6},
		{"LDX", c.ldx, c.imm, 2},
		{"???", c.xxx, c.imp, 6},
		{"LDY", c.ldy, c.zp0, 3},
		{"LDA", c.lda, c.zp0, 3},
		{"LDX", c.ldx, c.zp0, 3},
		{"???", c.xxx, c.imp, 3},
		{"TAY", c.tay, c.imp, 2},
		{"LDA", c.lda, c.imm, 2},
		{"TAX", c.tax, c.imp, 2},
		{"???", c.xxx, c.imp, 2},
		{"LDY", c.ldy, c.abs, 4},
		{"LDA", c.lda, c.abs, 4},
		{"LDX", c.ldx, c.abs, 4},
		{"???", c.xxx, c.imp, 4},
		{"BCS", c.bcs, c.rel, 2},
		{"LDA", c.lda, c.izy, 5},
		{"???", c.xxx, c.imp, 2},
		{"???", c.xxx, c.imp, 5},
		{"LDY", c.ldy, c.zpx, 4},
		{"LDA", c.lda, c.zpx, 4},
		{"LDX", c.ldx, c.zpy, 4},
		{"???", c.xxx, c.imp, 4},
		{"CLV", c.clv, c.imp, 2},
		{"LDA", c.lda, c.aby, 4},
		{"TSX", c.tsx, c.imp, 2},
		{"???", c.xxx, c.imp, 4},
		{"LDY", c.ldy, c.abx, 4},
		{"LDA", c.lda, c.abx, 4},
		{"LDX", c.ldx, c.aby, 4},
		{"???", c.xxx, c.imp, 4},
		{"CPY", c.cpy, c.imm, 2},
		{"CMP", c.cmp, c.izx, 6},
		{"???", c.nop, c.imp, 2},
		{"???", c.xxx, c.imp, 8},
		{"CPY", c.cpy, c.zp0, 3},
		{"CMP", c.cmp, c.zp0, 3},
		{"DEC", c.dec, c.zp0, 5},
		{"???", c.xxx, c.imp, 5},
		{"INY", c.iny, c.imp, 2},
		{"CMP", c.cmp, c.imm, 2},
		{"DEX", c.dex, c.imp, 2},
		{"???", c.xxx, c.imp, 2},
		{"CPY", c.cpy, c.abs, 4},
		{"CMP", c.cmp, c.abs, 4},
		{"DEC", c.dec, c.abs, 6},
		{"???", c.xxx, c.imp, 6},
		{"BNE", c.bne, c.rel, 2},
		{"CMP", c.cmp, c.izy, 5},
		{"???", c.xxx, c.imp, 2},
		{"???", c.xxx, c.imp, 8},
		{"???", c.nop, c.imp, 4},
		{"CMP", c.cmp, c.zpx, 4},
		{"DEC", c.dec, c.zpx, 6},
		{"???", c.xxx, c.imp, 6},
		{"CLD", c.cld, c.imp, 2},
		{"CMP", c.cmp, c.aby, 4},
		{"NOP", c.nop, c.imp, 2},
		{"???", c.xxx, c.imp, 7},
		{"???", c.nop, c.imp, 4},
		{"CMP", c.cmp, c.abx, 4},
		{"DEC", c.dec, c.abx, 7},
		{"???", c.xxx, c.imp, 7},
		{"CPX", c.cpx, c.imm, 2},
		{"SBC", c.sbc, c.izx, 6},
		{"???", c.nop, c.imp, 2},
		{"???", c.xxx, c.imp, 8},
		{"CPX", c.cpx, c.zp0, 3},
		{"SBC", c.sbc, c.zp0, 3},
		{"INC", c.inc, c.zp0, 5},
		{"???", c.xxx, c.imp, 5},
		{"INX", c.inx, c.imp, 2},
		{"SBC", c.sbc, c.imm, 2},
		{"NOP", c.nop, c.imp, 2},
		{"???", c.sbc, c.imp, 2},
		{"CPX", c.cpx, c.abs, 4},
		{"SBC", c.sbc, c.abs, 4},
		{"INC", c.inc, c.abs, 6},
		{"???", c.xxx, c.imp, 6},
		{"BEQ", c.beq, c.rel, 2},
		{"SBC", c.sbc, c.izy, 5},
		{"???", c.xxx, c.imp, 2},
		{"???", c.xxx, c.imp, 8},
		{"???", c.nop, c.imp, 4},
		{"SBC", c.sbc, c.zpx, 4},
		{"INC", c.inc, c.zpx, 6},
		{"???", c.xxx, c.imp, 6},
		{"SED", c.sed, c.imp, 2},
		{"SBC", c.sbc, c.aby, 4},
		{"NOP", c.nop, c.imp, 2},
		{"???", c.xxx, c.imp, 7},
		{"???", c.nop, c.imp, 4},
		{"SBC", c.sbc, c.abx, 4},
		{"INC", c.inc, c.abx, 7},
		{"???", c.xxx, c.imp, 7},
	}
	return c
}

// Public Methods

func (c *cpu) Clock() {
	if c.cycles == 0 {
		c.opcode = c.read(c.programCounter)
		c.programCounter++
		c.cycles = c.lookup[c.opcode].cycles
		addrMode := c.lookup[c.opcode].addrMode()
		operate := c.lookup[c.opcode].operate()
		if addrMode && operate {
			c.cycles++
		}
	}
	c.cycles--
}

// Private Methods

func (c *cpu) setFlag(flag uint8, value bool) {
	if value {
		c.status |= flag
	} else {
		c.status &= ^flag
	}
}

func (c *cpu) getFlag(flag uint8) uint8 {
	if (c.status & flag) > 0 {
		return 1
	}
	return 0
}

func (c *cpu) read(addr uint16) byte {
	return c.bus.Read(addr)
}

func (c *cpu) write(addr uint16, data byte) {
	c.bus.Write(addr, data)
}

func (c *cpu) imp() bool {
	c.fetched = c.accumulator
	return false
}

func (c *cpu) imm() bool {
	c.addrAbs = c.programCounter
	c.programCounter++
	return false
}

func (c *cpu) zp0() bool {
	c.addrAbs = uint16(c.read(c.programCounter))
	c.programCounter++
	c.addrAbs &= 0x00FF
	return false
}

func (c *cpu) zpx() bool {
	c.addrAbs = uint16(c.read(c.programCounter) + c.xRegister)
	c.programCounter++
	c.addrAbs &= 0x00FF
	return false
}

func (c *cpu) zpy() bool {
	c.addrAbs = uint16(c.read(c.programCounter) + c.yRegister)
	c.programCounter++
	c.addrAbs &= 0x00FF
	return false
}

func (c *cpu) abs() bool {
	lo := c.read(c.programCounter)
	c.programCounter++
	hi := c.read(c.programCounter)
	c.programCounter++
	c.addrAbs = uint16(hi)<<8 | uint16(lo)
	return false
}

func (c *cpu) abx() bool {
	lo := c.read(c.programCounter)
	c.programCounter++
	hi := c.read(c.programCounter)
	c.programCounter++
	c.addrAbs = uint16(hi)<<8 | uint16(lo)
	c.addrAbs += uint16(c.xRegister)
	return uint16(c.addrAbs&0xFF00) != uint16(hi)<<8
}

func (c *cpu) aby() bool {
	lo := c.read(c.programCounter)
	c.programCounter++
	hi := c.read(c.programCounter)
	c.programCounter++
	c.addrAbs = uint16(hi)<<8 | uint16(lo)
	c.addrAbs += uint16(c.yRegister)
	return uint16(c.addrAbs&0xFF00) != uint16(hi)<<8
}

func (c *cpu) ind() bool {
	lo := c.read(c.programCounter)
	c.programCounter++
	hi := c.read(c.programCounter)
	c.programCounter++

	ptr := uint16(hi)<<8 | uint16(lo)

	if lo == 0x00FF {
		c.addrAbs = uint16(c.read(ptr&0xFF00))<<8 | uint16(c.read(ptr))
	} else {
		c.addrAbs = uint16(c.read(ptr+1))<<8 | uint16(c.read(ptr))
	}

	return false
}

// func (c *cpu) idx() bool {
// 	t := c.read(c.programCounter)
// 	c.programCounter++

// 	lo := c.read(uint16(t+c.xRegister) & 0x00FF)
// 	hi := c.read(uint16(t+c.xRegister+1) & 0x00FF)
// 	c.addrAbs = uint16(hi)<<8 | uint16(lo)
// 	return false
// }

// func (c *cpu) idy() bool {
// 	t := c.read(c.programCounter)
// 	c.programCounter++
// 	lo := c.read(uint16(t) & 0x00FF)
// 	hi := c.read(uint16(t+1) & 0x00FF)
// 	c.addrAbs = uint16(hi)<<8 | uint16(lo)
// 	c.addrAbs += uint16(c.yRegister)
// 	return (c.addrAbs & 0xFF00) != uint16(hi)<<8
// }

func (c *cpu) rel() bool {
	c.addrRel = uint16((c.programCounter))
	c.programCounter++
	if (c.addrRel & 0x80) != 0 {
		c.addrRel |= 0xFF00
	}
	return false
}

// Instructions

func (c *cpu) fetch() byte {
	if reflect.ValueOf(c.lookup[c.opcode].addrMode).Pointer() == reflect.ValueOf(c.imp).Pointer() {
		c.fetched = c.read(c.addrAbs)
	}
	return c.fetched
}

func (c *cpu) and() bool {
	c.fetch()
	c.accumulator = c.accumulator & c.fetched
	c.setFlag(c.flags.z, c.accumulator == 0x00)
	c.setFlag(c.flags.n, (c.accumulator&0x80) != 0)
	return true
}

func (c *cpu) asl() bool {
	c.fetch()
	temp := c.fetched << 1
	c.setFlag(c.flags.c, (uint16(temp)&0xFF00) > 0)
	c.setFlag(c.flags.z, (temp&0x00FF) == 0)
	c.setFlag(c.flags.n, (temp&0x80) != 0)
	if reflect.ValueOf(c.lookup[c.opcode].addrMode).Pointer() == reflect.ValueOf(c.imp).Pointer() {
		c.accumulator = temp & 0x00FF
	} else {
		c.write(c.addrAbs, temp&0x00FF)
	}
	return false
}

func (c *cpu) bcs() bool {
	if c.getFlag(c.flags.c) == 1 {
		c.cycles++
		c.addrAbs = c.programCounter + c.addrRel
		if (c.addrAbs & 0xFF00) != (c.programCounter & 0xFF00) {
			c.cycles++
		}
		c.programCounter = c.addrAbs
	}
	return false
}

func (c *cpu) bcc() bool {
	if c.getFlag(c.flags.c) == 0 {
		c.cycles++
		c.addrAbs = c.programCounter + c.addrRel
		if (c.addrAbs & 0xFF00) != (c.programCounter & 0xFF00) {
			c.cycles++
		}
		c.programCounter = c.addrAbs
	}
	return false
}

func (c *cpu) beq() bool {
	if c.getFlag(c.flags.z) == 1 {
		c.cycles++
		c.addrAbs = c.programCounter + c.addrRel
		if (c.addrAbs & 0xFF00) != (c.programCounter & 0xFF00) {
			c.cycles++
		}
		c.programCounter = c.addrAbs
	}
	return false
}

func (c *cpu) bit() bool {
	c.fetch()
	temp := c.accumulator & c.fetched
	c.setFlag(c.flags.z, (temp&0x00FF) == 0)
	c.setFlag(c.flags.n, c.fetched&(1<<7) > 0)
	c.setFlag(c.flags.v, c.fetched&(1<<6) > 0)
	return false
}

func (c *cpu) bmi() bool {
	if c.getFlag(c.flags.n) == 1 {
		c.cycles++
		c.addrAbs = c.programCounter + c.addrRel
		if (c.addrAbs & 0xFF00) != (c.programCounter & 0xFF00) {
			c.cycles++
		}
		c.programCounter = c.addrAbs
	}
	return false
}

func (c *cpu) bne() bool {
	if c.getFlag(c.flags.z) == 0 {
		c.cycles++
		c.addrAbs = c.programCounter + c.addrRel
		if (c.addrAbs & 0xFF00) != (c.programCounter & 0xFF00) {
			c.cycles++
		}
		c.programCounter = c.addrAbs
	}
	return false
}

func (c *cpu) bpl() bool {
	if c.getFlag(c.flags.n) == 0 {
		c.cycles++
		c.addrAbs = c.programCounter + c.addrRel
		if (c.addrAbs & 0xFF00) != (c.programCounter & 0xFF00) {
			c.cycles++
		}
		c.programCounter = c.addrAbs
	}
	return false
}

func (c *cpu) bvc() bool {
	if c.getFlag(c.flags.z) == 0 {
		c.cycles++
		c.addrAbs = c.programCounter + c.addrRel
		if (c.addrAbs & 0xFF00) != (c.programCounter & 0xFF00) {
			c.cycles++
		}
		c.programCounter = c.addrAbs
	}
	return false
}

func (c *cpu) bvs() bool {
	if c.getFlag(c.flags.v) == 1 {
		c.cycles++
		c.addrAbs = c.programCounter + c.addrRel
		if (c.addrAbs & 0xFF00) != (c.programCounter & 0xFF00) {
			c.cycles++
		}
		c.programCounter = c.addrAbs
	}
	return false
}

func (c *cpu) clc() bool {
	c.setFlag(c.flags.c, false)
	return false
}

func (c *cpu) cld() bool {
	c.setFlag(c.flags.d, false)
	return false
}

func (c *cpu) cli() bool {
	c.setFlag(c.flags.i, false)
	return false
}

func (c *cpu) clv() bool {
	c.setFlag(c.flags.v, false)
	return false
}

func (c *cpu) cmp() bool {
	c.fetch()
	temp := uint16(c.accumulator - c.fetched)
	c.setFlag(c.flags.c, temp < uint16(c.accumulator))
	c.setFlag(c.flags.z, (temp&0x00FF) == 0)
	c.setFlag(c.flags.n, (temp&0x0080) != 0)
	return false
}

func (c *cpu) cpx() bool {
	c.fetch()
	temp := uint16(c.xRegister - c.fetched)
	c.setFlag(c.flags.c, temp < uint16(c.xRegister))
	c.setFlag(c.flags.z, (temp&0x00FF) == 0)
	c.setFlag(c.flags.n, (temp&0x0080) != 0)
	return false
}

func (c *cpu) cpy() bool {
	c.fetch()
	temp := uint16(c.yRegister - c.fetched)
	c.setFlag(c.flags.c, temp < uint16(c.yRegister))
	c.setFlag(c.flags.z, (temp&0x00FF) == 0)
	c.setFlag(c.flags.n, (temp&0x0080) != 0)
	return false
}

func (c *cpu) dec() bool {
	c.fetch()
	temp := c.fetched - 1
	c.write(c.addrAbs, temp&0x00FF)
	c.setFlag(c.flags.z, (temp&0x00FF) == 0)
	c.setFlag(c.flags.n, (temp&0x80) != 0)
	return false
}

func (c *cpu) dex() bool {
	c.xRegister--
	c.setFlag(c.flags.z, (c.xRegister&0x00FF) == 0)
	c.setFlag(c.flags.n, (c.xRegister&0x80) != 0)
	return false
}

func (c *cpu) dey() bool {
	c.yRegister--
	c.setFlag(c.flags.z, (c.yRegister&0x00FF) == 0)
	c.setFlag(c.flags.n, (c.yRegister&0x80) != 0)
	return false
}

func (c *cpu) eor() bool {
	c.fetch()
	c.accumulator = c.accumulator ^ c.fetched
	c.setFlag(c.flags.z, c.accumulator == 0x00)
	c.setFlag(c.flags.n, (c.accumulator&0x80) != 0)
	return true
}

func (c *cpu) inc() bool {
	c.fetch()
	temp := c.fetched + 1
	c.write(c.addrAbs, temp&0x00FF)
	c.setFlag(c.flags.z, (temp&0x00FF) == 0)
	c.setFlag(c.flags.n, (temp&0x80) != 0)
	return false
}

func (c *cpu) inx() bool {
	c.xRegister++
	c.setFlag(c.flags.z, (c.xRegister&0x00FF) == 0)
	c.setFlag(c.flags.n, (c.xRegister&0x80) != 0)
	return false
}

func (c *cpu) iny() bool {
	c.yRegister++
	c.setFlag(c.flags.z, (c.yRegister&0x00FF) == 0)
	c.setFlag(c.flags.n, (c.yRegister&0x80) != 0)
	return false
}

func (c *cpu) jmp() bool {
	c.programCounter = c.addrAbs
	return false
}

func (c *cpu) jsr() bool {
	c.programCounter--

	c.write(0x0100+uint16(c.stackPointer), byte((c.programCounter>>8)&0x00FF))
	c.stackPointer--
	c.write(0x0100+uint16(c.stackPointer), byte(c.programCounter&0x00FF))
	c.stackPointer--
	c.programCounter = c.addrAbs
	return false
}

func (c *cpu) lda() bool {
	c.fetch()
	c.accumulator = c.fetched
	c.setFlag(c.flags.z, c.accumulator == 0x00)
	c.setFlag(c.flags.n, (c.accumulator&0x80) != 0)
	return false
}

func (c *cpu) ldx() bool {
	c.fetch()
	c.xRegister = c.fetched
	c.setFlag(c.flags.z, c.xRegister == 0x00)
	c.setFlag(c.flags.n, (c.xRegister&0x80) != 0)
	return false
}

func (c *cpu) ldy() bool {
	c.fetch()
	c.yRegister = c.fetched
	c.setFlag(c.flags.z, c.yRegister == 0x00)
	c.setFlag(c.flags.n, (c.yRegister&0x80) != 0)
	return false
}

func (c *cpu) lsr() bool {
	c.fetch()
	c.setFlag(c.flags.c, (c.fetched&0x01) > 0)
	temp := c.fetched >> 1
	c.setFlag(c.flags.z, (temp&0x00FF) == 0)
	c.setFlag(c.flags.n, (temp&0x80) != 0)
	if reflect.ValueOf(c.lookup[c.opcode].addrMode).Pointer() == reflect.ValueOf(c.imp).Pointer() {
		c.accumulator = temp & 0x00FF
	} else {
		c.write(c.addrAbs, temp&0x00FF)
	}
	return false
}

func (c *cpu) nop() bool {
	switch c.opcode {
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

func (c *cpu) adc() bool {
	c.fetch()
	temp := uint16(c.accumulator) + uint16(c.fetched) + uint16(c.getFlag(c.flags.c))
	c.setFlag(c.flags.c, temp > 255)
	c.setFlag(c.flags.z, (temp&0x00FF) == 0)
	c.setFlag(c.flags.n, (temp&0x80) != 0)
	c.setFlag(c.flags.v, ((c.accumulator^c.fetched)&(c.accumulator^byte(temp))&0x80) != 0)
	c.accumulator = byte(temp & 0x00FF)
	return true
}

func (c *cpu) sbc() bool {
	c.fetch()
	value := uint16(c.fetched) ^ 0x00FF
	temp := uint16(c.accumulator) + value + uint16(c.getFlag(c.flags.c))
	c.setFlag(c.flags.c, temp&0xFF00 != 0)
	c.setFlag(c.flags.z, (temp&0x00FF) == 0)
	c.setFlag(c.flags.n, (temp&0x80) != 0)
	c.setFlag(c.flags.v, (temp^(uint16(c.accumulator)^value))&0x80 != 0)
	c.accumulator = byte(temp & 0x00FF)
	return true
}

func (c *cpu) pha() bool {
	c.write(0x0100+uint16(c.stackPointer), c.accumulator)
	c.stackPointer--
	return false
}

func (c *cpu) php() bool {
	c.write(0x0100+uint16(c.stackPointer), c.status|c.flags.b|c.flags.u)
	c.setFlag(c.flags.b, false)
	c.setFlag(c.flags.u, false)
	c.stackPointer--
	return false
}

func (c *cpu) pla() bool {
	c.stackPointer++
	c.accumulator = c.read(0x0100 + uint16(c.stackPointer))
	c.setFlag(c.flags.z, c.accumulator == 0x00)
	c.setFlag(c.flags.n, (c.accumulator&0x80) != 0)
	return false
}

func (c *cpu) plp() bool {
	c.stackPointer++
	c.status = c.read(0x0100 + uint16(c.stackPointer))
	c.setFlag(c.flags.u, true)
	return false
}

func (c *cpu) rol() bool {
	c.fetch()
	temp := (c.fetched << 1) | c.getFlag(c.flags.c)
	c.setFlag(c.flags.c, (uint16(temp)&0xFF00) > 0)
	c.setFlag(c.flags.z, (temp&0x00FF) == 0)
	c.setFlag(c.flags.n, (temp&0x0080) != 0)
	if reflect.ValueOf(c.lookup[c.opcode].addrMode).Pointer() == reflect.ValueOf(c.imp).Pointer() {
		c.accumulator = temp & 0x00FF
	} else {
		c.write(c.addrAbs, temp&0x00FF)
	}
	return false
}

func (c *cpu) ror() bool {
	c.fetch()
	temp := (c.getFlag(c.flags.c) << 7) | (c.fetched >> 1)
	c.setFlag(c.flags.c, (uint16(temp)&0x01) != 0)
	c.setFlag(c.flags.z, (temp&0x00FF) == 0)
	c.setFlag(c.flags.n, (temp&0x80) != 0)
	if reflect.ValueOf(c.lookup[c.opcode].addrMode).Pointer() == reflect.ValueOf(c.imp).Pointer() {
		c.accumulator = temp & 0x00FF
	} else {
		c.write(c.addrAbs, temp&0x00FF)
	}
	return false
}

// func (c *cpu) irq() {
// 	if c.getFlag(c.flags.i) == 0 {
// 		c.write(0x0100+uint16(c.stackPointer), byte((c.programCounter>>8)&0x00FF))
// 		c.stackPointer--
// 		c.write(0x0100+uint16(c.stackPointer), byte(c.programCounter&0x00FF))
// 		c.stackPointer--

// 		c.setFlag(c.flags.b, false)
// 		c.setFlag(c.flags.u, true)
// 		c.setFlag(c.flags.i, true)
// 		c.write(0x0100+uint16(c.stackPointer), c.status)
// 		c.stackPointer--
// 		c.addrAbs = 0xFFFE
// 		lo := c.read(c.addrAbs + 0)
// 		hi := c.read(c.addrAbs + 1)
// 		c.programCounter = (uint16(hi) << 8) | uint16(lo)
// 		c.cycles = 7
// 	}
// }

// func (c *cpu) nmi() {
// 	c.write(0x0100+uint16(c.stackPointer), byte((c.programCounter>>8)&0x00FF))
// 	c.stackPointer--
// 	c.write(0x0100+uint16(c.stackPointer), byte(c.programCounter&0x00FF))
// 	c.stackPointer--

// 	c.setFlag(c.flags.b, false)
// 	c.setFlag(c.flags.u, true)
// 	c.setFlag(c.flags.i, true)
// 	c.write(0x0100+uint16(c.stackPointer), c.status)
// 	c.stackPointer--
// 	c.addrAbs = 0xFFFA
// 	lo := c.read(c.addrAbs + 0)
// 	hi := c.read(c.addrAbs + 1)
// 	c.programCounter = (uint16(hi) << 8) | uint16(lo)
// 	c.cycles = 8
// }

func (c *cpu) rti() bool {
	c.stackPointer++
	c.status = c.read(0x0100 + uint16(c.stackPointer))
	c.status &= ^c.flags.b
	c.status &= ^c.flags.u
	c.stackPointer++
	c.programCounter = uint16(c.read(0x0100 + uint16(c.stackPointer)))
	c.stackPointer++
	c.programCounter |= uint16(c.read(0x0100+uint16(c.stackPointer))) << 8
	return false
}

func (c *cpu) rts() bool {
	c.stackPointer++
	c.programCounter = uint16(c.read(0x0100 + uint16(c.stackPointer)))
	c.stackPointer++
	c.programCounter |= uint16(c.read(0x0100+uint16(c.stackPointer))) << 8
	c.programCounter++
	return false
}

func (c *cpu) sec() bool {
	c.setFlag(c.flags.c, true)
	return false
}

func (c *cpu) sed() bool {
	c.setFlag(c.flags.d, true)
	return false
}

func (c *cpu) sei() bool {
	c.setFlag(c.flags.i, true)
	return false
}

func (c *cpu) sta() bool {
	c.write(c.addrAbs, c.accumulator)
	return false
}

func (c *cpu) stx() bool {
	c.write(c.addrAbs, c.xRegister)
	return false
}

func (c *cpu) sty() bool {
	c.write(c.addrAbs, c.yRegister)
	return false
}

func (c *cpu) tax() bool {
	c.xRegister = c.accumulator
	c.setFlag(c.flags.z, c.xRegister == 0x00)
	c.setFlag(c.flags.n, (c.xRegister&0x80) != 0)
	return false
}

func (c *cpu) tay() bool {
	c.yRegister = c.accumulator
	c.setFlag(c.flags.z, c.yRegister == 0x00)
	c.setFlag(c.flags.n, (c.yRegister&0x80) != 0)
	return false
}

func (c *cpu) tsx() bool {
	c.xRegister = c.stackPointer
	c.setFlag(c.flags.z, c.xRegister == 0x00)
	c.setFlag(c.flags.n, (c.xRegister&0x80) != 0)
	return false
}

func (c *cpu) txa() bool {
	c.accumulator = c.xRegister
	c.setFlag(c.flags.z, c.accumulator == 0x00)
	c.setFlag(c.flags.n, (c.accumulator&0x80) != 0)
	return false
}

func (c *cpu) txs() bool {
	c.stackPointer = c.xRegister
	return false
}

func (c *cpu) tya() bool {
	c.accumulator = c.yRegister
	c.setFlag(c.flags.z, c.accumulator == 0x00)
	c.setFlag(c.flags.n, (c.accumulator&0x80) != 0)
	return false
}

func (c *cpu) xxx() bool {
	return false
}

func (c *cpu) brk() bool {
	c.programCounter++
	c.setFlag(c.flags.i, true)
	c.write(0x0100+uint16(c.stackPointer), byte((c.programCounter>>8)&0x00FF))
	c.stackPointer--
	c.write(0x0100+uint16(c.stackPointer), byte(c.programCounter&0x00FF))
	c.stackPointer--

	c.setFlag(c.flags.b, true)
	c.write(0x0100+uint16(c.stackPointer), c.status)
	c.stackPointer--
	c.setFlag(c.flags.b, false)

	lo := c.read(0xFFFE)
	hi := c.read(0xFFFF)
	c.programCounter = uint16(hi)<<8 | uint16(lo)
	return false
}

func (c *cpu) ora() bool {
	c.fetch()
	c.accumulator = c.accumulator | c.fetched
	c.setFlag(c.flags.z, c.accumulator == 0x00)
	c.setFlag(c.flags.n, (c.accumulator&0x80) != 0)
	return true
}

func (c *cpu) izx() bool {
	t := uint16(c.read(c.programCounter))
	c.programCounter++
	lo := c.read((t + uint16(c.xRegister)&0x00FF))
	hi := c.read(((t + uint16(c.xRegister) + 1) & 0x00FF))

	c.addrAbs = uint16(hi)<<8 | uint16(lo)
	return false
}

func (c *cpu) izy() bool {
	t := uint16(c.read(c.programCounter))
	c.programCounter++
	lo := c.read(t & 0x00FF)
	hi := c.read((t + 1) & 0x00FF)

	c.addrAbs = uint16(hi)<<8 | uint16(lo)
	c.addrAbs += uint16(c.yRegister)
	return (c.addrAbs & 0xFF00) != uint16(hi)<<8
}

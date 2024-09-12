package cartridge

import (
	"encoding/binary"
	"fmt"
	"os"
)

var nMapperID byte = 0

type sHeader struct {
	Name           [4]byte
	Prg_rom_chunks byte
	Chr_rom_chunks byte
	Mapper1        byte
	Mapper2        byte
	Prg_ram_size   byte
	Tv_system1     byte
	Tv_system2     byte
	Unused         [5]byte
}

type rom struct {
	Header     [16]byte
	Prg_Memory []byte
	Chr_Memory []byte
}

var Rom = rom{}

func ReadCartridge() {
	file, err := os.Open("roms/bartman.nes")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	var header sHeader

	err = binary.Read(file, binary.BigEndian, &header)
	if err != nil {
		fmt.Println(err)
		return
	}

	nMapperID = (header.Mapper2>>4)<<4 | header.Mapper1>>4
	var nFileFormat uint8 = 1

	if nFileFormat == 1 {

		Rom.Prg_Memory = make([]byte, int(header.Prg_rom_chunks)*16*1024)
		Rom.Chr_Memory = make([]byte, int(header.Chr_rom_chunks)*16*1024)

		err = binary.Read(file, binary.BigEndian, &Rom.Prg_Memory)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func CPUWrite(addr uint16, data byte) bool {
	return false
}
func CPURead(addr uint16, bReadOnly bool) bool {
	return false
}

func PPUWrite(addr uint16, data byte) bool {
	return false
}
func PPURead(addr uint16, bReadOnly bool) bool {
	return false
}

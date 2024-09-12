package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"github.com/patrickn2/gonesemulator/bus"
	"github.com/patrickn2/gonesemulator/cpu"
)

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

func main() {

	file, err := os.Open("roms/bartman.nes")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	var header sHeader

	byteHeader := [16]byte{}

	_, err = file.Read(byteHeader[:])
	if err != nil {
		log.Fatalln(err)
	}

	err = binary.Read(bytes.NewReader(byteHeader[:]), binary.BigEndian, &header)
	if err != nil {
		log.Fatalln(err)
	}

	var nMapperID byte = (header.Mapper2>>4)<<4 | header.Mapper1>>4

	prgRom := make([]byte, int(header.Prg_rom_chunks)*16*1024)
	chrRom := make([]byte, int(header.Chr_rom_chunks)*8*1024)
	// var nameTableArrangement byte = header.Mapper1 & 0b00000001 // 1 : Horizontal, 0 : Vertical
	// var batteryBacked bool = header.Mapper1&0b00000010 > 0
	var trainer bool = header.Mapper1&0b00000100 > 0
	// var alternativeNameTable bool = header.Mapper1&0b00001000 > 0

	trainerData := make([]byte, 512)

	if trainer {
		_, err = file.Read(trainerData[:])
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	_, err = file.Read(prgRom[:])
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = file.Read(chrRom[:])
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("nMapperID: %d", nMapperID)
	bus := bus.New()
	cpu := cpu.New(bus)

	cpu.Clock()

}

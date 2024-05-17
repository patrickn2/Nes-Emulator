package main

import (
	"encoding/binary"
	"fmt"
	"os"
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

	err = binary.Read(file, binary.BigEndian, &header)
	if err != nil {
		fmt.Println(err)
		return
	}

	// fmt.Printf("Mapper2: %d, Mapper1: %b", (header.Mapper2>>4)<<4, header.Mapper1)

	fmt.Printf("Prg ROM Chunks: %d", header.Prg_rom_chunks)

}

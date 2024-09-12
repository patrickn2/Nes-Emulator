package cartridge

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

type sHeader struct {
	Name         [4]byte
	PrgRomChunks byte
	ChrRomChunks byte
	Flags6       byte
	Flags7       byte
	PrgRamSize   byte
	TvSystem1    byte
	TvSystem2    byte
	Unused       [5]byte
}

type cartridge struct {
	header    sHeader
	prgMemory []byte
	chrMemory []byte
}

func New(fileLocation string) *cartridge {
	file, err := os.Open(fileLocation)
	if err != nil {
		log.Fatalln("error opening file", err)
	}
	defer file.Close()
	var header sHeader

	byteHeader := make([]byte, 0x10)

	_, err = file.Read(byteHeader)
	if err != nil {
		log.Fatalln("error reading file header", err)
	}

	err = binary.Read(bytes.NewReader(byteHeader), binary.BigEndian, &header)
	if err != nil {
		log.Fatalln("error reading to header struct", err)
	}

	if !bytes.Equal(header.Name[:], []byte{'N', 'E', 'S', 26}) {
		log.Fatalln("invalid file type", string(header.Name[:]))
	}

	mapperID := (header.Flags7>>4)<<4 | header.Flags6>>4
	prgRom := make([]byte, int(header.PrgRomChunks)*16*1024)
	chrRom := make([]byte, int(header.ChrRomChunks)*8*1024)
	nameTableArrangement := header.Flags6 & 0b00000001 // 1 : Horizontal, 0 : Vertical
	batteryBacked := header.Flags6&0b00000010 > 0
	trainer := header.Flags6&0b00000100 > 0
	alternativeNameTable := header.Flags6&0b00001000 > 0
	vsUnisystem := header.Flags7&0b00000001 > 0
	playChoice := header.Flags7&0b00000010 > 0
	flagsIn8_15Nes2 := header.Flags7&0b00001100 == 2

	trainerData := make([]byte, 512)

	if trainer {
		_, err = file.Read(trainerData)
		if err != nil {
			log.Fatalln("error reading trainer", err)
		}
	}

	_, err = file.Read(prgRom)
	if err != nil {
		log.Fatalln("error reading prg Rom", err)
	}
	_, err = file.Read(chrRom)
	if err != nil {
		log.Fatalln("error reading chr Rom", err)
	}

	fmt.Println("NES Cartridge Information")
	fmt.Printf("PRG ROM Chunks: %d, size %d\n", header.PrgRomChunks, int(header.PrgRomChunks)*16*1024)
	fmt.Printf("CHR ROM Chunks: %d, size %d\n", header.ChrRomChunks, int(header.ChrRomChunks)*8*1024)
	fmt.Printf("Nametable arrangement: %d\n", nameTableArrangement)
	fmt.Printf("Battery Backed Cartridge: %t\n", batteryBacked)
	fmt.Printf("Trainer: %t\n", trainer)
	fmt.Printf("Alternative nametable layout:  %t\n", alternativeNameTable)
	fmt.Printf("VS Unisystem: %t\n", vsUnisystem)
	fmt.Printf("PlayChoice-10: %t\n", playChoice)
	fmt.Printf("NES 2.0 format: %t\n", flagsIn8_15Nes2)
	fmt.Printf("Mapped Id: %d\n", mapperID)

	return &cartridge{
		header:    header,
		prgMemory: prgRom,
		chrMemory: chrRom,
	}
}

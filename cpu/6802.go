package cpu

import (
	"github.com/patrickn2/gonesemulator/bus"
)

func Read(addr uint16) byte {
	return bus.Read(addr, false)
}

func Write(addr uint16, data byte) {
	bus.Write(addr, data)
}

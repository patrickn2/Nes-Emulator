package mapper

var nPRGBanks uint8 = 0
var nCHRBanks uint8 = 0

func CPUMapRead(addr uint16, mapped_addr uint32) bool {
	return false
}
func CPUMapWrite(addr uint16, mapped_addr uint32) bool {
	return false
}
func PPUMapRead(addr uint16, mapped_addr uint32) bool {
	return false
}
func PPUMapWrite(addr uint16, mapped_addr uint32) bool {
	return false
}

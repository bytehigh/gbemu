package main

const mem_size = 65536
const rom_start = 0x0100

var gbmmu mmu

type mmu struct {
	memory [mem_size]byte
}

func (gbmmu *mmu) fetchByte(address uint16) byte {
	tstates += 4
	return gbmmu.memory[address]
}

func (gbmmu *mmu) storeByte(address uint16, value byte) {
	tstates += 4
	gbmmu.memory[address] = value
}

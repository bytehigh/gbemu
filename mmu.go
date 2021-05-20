package main

const mem_size = 65536

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

	switch address {
	case 0xFF50:
		// if DMG rom is turned off, copy the first 256 bytes of the ROM into memory
		if value > 0 {
			bootpage := getBootpage()
			for i, opcode := range bootpage {
				gbmmu.memory[i] = opcode
			}
		}
	default:

	}
}

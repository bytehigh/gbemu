package main

const mem_size = 65536

type mmu struct {
	memory [mem_size]byte
}

type cpu struct {
	a, b, c, d, e, h, l, f int8
	pc, sp                 int16
}

func (gbcpu cpu) tick(gbmmu mmu) {
	var opcode byte = gbmmu.memory[gbcpu.pc]

	return
}

func (gbcpu cpu) nop() {
	return
}

func (gbcpu cpu) ld_bc_d16() {
	return
}

func main() {
	gbmmu := mmu{}
	gbcpu := cpu{}

	//load ROM into memory
	//todo

	//game loop
	//todo

	//execute clock cycle
	gbcpu.tick(gbmmu)

}

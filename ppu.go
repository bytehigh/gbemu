package main

type ppu struct {
	LY uint16
}

func (gbppu *ppu) initialise() {
	gbppu.LY = 0xFF44
}

//broken implementation of hblank
func (gbppu *ppu) hblank() {
	gbmmu.memory[gbppu.LY] = gbmmu.memory[gbppu.LY] + 1

	//if LY exceeds max value, reset
	if gbmmu.memory[gbppu.LY] > 0x99 {
		gbmmu.memory[gbppu.LY] = 0
	}
}

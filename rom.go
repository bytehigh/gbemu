package main

import (
	"encoding/hex"
)

type rom struct {
	entry    uint32
	logo     []byte
	title    [16]byte
	man_code [4]byte
	//<todo>
}

func (gbrom *rom) initialise() {
	//initialise array sizes
	gbrom.logo = make([]byte, 16)

	//set up variables
	var nintendo_logo = "CEED6666CC0D000B03730083000C000D0008111F8889000EDCCC6EE6DDDDD999BBBB67636E0EECCCDDDC999FBBB9333E"
	var err error

	//set up rom structure
	gbrom.entry = uint32(0x00C30150)
	gbrom.logo, err = hex.DecodeString(nintendo_logo)
	if err != nil {
		panic(err)
	}
}

func (gbrom *rom) load() {
	var mem_pos = rom_start

	//var bob = make(gbmmu.memory, mem_size)

	mem_pos += 4

	for i := 0; i < 48; i++ {
		gbmmu.memory[mem_pos] = gbrom.logo[i]
		mem_pos += 1
	}
}

package main

const mem_size = 65536
const rom_start = 0x0100

var gbmmu mmu

type mmu struct {
	memory [mem_size]byte
}

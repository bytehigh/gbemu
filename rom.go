package main

import (
	"bufio"
	"encoding/hex"
	"log"
	"os"
)

const rom_start = 0x0100

type rom struct {
	entry    uint32
	logo     []byte
	title    [16]byte
	man_code [4]byte
	//<todo>
}

// store first 256 bytes of ROM until boot.rom page is turned off
var bootpage [0x100]byte

func getBootpage() []byte {
	return bootpage[:]
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

//func check(e error) {
//	if e != nil {
//		panic(e)
//	}
//}

func (gbrom *rom) load() {
	var mem_pos = 0

	// Open file and create scanner on top of it
	file, err := os.Open("Tetris (World).gb")
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)

	// Default scanner is bufio.ScanLines. Lets use ScanWords.
	// Could also use a custom function of SplitFunc type
	scanner.Split(bufio.ScanBytes)

	for scanner.Scan() {
		b := scanner.Bytes()

		if mem_pos < rom_start {
			bootpage[mem_pos] = b[0]
		}

		if mem_pos >= rom_start {
			gbmmu.memory[mem_pos] = b[0]
		}
		mem_pos += 1
	}

	//success := scanner.Bytes()
	//if success == false {
	//	// False on error or EOF. Check error
	//	err = scanner.Err()
	//	if err == nil {
	//		log.Println("Scan completed and reached EOF")
	//	} else {
	//		log.Fatal(err)
	//		check(err)
	//	}
	//}

	//var bob = make(gbmmu.memory, mem_size)

	// original placheholder code to load fake cart logo so boot rom does not crash
	//mem_pos += 4

	//for i := 0; i < 48; i++ {
	//	gbmmu.memory[mem_pos] = gbrom.logo[i]
	//	mem_pos += 1
	//}
}

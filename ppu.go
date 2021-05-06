package main

import (
	"fmt"
	"image/color"
	"strconv"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

var gbscreen *pixel.PictureData
var sLogo string = "f000f000fc00fc00fc00fc00f300f3003c003c003c003c003c003c003c003c00f000f000f000f00000000000f300f300000000000000000000000000cf00cf00000000000f000f003f003f000f000f000000000000000000c000c0000f000f00000000000000000000000000f000f000000000000000000000000000f300f300000000000000000000000000c000c000030003000300030003000300ff00ff00c000c000c000c000c000c000c300c300000000000000000000000000fc00fc00f300f300f000f000f000f000f000f0003c003c00fc00fc00fc00fc003c003c00f300f300f300f300f300f300f300f300f300f300c300c300c300c300c300c300cf00cf00cf00cf00cf00cf00cf00cf003c003c003f003f003c003c000f000f003c003c00fc00fc0000000000fc00fc00fc00fc00f000f000f000f000f000f000f300f300f300f300f300f300f000f000c300c300c300c300c300c300ff00ff00cf00cf00cf00cf00cf00cf00c300c3000f000f000f000f000f000f00fc00fc003c004200b900a500b900a50042003c"

type ppu struct {
	LY          uint16
	tilePattern uint16
}

func (gbppu *ppu) initialise() {
	gbppu.LY = 0xFF44
	gbppu.tilePattern = 0x8000

	gbscreen = pixel.MakePictureData(pixel.R(0, 0, 160, 144))
}

func (gbppu *ppu) showTilePattern(gbscreen *pixel.PictureData, tile uint16) {

	row := uint16(0)
	offset := tile
	//for f := 0; f < len(logo); f += 2 {
	for f := uint16(0); f < 16; f += 2 {
		address := gbppu.tilePattern + (f + (tile * 16))
		byte1 := gbmmu.fetchByte(address)
		byte2 := gbmmu.fetchByte(address + 1)
		bin := fmt.Sprintf("%08b%08b", byte1, byte2)
		fmt.Printf("%04x: %02x %02x: %s", address, byte1, byte2, bin)
		for j := uint16(0); j < 8; j++ {
			bit_0, _ := strconv.Atoi(string(bin[j]))
			bit_1, _ := strconv.Atoi(string(bin[j+8]))
			if bit_0+bit_1 > 0 {
				//offset := (f/12 * 4) + j + (160 * row)
				gbscreen.Pix[row*160+j+(tile*8)] = color.RGBA{255, 255, 255, 1}
				//fmt.Printf("Setting pixel at %d\n", row*160+offset)
			}
			offset++
		}
		row++
		offset = tile
		//if (f > 0) && (f%16 == 0) {
		//	row++
		//	offset = 0
		//	fmt.Printf("\n")
		//}
		fmt.Printf("\n")
	}
}

func (gbppu *ppu) processTileMap() {
	//mapPtr := 0x9800
	//tileX := 0
	//tileY := 0
	//tile := gbmmu.memory[mapPtr]
	//logo, _ := hex.DecodeString(sLogo)

	gbppu.showTilePattern(gbscreen, 1)
	gbppu.showTilePattern(gbscreen, 2)
	gbppu.showTilePattern(gbscreen, 3)
	gbppu.showTilePattern(gbscreen, 4)
	gbppu.showTilePattern(gbscreen, 5)
	gbppu.showTilePattern(gbscreen, 6)
	gbppu.showTilePattern(gbscreen, 7)
	gbppu.showTilePattern(gbscreen, 8)
	gbppu.showTilePattern(gbscreen, 9)
	gbppu.showTilePattern(gbscreen, 10)
	gbppu.showTilePattern(gbscreen, 11)
	gbppu.showTilePattern(gbscreen, 12)
}

//broken implementation of hblank
func (gbppu *ppu) hblank() {
	gbmmu.storeByte(gbppu.LY, gbmmu.fetchByte(gbppu.LY)+1)

	//if LY exceeds max value, reset
	if gbmmu.memory[gbppu.LY] > 0x99 {
		gbmmu.storeByte(gbppu.LY, 0)
	}
}

func (gbppu *ppu) vblank(win *pixelgl.Window) {
	sprite := pixel.NewSprite(gbscreen, gbscreen.Bounds())
	sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
	win.Update()
}

package main

import (
	"fmt"
	"image/color"
	"strconv"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

// Non-GBC colours
// darkest green 11 - RGB: 15,56,15
// dark green 10 - RGB: 48,98,48
// light green 01 - RGB: 139,172,15
// lightest green 00 - 155,188,15

var gbscreen *pixel.PictureData
var sprite *pixel.Sprite

//const SCRWIDTH uint16 = 160
//const SCRHEIGHT uint16 = 144

const SCRWIDTH uint16 = 256
const SCRHEIGHT uint16 = 256

//var sLogo string = "f000f000fc00fc00fc00fc00f300f3003c003c003c003c003c003c003c003c00f000f000f000f00000000000f300f300000000000000000000000000cf00cf00000000000f000f003f003f000f000f000000000000000000c000c0000f000f00000000000000000000000000f000f000000000000000000000000000f300f300000000000000000000000000c000c000030003000300030003000300ff00ff00c000c000c000c000c000c000c300c300000000000000000000000000fc00fc00f300f300f000f000f000f000f000f0003c003c00fc00fc00fc00fc003c003c00f300f300f300f300f300f300f300f300f300f300c300c300c300c300c300c300cf00cf00cf00cf00cf00cf00cf00cf003c003c003f003f003c003c000f000f003c003c00fc00fc0000000000fc00fc00fc00fc00f000f000f000f000f000f000f300f300f300f300f300f300f000f000c300c300c300c300c300c300ff00ff00cf00cf00cf00cf00cf00cf00c300c3000f000f000f000f000f000f00fc00fc003c004200b900a500b900a50042003c"

type ppu struct {
	LCDC        uint16 //FF40
	STAT        uint16 //FF41
	SCY         uint16 //FF42
	SCX         uint16 //FF43
	LY          uint16 //FF44
	LYC         uint16 //FF45
	BGP         uint16 //FF47 non-CGB
	OBP0        uint16 //FF48 non-CGB
	OBP1        uint16 //FF49 non-CGB
	tilePattern uint16
	tileMap     uint16
}

func (gbppu *ppu) initialise() {
	gbppu.LCDC = 0xFF40
	gbppu.STAT = 0xFF41
	gbppu.SCY = 0xFF42
	gbppu.SCX = 0xFF43
	gbppu.LY = 0xFF44
	gbppu.LYC = 0xFF45
	gbppu.tilePattern = 0x8000
	gbppu.tileMap = 0x9900

	gbscreen = pixel.MakePictureData(pixel.R(0, 0, float64(SCRWIDTH), float64(SCRHEIGHT)))
}

func (gbppu *ppu) drawLine(gbscreen *pixel.PictureData, row uint16) {
	var last_pixel uint16 = uint16(len(gbscreen.Pix)-1) + 1

	for i := uint16(0); i < 32; i++ {
		var tilePos uint16 = row/8*32 + i
		h_tile_start := (tilePos % 32) * 8

		tile := uint16(gbmmu.fetchByte(gbppu.tileMap+tilePos)) * 16
		address := gbppu.tilePattern + (uint16(row%8)*2 + tile)
		byte1 := gbmmu.fetchByte(address)
		byte2 := gbmmu.fetchByte(address + 1)
		bin := fmt.Sprintf("%08b%08b", byte1, byte2)
		debugLog(fmt.Sprintf("%04x: %02x %02x: %s", address, byte1, byte2, bin))
		// calculate the number of pixels to subtract from the end of pixel array to get the current row start
		var row_start_disp uint16 = (row + 1) * SCRWIDTH
		pixel := uint16(last_pixel) - row_start_disp + h_tile_start

		//for all eight pixels of the tile row
		for j := uint16(0); j < 8; j++ {
			bit_0, _ := strconv.Atoi(string(bin[j]))
			bit_1, _ := strconv.Atoi(string(bin[j+8]))
			var colour color.RGBA
			//todo - introduce palette functionality
			switch bit_0*2 + bit_1 {
			case 0:
				colour = color.RGBA{155, 188, 15, 1}
			case 1:
				colour = color.RGBA{139, 172, 15, 0}
			case 2:
				colour = color.RGBA{48, 98, 48, 0}
			case 3:
				colour = color.RGBA{15, 56, 15, 0}
			default:
				colour = color.RGBA{255, 255, 255, 0}
			}

			gbscreen.Pix[pixel+j] = colour
			debugLog(fmt.Sprintf("Setting pixel at %d\n", pixel))
		}
		tile++
		debugLog("\n")
	}
	//}
}

func (gbppu *ppu) showTilePattern(gbscreen *pixel.PictureData, tile uint16, tilePos uint16) {
	var row uint16 = uint16(tilePos/32) * 8
	var last_pixel uint16 = uint16(len(gbscreen.Pix) - 1)

	//for all sixteen bytes of a tile (2 at a time)
	for f := uint16(0); f < 16; f += 2 {
		address := gbppu.tilePattern + (f + (tile * 16))
		byte1 := gbmmu.fetchByte(address)
		byte2 := gbmmu.fetchByte(address + 1)
		bin := fmt.Sprintf("%08b%08b", byte1, byte2)
		debugLog(fmt.Sprintf("%04x: %02x %02x: %s", address, byte1, byte2, bin))
		// calculate the number of pixels to subtract from the end of pixel array to get the current row start
		var row_start_disp uint16 = (row + 1) * SCRWIDTH

		//for all eight pixels of a tile row
		for j := uint16(0); j < 8; j++ {
			bit_0, _ := strconv.Atoi(string(bin[j]))
			bit_1, _ := strconv.Atoi(string(bin[j+8]))
			var colour color.RGBA
			//todo - introduce palette functionality
			switch bit_0*2 + bit_1 {
			case 0:
				colour = color.RGBA{155, 188, 15, 1}
			case 1:
				colour = color.RGBA{139, 172, 15, 0}
			case 2:
				colour = color.RGBA{48, 98, 48, 0}
			case 3:
				colour = color.RGBA{15, 56, 15, 0}
			default:
				colour = color.RGBA{255, 255, 255, 0}
			}

			h_tile_start := (tilePos % 32) * 8
			pixel := uint16(last_pixel) - row_start_disp + h_tile_start + j
			gbscreen.Pix[pixel] = colour
			debugLog(fmt.Sprintf("Setting pixel at %d\n", pixel))
		}
		row++
		debugLog("\n")
	}
}

func (gbppu *ppu) processTileMap() {
	//mapPtr := 0x9800
	//tileX := 0
	//tileY := 0
	//tile := gbmmu.memory[mapPtr]
	//logo, _ := hex.DecodeString(sLogo)

	// loop for 32x32 tiles
	//for f := uint16(0); f < 1024; f++ {
	//	gbppu.showTilePattern(gbscreen, uint16(gbmmu.fetchByte(gbppu.tileMap+f)), f)
	//}
	for f := uint16(0); f < 256; f++ {
		gbppu.drawLine(gbscreen, f)
	}
}

//broken implementation of hblank
func (gbppu *ppu) hblank(win *pixelgl.Window) {
	gbmmu.storeByte(gbppu.LY, gbmmu.fetchByte(gbppu.LY)+1)

	if gbmmu.fetchByte(gbppu.LY) == 144 {
		gbppu.vblank(win)
	}
}

//draw the screen and update the window
func (gbppu *ppu) vblank(win *pixelgl.Window) {
	win.Clear(colornames.Black)
	gbppu.processTileMap()
	sprite = pixel.NewSprite(gbscreen, gbscreen.Bounds())
	sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
	win.Update()

	// vblank operates from LY=144 to 153 and then resets
	if gbmmu.fetchByte(gbppu.LY) > 153 {
		gbmmu.storeByte(gbppu.LY, 0)
	}
}

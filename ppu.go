package main

import (
	"image/color"

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
var gbColours [4]color.RGBA

const SCRWIDTH uint16 = 160
const SCRHEIGHT uint16 = 144

//const SCRWIDTH uint16 = 256
//const SCRHEIGHT uint16 = 256

//var sLogo string = "f000f000fc00fc00fc00fc00f300f3003c003c003c003c003c003c003c003c00f000f000f000f00000000000f300f300000000000000000000000000cf00cf00000000000f000f003f003f000f000f000000000000000000c000c0000f000f00000000000000000000000000f000f000000000000000000000000000f300f300000000000000000000000000c000c000030003000300030003000300ff00ff00c000c000c000c000c000c000c300c300000000000000000000000000fc00fc00f300f300f000f000f000f000f000f0003c003c00fc00fc00fc00fc003c003c00f300f300f300f300f300f300f300f300f300f300c300c300c300c300c300c300cf00cf00cf00cf00cf00cf00cf00cf003c003c003f003f003c003c000f000f003c003c00fc00fc0000000000fc00fc00fc00fc00f000f000f000f000f000f000f300f300f300f300f300f300f000f000c300c300c300c300c300c300ff00ff00cf00cf00cf00cf00cf00cf00c300c3000f000f000f000f000f000f00fc00fc003c004200b900a500b900a50042003c"

//holds the ADDRESS of these registers, not the CONTENTS (which are in memory)
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
	gbppu.tileMap = 0x9800

	gbColours[0] = color.RGBA{155, 188, 15, 1}
	//gbColours[0] = color.RGBA{0, 0, 0, 0}
	gbColours[1] = color.RGBA{139, 172, 15, 0}
	gbColours[2] = color.RGBA{48, 98, 48, 0}
	gbColours[3] = color.RGBA{15, 56, 15, 0}

	gbscreen = pixel.MakePictureData(pixel.R(0, 0, float64(SCRWIDTH), float64(SCRHEIGHT)))
}

func (gbppu *ppu) drawLine(gbscreen *pixel.PictureData) {
	last_pixel := uint16(len(gbscreen.Pix)-1) + 1
	screenRow := uint16(gbmmu.fetchByte(gbppu.LY))
	// calculate the number of pixels to subtract from the end of pixel array to get the current row start
	row_start_disp := (screenRow + 1) * SCRWIDTH
	bgRow := uint16(gbmmu.fetchByte(gbppu.SCY)) + screenRow

	//write 20*8 = 160 pixels for each row
	tilePos := bgRow / 8 * 32
	h_tile_start := (tilePos % 32) * 8
	for i := uint16(0); i < 20; i++ {
		tile := uint16(gbmmu.fetchByte(gbppu.tileMap+tilePos)) * 16
		tileRowAddress := gbppu.tilePattern + (uint16(bgRow%8)*2 + tile)
		byte1 := gbmmu.fetchByte(tileRowAddress)
		byte2 := gbmmu.fetchByte(tileRowAddress + 1)
		//bin := fmt.Sprintf("%08b%08b", byte1, byte2)
		//debugLog(fmt.Sprintf("%04x: %02x %02x: %s", tileRowAddress, byte1, byte2, bin))

		pixelIndex := uint16(last_pixel) - row_start_disp + h_tile_start

		//for all eight pixels of the tile row
		for j := 0; j < 8; j++ {
			//newByte1 := bits.RotateLeft8(byte1, -j) &^ 0b11111110
			//newByte2 := bits.RotateLeft8(byte2, -j) &^ 0b11111110
			var newByte1, newByte2 byte
			if byte1|byte2 > 0 {
				newByte1 = byte1 >> j &^ 0b11111110
				newByte2 = byte2 >> j &^ 0b11111110
				gbscreen.Pix[pixelIndex+(7-uint16(j))] = gbColours[int(newByte1)*2+int(newByte2)]
			}

			//colour := byte1 >> (6-j) + byte2 >> (8-j)

			//bit_0, _ := strconv.Atoi(string(bin[j]))
			//bit_1, _ := strconv.Atoi(string(bin[j+8]))
			//todo - introduce palette functionality
			//index := last_pixel - uint16(gbscreen.Index(pixel.V(float64(row), float64(i*8+j))))
			//gbscreen.Pix[pixelIndex+uint16(j)] = gbColours[bit_0*2+bit_1]
			//gbscreen.Pix[pixelIndex+(7-uint16(j))] = gbColours[int(newByte1)*2+int(newByte2)]

			//debugLog(fmt.Sprintf("Setting pixel at %d / %d\n", index, pixelIndex))
		}
		tilePos++
		h_tile_start += 8
		//debugLog("\n")
	}
}

//broken implementation of hblank
func (gbppu *ppu) hblank(win *pixelgl.Window) {
	//If LCD and PPU is enabled
	if isBitSet(gbmmu.fetchByte(gbppu.LCDC), 7) {
		//hblank after 63 clocks, but completes after 114 so need to change!
		if tstates > 63 {

			gbmmu.storeByte(gbppu.LY, gbmmu.fetchByte(gbppu.LY)+1)
			//fmt.Printf("In hblank. LY is %d, SCY is %d. Tstates=%d\n", gbmmu.fetchByte(gbppu.LY), gbmmu.fetchByte(gbppu.SCY), tstates)
			if gbmmu.fetchByte(gbppu.LY) < 144 {
				gbppu.drawLine(gbscreen)
			}
			tstates = 0
			if gbmmu.fetchByte(gbppu.LY) == 144 {
				//fmt.Printf("Calling vblank. LY is %d, SCY is %d. Tstates=%d\n", gbmmu.fetchByte(gbppu.LY), gbmmu.fetchByte(gbppu.SCY), tstates)
				gbppu.vblank(win)
			}
		}
	}
}

//draw the screen and update the window
func (gbppu *ppu) vblank(win *pixelgl.Window) {
	debugLog("In vblank\n", DEBUG_INFO)
	//win.Clear(color.RGBA{155, 188, 15, 0})
	win.Clear(colornames.Black)
	sprite = pixel.NewSprite(gbscreen, gbscreen.Bounds())
	sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
	win.Update()

	// vblank operates from LY=144 to 153 and then resets
	//if gbmmu.fetchByte(gbppu.LY) > 153 {
	//	gbmmu.storeByte(gbppu.LY, 0)
	//}
}

func (gbppu *ppu) processTileMap(win *pixelgl.Window) {
	//If LCD and PPU is enabled
	if isBitSet(gbmmu.fetchByte(gbppu.LCDC), 7) {

		LY := gbmmu.fetchByte(gbppu.LY) + 1
		gbmmu.storeByte(gbppu.LY, LY)
		//fmt.Printf("In hblank. LY is %d, SCY is %d. Tstates=%d\n", gbmmu.fetchByte(gbppu.LY), gbmmu.fetchByte(gbppu.SCY), tstates)
		//if tstates > 63 {
		if gbmmu.fetchByte(gbppu.LY) == 144 {
			//fmt.Printf("vblank\n")

			// loop for 32x32 tiles
			scy := gbmmu.fetchByte(gbppu.SCY)
			rowOffset := scy / 8
			for f := uint16(0); f < 18; f++ {
				for j := uint16(0); j < 20; j++ {
					tileIndex := (32 * (f + uint16(rowOffset))) + j
					var row = int(f)*8 - int(scy)%8
					if row > 0 {
						gbppu.showTilePattern(gbscreen, uint16(gbmmu.fetchByte(gbppu.tileMap+tileIndex)), tileIndex, uint16(row-1))
					}
				}
			}

			//fmt.Printf("Calling vblank. LY is %d, SCY is %d. Tstates=%d\n", gbmmu.fetchByte(gbppu.LY), gbmmu.fetchByte(gbppu.SCY), tstates)
			gbppu.vblank(win)
		}
	}
}

func (gbppu *ppu) showTilePattern(gbscreen *pixel.PictureData, tile uint16, tilePos uint16, row uint16) {
	//var row uint16 = uint16(tilePos/32) * 8
	var h_tile_start uint16 = (tilePos % 32) * 8
	var last_pixel uint16 = uint16(len(gbscreen.Pix)-1) + 1

	//for all sixteen bytes of a tile (2 at a time)
	address := gbppu.tilePattern + (tile * 16) - 2
	for f := uint16(0); f < 16; f += 2 {
		address += 2
		byte1 := gbmmu.fetchByte(address)
		byte2 := gbmmu.fetchByte(address + 1)
		// calculate the number of pixels to subtract from the end of pixel array to get the current row start
		var row_start_disp uint16 = (row + 1) * SCRWIDTH
		pixelIndex := uint16(last_pixel) - row_start_disp + h_tile_start
		if pixelIndex > 23040 {
			pixelIndex += 0
		}

		//for all eight pixels of a tile row
		for j := uint16(0); j < 8; j++ {
			var newByte1, newByte2 byte
			if byte1|byte2 > 0 {
				newByte1 = byte1 >> j &^ 0b11111110
				newByte2 = byte2 >> j &^ 0b11111110
			}
			//todo - introduce palette functionality
			gbscreen.Pix[pixelIndex+(7-uint16(j))] = gbColours[int(newByte1)*2+int(newByte2)]
			//debugLog(fmt.Sprintf("Setting pixel at %d\n", pixel))
		}
		row++
		if row >= SCRHEIGHT {
			row = 0
		}
		//debugLog("\n")
	}
}

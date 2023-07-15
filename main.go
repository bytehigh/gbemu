package main

import (
	"encoding/hex"
	"fmt"
	_ "image/png"
	"log"
	"os"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

var DEBUG uint8 = DEBUG_NONE

const (
	DEBUG_NONE uint8 = iota
	DEBUG_PC
	DEBUG_PUSHPOP
	DEBUG_VAR
	DEBUG_JP
	DEBUG_INFO
)

// const boot_rom string = "31 FE FF AF 21 FF 9F 32 CB 7C 20 FB 21 26 FF 0E 11 3E 80 32 E2 0C 3E F3 E2 32 3E 77 77 3E FC E0 47 11 04 01 21 10 80 1A CD 95 00 CD 96 00 13 7B FE 34 20 F3 11 D8 00 06 08 1A 13 22 23 05 20 F9 3E 19 EA 10 99 21 2F 99 0E 0C 3D 28 08 32 0D 20 F9 2E 0F 18 F3 67 3E 64 57 E0 42 3E 91 E0 40 04 1E 02 0E 0C F0 44 FE 90 20 FA 0D 20 F7 1D 20 F2 0E 13 24 7C 1E 83 FE 62 28 06 1E C1 FE 64 20 06 7B E2 0C 3E 87 E2 F0 42 90 E0 42 15 20 D2 05 20 4F 16 20 18 CB 4F 06 04 C5 CB 11 17 C1 CB 11 17 05 20 F5 22 23 22 23 C9 CE ED 66 66 CC 0D 00 0B 03 73 00 83 00 0C 00 0D 00 08 11 1F 88 89 00 0E DC CC 6E E6 DD DD D9 99 BB BB 67 63 6E 0E EC CC DD DC 99 9F BB B9 33 3E 3C 42 B9 A5 B9 A5 42 3C 21 04 01 11 A8 00 1A 13 BE 20 FE 23 7D FE 34 20 F5 06 19 78 86 23 05 20 FB 86 20 FE 3E 01 E0 50"
const boot_rom string = "31FEFFAF21FF9F32CB7C20FB2126FF0E113E8032E20C3EF3E2323E77773EFCE0471104012110801ACD9500CD9600137BFE3420F311D80006081A1322230520F93E19EA1099212F990E0C3D2808320D20F92E0F18F3673E6457E0423E91E040041E020E0CF044FE9020FA0D20F71D20F20E13247C1E83FE6228061EC1FE6420067BE20C3E87E2F04290E0421520D205204F162018CB4F0604C5CB1117C1CB11170520F522232223C9CEED6666CC0D000B03730083000C000D0008111F8889000EDCCC6EE6DDDDD999BBBB67636E0EECCCDDDC999FBBB9333E3C42B9A5B9A5423C21040111A8001A13BE20FE237DFE3420F506197886230520FB8620FE3E01E050"

var tstates uint16

type cpu struct {
	a, b, c, d, e, h, l byte
	f                   Bits
	pc, sp              uint16
	opcodes             map[uint16]string
	cb_prefix           bool
}

type Bits uint8

// Bit	7	6	5	4	3	2	1	0
// Flag	S	Z	F5	H	F3	P/V	N	C
const (
	C Bits = 1 << iota
	N
	PV
	F3
	H
	F5
	Z
	S
)

func Set(b, flag Bits) Bits    { return b | flag }
func Clear(b, flag Bits) Bits  { return b &^ flag }
func Toggle(b, flag Bits) Bits { return b ^ flag }
func Has(b, flag Bits) bool    { return b&flag != 0 }

func debugLog(message string, messageType uint8) {
	if DEBUG == messageType {
		fmt.Print(message)
	}
}

// #region
func (gbcpu *cpu) initialise() {
	gbcpu.cb_prefix = false
	gbcpu.opcodes = map[uint16]string{
		// 0x00
		0x0000: "nop", 0x0001: "ld_bc_d16", 0x0002: "ld_bc_a", 0x0003: "inc_bc",
		0x0004: "inc_b", 0x0005: "dec_b", 0x0006: "ld_b_d8", 0x0007: "rlca",
		0x0008: "ld_a16_sp", 0x0009: "add_hl_bc", 0x000A: "ld_a_bc", 0x000B: "dec_bc",
		0x000C: "inc_c", 0x000D: "dec_c", 0x000E: "ld_c_d8", 0x000F: "rrca",
		// 0x01
		0x0010: "stop_0", 0x0011: "ld_de_d16", 0x0012: "ld_de_a", 0x0013: "inc_de",
		0x0014: "inc_d", 0x0015: "dec_d", 0x0016: "ld_d_d8", 0x0017: "rla",
		0x0018: "jr_r8", 0x0019: "add_hl_de", 0x001A: "ld_a_de",
		/*"dec_de": 0x1B, */
		0x001C: "inc_e", 0x001D: "dec_e", 0x001E: "ld_e_d8",
		/* "rra": 0x1F,
		 */
		// 0x20
		0x0020: "jr_nz_r8", 0x0021: "ld_hl_d16", 0x0022: "ld_hl_plus_a", 0x0023: "inc_hl",
		0x0024: "inc_h",
		/*"dec_h": 0x25, "ld_h_d8": 0x26, "daa": 0x27,*/
		0x0028: "jr_z_r8",
		/*"add_hl_hl": 0x29,
		 */
		0x002A: "ld_a_hl_plus", 0x002B: "dec_hl",
		/* "inc_l": 0x2C, "dec_l": 0x2D,
		 */
		0x002E: "ld_l_d8", 0x002F: "cpl",
		// 0x30
		0x0030: "jr_nc_r8", 0x0031: "ld_sp_d16", 0x0032: "ld_hl_minus_a",
		/*:inc_sp, :inc__hl, */
		0x0035: "dec__hl", 0x0036: "ld_hl_d8",
		/*:scf,
		  :jr_c_r8, :add_hl_sp, :ld_a_hl_minus, :dec_sp, :inc_a, */
		0x003D: "dec_a", 0x003E: "ld_a_d8",
		/* // 0x40
		   :ld_b_b, :ld_b_c, :ld_b_d, :ld_b_e, :ld_b_h, :ld_b_l, :ld_b_hl, */
		0x0047: "ld_b_a",
		/* :ld_c_b, :ld_c_c, :ld_c_d, :ld_c_e, :ld_c_h, :ld_c_l, :ld_c_hl, */
		0x004F: "ld_c_a",
		/*
		  	// 0x50
		      :ld_d_b, :ld_d_c, :ld_d_d, :ld_d_e, :ld_d_h, :ld_d_l,
		*/
		0x0056: "ld_d_hl", 0x0057: "ld_d_a",
		/*	  :ld_e_b, :ld_e_c, :ld_e_d, :ld_e_e, :ld_e_h, :ld_e_l, */
		0x005E: "ld_e_hl", 0x005F: "ld_e_a",
		// 0x60
		/*:ld_h_b, :ld_h_c, :ld_h_d, :ld_h_e, :ld_h_h, :ld_h_l, :ld_h_hl,
		 */
		0x0067: "ld_h_a",
		/*:ld_l_b, :ld_l_c, :ld_l_d, :ld_l_e, :ld_l_h, :ld_l_l, :ld_l_hl, :ld_l_a,
		  // 0x70
		  :ld_hl_b, :ld_hl_c, :ld_hl_d, :ld_hl_e, :ld_hl_h, :ld_hl_l, :halt,
		*/
		0x0077: "ld_hl_a", 0x0078: "ld_a_b", 0x0079: "ld_a_c",
		/*:ld_a_d,
		 */
		0x007B: "ld_a_e", 0x007C: "ld_a_h", 0x007D: "ld_a_l",
		/*
			0x007E: "ld_a_hl",
			 :ld_a_a,
			  // 0x80
			  :add_a_b, :add_a_c, :add_a_d, :add_a_e, :add_a_h, :add_a_l,
		*/
		0x0086: "add_a_hl", 0x0087: "add_a_a",
		/* :adc_a_b,
		:adc_a_c, :adc_a_d, :adc_a_e, :adc_a_h, :adc_a_l, :adc_a_hl, :adc_a_a,
		*/
		// 0x90
		0x0090: "sub_b",
		/*
				  :sub_c, :sub_d, :sub_e, :sub_h, :sub_l, :sub_hl, :sub_a, :sbc_a_b, :sbc_a_c, :sbc_a_d,
			      :sbc_a_e, :sbc_a_h, :sbc_a_l, :sbc_a_hl, :sbc_a_a,
		*/
		// 0xA0
		/* "and_b":  0xA0,	*/
		0x00A1: "and_c",
		/*	"and_d":  0xA2,	"and_e":  0xA3,
			"and_h":  0xA4,	"and_l":  0xA5,	"and_hl": 0xA6,
		*/
		0x00A7: "and_a",
		/*	"xor_b":  0xA8,	*/
		0x00A9: "xor_c",
		/* "xor_d":  0xAA,	"xor_e":  0xAB,
		"xor_h":  0xAC,	"xor_l":  0xAD,	"xor_hl": 0xAE,
		*/
		0x00AF: "xor_a",
		0x00B0: "or_b", 0x00B1: "or_c",
		/*
			:or_d, :or_e, :or_h, :or_l, :or_hl, :or_a, :cp_b, :cp_c, :cp_d, :cp_e, :cp_h, :cp_l,
		*/
		0x00BE: "cp_hl",
		/* :cp_a,
		   // 0xC0
		   :ret_nz,
		*/
		0x00C1: "pop_bc",
		/*
		   :jp_nz_a16,
		*/
		0x00C3: "jp_a16",
		/*:call_nz_a16,
		 */
		0x00C5: "push_bc",
		/*
			:add_a_d8, :rst_00h, :ret_z,
		*/
		0x00C9: "ret", 0x00CA: "jp_z_a16",
		/*:prefix_cb, :call_z_a16,
		 */
		0x00CD: "call_a16",
		/*
			:adc_a_d8, :rst_08h,
			   // 0xD0
			   :ret_nc,
		*/
		0x00D1: "pop_de",
		/*:jp_nc_a16, :xx, :call_nc_a16,
		 */
		0x00D5: "push_de",
		/*:sub_d8, :rst_10h, :ret_c, :reti, :jp_c_a16, :xx,
		  :call_c_a16, :xx, :sbc_a_d8, :rst_18h,
		*/
		// 0xE0
		0x00E0: "ldh_a8_a", 0x00E1: "pop_hl", 0x00E2: "ld_dc_a", /* :xx, :xx,*/
		0x00E5: "push_hl", 0x00E6: "and_d8",
		/*:rst_20h, :add_sp_r8,
		 */
		0x00E9: "jp_dhl", 0x00EA: "ld_a16_a",
		/*   :xx, :xx, :xx, :xor_d8,
		 */
		0x00EF: "rst_28h",
		// 0xF0
		0x00F0: "ldh_a_a8",
		/* :pop_af, :ld_a_dc,
		 */
		0x00F3: "di",
		/*:xx, */
		0x00F5: "push_af",
		/*:or_d8, :rst_30h, :ld_hl_sp_r8, :ld_sp_hl,*/
		0x00FA: "ld_a_a16", 0x00FB: "ei",
		/*:xx, :xx,*/
		0x00FE: "cp_d8",
		/*:rst_38h
		 */
		// 0xCB7C
		0xCB7C: "bit_7_h",
		/*
					   CB_OPCODE = [
			        # 0x00
			        :rlc_b, :rlc_c, :rlc_d, :rlc_e, :rlc_h, :rlc_l, :rlc_hl, :rlc_a, :rrc_b, :rrc_c, :rrc_d, :rrc_e, :rrc_h, :rrc_l, :rrc_hl, :rrc_a,
			        # 0x10
			        :rl_b, :rl_c, :rl_d, :rl_e, :rl_h, :rl_l, :rl_hl,
		*/
		0xCB11: "rl_c",
		/*			:rr_b, :rr_c, :rr_d, :rr_e, :rr_h, :rr_l, :rr_hl, :rr_a,
		# 0x20
		:sla_b, :sla_c, :sla_d, :sla_e, :sla_h, :sla_l, :sla_hl, :sla_a, :sra_b, :sra_c, :sra_d, :sra_e, :sra_h, :sra_l, :sra_hl, :sra_a,
		# 0x30
		:swap_b, :swap_c, :swap_d, :swap_e, :swap_h, :swap_l, :swap_hl,*/
		0xCB37: "swap_a",
		/*:srl_b, :srl_c, :srl_d, :srl_e, :srl_h, :srl_l, :srl_hl, :srl_a,
		# 0x40
		:bit_0_b, :bit_0_c, :bit_0_d, :bit_0_e, :bit_0_h, :bit_0_l, :bit_0_hl, :bit_0_a, :bit_1_b, :bit_1_c, :bit_1_d, :bit_1_e, :bit_1_h,
		:bit_1_l, :bit_1_hl, :bit_1_a,
		# 0x50
		:bit_2_b, :bit_2_c, :bit_2_d, :bit_2_e, :bit_2_h, :bit_2_l, :bit_2_hl, :bit_2_a, :bit_3_b, :bit_3_c, :bit_3_d, :bit_3_e, :bit_3_h, :bit_3_l,
		:bit_3_hl, :bit_3_a,
		# 0x60
		:bit_4_b, :bit_4_c, :bit_4_d, :bit_4_e, :bit_4_h, :bit_4_l, :bit_4_hl, :bit_4_a, :bit_5_b, :bit_5_c, :bit_5_d, :bit_5_e, :bit_5_h, :bit_5_l,
		:bit_5_hl, :bit_5_a,
		# 0x70
		:bit_6_b, :bit_6_c, :bit_6_d, :bit_6_e, :bit_6_h, :bit_6_l, :bit_6_hl, :bit_6_a, :bit_7_b, :bit_7_c, :bit_7_d, :bit_7_e, :bit_7_h, :bit_7_l,
		:bit_7_hl, :bit_7_a,
		# 0x80
		:res_0_b, :res_0_c, :res_0_d, :res_0_e, :res_0_h, :res_0_l, :res_0_hl,
		*/
		0xCB87: "res_0_a",
		/*:res_1_b, :res_1_c, :res_1_d, :res_1_e, :res_1_h, :res_1_l,
		:res_1_hl, :res_1_a,
		# 0x90
		:res_2_b, :res_2_c, :res_2_d, :res_2_e, :res_2_h, :res_2_l, :res_2_hl, :res_2_a, :res_3_b, :res_3_c, :res_3_d, :res_3_e, :res_3_h, :res_3_l,
		:res_3_hl, :res_3_a,
		# 0xA0
		:res_4_b, :res_4_c, :res_4_d, :res_4_e, :res_4_h, :res_4_l, :res_4_hl, :res_4_a, :res_5_b, :res_5_c, :res_5_d, :res_5_e, :res_5_h, :res_5_l,
		:res_5_hl, :res_5_a,
		# 0xB0
		:res_6_b, :res_6_c, :res_6_d, :res_6_e, :res_6_h, :res_6_l, :res_6_hl, :res_6_a, :res_7_b, :res_7_c, :res_7_d, :res_7_e, :res_7_h, :res_7_l,
		:res_7_hl, :res_7_a,
		# 0xC0
		:set_0_b, :set_0_c, :set_0_d, :set_0_e, :set_0_h, :set_0_l, :set_0_hl, :set_0_a, :set_1_b, :set_1_c, :set_1_d, :set_1_e, :set_1_h, :set_1_l,
		:set_1_hl, :set_1_a,
		# 0xD0
		:set_2_b, :set_2_c, :set_2_d, :set_2_e, :set_2_h, :set_2_l, :set_2_hl, :set_2_a, :set_3_b, :set_3_c, :set_3_d, :set_3_e, :set_3_h, :set_3_l,
		:set_3_hl, :set_3_a,
		# 0xE0
		:set_4_b, :set_4_c, :set_4_d, :set_4_e, :set_4_h, :set_4_l, :set_4_hl, :set_4_a, :set_5_b, :set_5_c, :set_5_d, :set_5_e, :set_5_h, :set_5_l,
		:set_5_hl, :set_5_a,
		# 0xF0
		:set_6_b, :set_6_c, :set_6_d, :set_6_e, :set_6_h, :set_6_l, :set_6_hl, :set_6_a, :set_7_b, :set_7_c, :set_7_d, :set_7_e, :set_7_h, :set_7_l,
		:set_7_hl, :set_7_a
		*/
	}
}

//#endregion

func makeWord(msb, lsb byte) uint16 {
	return 256*uint16(msb) + uint16(lsb)
}

func getlsb(word uint16) byte {
	return byte(word)
}

func getmsb(word uint16) byte {
	return byte(word >> 8)
}

func isBitSet(value byte, bit int) bool {
	switch bit {
	case 0:
	case 1:
		return (value & 0b00000001) > 0
	case 7:
		return (value & 0b10000000) > 0
	default:
	}
	return false
}

// fetch next instruction at the program counter (PC)
func (gbcpu *cpu) fetch() byte {
	var opcode byte = gbmmu.fetchByte(gbcpu.pc)
	gbcpu.pc++
	tstates += 4

	return opcode
}

// execute a clock cycle
func (gbcpu *cpu) tick(gbmmu mmu, gbppu ppu) {
	//get the opcode at the current program counter (PC)
	//var opcode byte = gbmmu.memory[gbcpu.pc]
	//var asm string
	var opcode = gbcpu.fetch()
	if opcode == 0xCB {
		gbcpu.cb_prefix = true
		opcode = gbcpu.fetch()
		//asm = gbcpu.opcodes[uint16(0xCB)<<8+uint16(opcode)]
	} //else {
	asm := gbcpu.opcodes[uint16(opcode)]
	//}
	debugLog(fmt.Sprintf("PC: %04x Opcode is %02x %s\n", gbcpu.pc-1, opcode, asm), DEBUG_PC)
	//if isBitSet(gbmmu.fetchByte(gbppu.LCDC), 7) {
	//	fmt.Printf("PC: %04x Opcode is %02x %s\n", gbcpu.pc-1, opcode, asm)
	//}
	//perform relevant operation based on the current opcode
	//todo create 16 bit word with opcode and have one switch statement?
	if !gbcpu.cb_prefix {
		switch opcode {
		case 0x00:
			gbcpu.nop()
		case 0x01:
			gbcpu.ld_bc_d16()
		case 0x02:
			gbcpu.ld_bc_a()
		case 0x03:
			gbcpu.inc_bc()
		case 0x04:
			gbcpu.inc_b()
		case 0x05:
			gbcpu.dec_b()
		case 0x06:
			gbcpu.ld_b_d8()
		case 0x0B:
			gbcpu.dec_bc()
		case 0x0C:
			gbcpu.inc_c()
		case 0x0D:
			gbcpu.dec_c()
		case 0x0E:
			gbcpu.ld_c_d8()
		case 0x11:
			gbcpu.ld_de_d16()
		case 0x12:
			gbcpu.ld_de_a()
		case 0x13:
			gbcpu.inc_de()
		case 0x15:
			gbcpu.dec_d()
		case 0x16:
			gbcpu.ld_d_d8()
		case 0x17:
			gbcpu.rla()
		case 0x18:
			gbcpu.jr_r8()
		case 0x19:
			gbcpu.add_hl_de()
		case 0x1A:
			gbcpu.ld_a_de()
		case 0x1C:
			gbcpu.inc_e()
		case 0x1D:
			gbcpu.dec_e()
		case 0x1E:
			gbcpu.ld_e_d8()
		case 0x20:
			gbcpu.jr_nz_r8()
		case 0x21:
			gbcpu.ld_hl_d16()
		case 0x22:
			gbcpu.ld_hl_plus_a()
		case 0x23:
			gbcpu.inc_hl()
		case 0x24:
			gbcpu.inc_h()
		case 0x28:
			gbcpu.jr_z_r8()
		case 0x2A:
			gbcpu.ld_a_hl_plus()
		case 0x2B:
			gbcpu.dec_hl()
		case 0x2E:
			gbcpu.ld_l_d8()
		case 0x2F:
			gbcpu.cpl()
		case 0x31:
			gbcpu.ld_sp_d16()
		case 0x32:
			gbcpu.ld_hl_minus_a()
		case 0x35:
			gbcpu.dec__hl()
		case 0x36:
			gbcpu.ld_hl_d8()
		case 0x3D:
			gbcpu.dec_a()
		case 0x3E:
			gbcpu.ld_a_d8()
		case 0x47:
			gbcpu.ld_b_a()
		case 0x4F:
			gbcpu.ld_c_a()
		case 0x56:
			gbcpu.ld_d_hl()
		case 0x57:
			gbcpu.ld_d_a()
		case 0x5E:
			gbcpu.ld_e_hl()
		case 0x5F:
			gbcpu.ld_e_a()
		case 0x67:
			gbcpu.ld_h_a()
		case 0x77:
			gbcpu.ld_hl_a()
		case 0x78:
			gbcpu.ld_a_b()
		case 0x79:
			gbcpu.ld_a_c()
		case 0x7B:
			gbcpu.ld_a_e()
		case 0x7C:
			gbcpu.ld_a_h()
		case 0x7D:
			gbcpu.ld_a_l()
		case 0x86:
			gbcpu.add_a_hl()
		case 0x87:
			gbcpu.add_a_a()
		case 0x90:
			gbcpu.sub_b()
		case 0xA1:
			gbcpu.and_c()
		case 0xA7:
			gbcpu.and_a()
		case 0xA9:
			gbcpu.xor_c()
		case 0xAF:
			gbcpu.xor_a()
		case 0xB0:
			gbcpu.or_b()
		case 0xB1:
			gbcpu.or_c()
		case 0xBE:
			gbcpu.cp_hl()
		case 0xC1:
			gbcpu.pop_bc()
		case 0xC3:
			gbcpu.jp_a16()
		case 0xC5:
			gbcpu.push_bc()
		case 0xC9:
			gbcpu.ret()
		case 0xCA:
			gbcpu.jp_z_a16()
		case 0xCD:
			gbcpu.call_a16()
		case 0xD1:
			gbcpu.pop_de()
		case 0xD5:
			gbcpu.push_de()
		case 0xE0:
			gbcpu.ldh_a8_a()
		case 0xE1:
			gbcpu.pop_hl()
		case 0xE2:
			gbcpu.ld_dc_a()
		case 0xE5:
			gbcpu.push_hl()
		case 0xE6:
			gbcpu.and_d8()
		case 0xE9:
			gbcpu.jp_dhl()
		case 0xEA:
			gbcpu.ld_a16_a()
		case 0xEF:
			gbcpu.rst_28h()
		case 0xF0:
			gbcpu.ldh_a_a8()
		case 0xF3:
			gbcpu.di()
		case 0xF5:
			gbcpu.push_af()
		case 0xFA:
			gbcpu.ld_a_a16()
		case 0xFB:
			gbcpu.ei()
		case 0xFE:
			gbcpu.cp_d8()
		default:
			fmt.Printf("Opcode %02x at PC=%04x not implemented. Exiting\n", opcode, gbcpu.pc)
			os.Exit(1)
		}
	} else {
		// CB prefix instructions
		switch opcode {
		case 0x11:
			gbcpu.rl_c()
		case 0x37:
			gbcpu.swap_a()
		case 0x7C:
			gbcpu.bit_7_h()
		case 0x87:
			gbcpu.res_0_a()
		default:
			fmt.Printf("Extended CB opcode %02x at PC=%04x not implemented. Exiting\n", opcode, gbcpu.pc)
			os.Exit(1)
		}
		gbcpu.cb_prefix = false
	}
}

// 0x0000
func (gbcpu *cpu) nop() {
	//do nothing
}

// 0x0001
func (gbcpu *cpu) ld_bc_d16() {
	//LSB first
	gbcpu.c = gbcpu.fetch()
	gbcpu.b = gbcpu.fetch()
}

// 0x0002
func (gbcpu *cpu) ld_bc_a() {
	var bc = makeWord(gbcpu.b, gbcpu.c)
	gbcpu.a = gbmmu.fetchByte(bc)
}

// 0x0003
func (gbcpu *cpu) inc_bc() {
	var bc uint16 = 256*uint16(gbcpu.b) + uint16(gbcpu.c)
	bc++
	gbcpu.b = uint8(bc >> 8)
	gbcpu.c = uint8(bc & 0xFF)
	debugLog(fmt.Sprintf("b is %b, c is %b", gbcpu.b, gbcpu.c), DEBUG_VAR)
}

// 0x0004
func (gbcpu *cpu) inc_b() {
	gbcpu.b++
}

// 0x0005
func (gbcpu *cpu) dec_b() {
	//set Z flag as appropriate
	if gbcpu.b == 0x01 {
		gbcpu.f = Set(gbcpu.f, Z)
	} else {
		gbcpu.f = Clear(gbcpu.f, Z)
	}
	//todo - implement other flags
	gbcpu.b--
	debugLog(fmt.Sprintf("b is %02x\n", gbcpu.b), DEBUG_VAR)
}

// 0x0006
func (gbcpu *cpu) ld_b_d8() {
	gbcpu.b = gbcpu.fetch()
}

// 0x000B
func (gbcpu *cpu) dec_bc() {
	var bc uint16 = 256*uint16(gbcpu.b) + uint16(gbcpu.c)
	bc++
	gbcpu.b = uint8(bc >> 8)
	gbcpu.c = uint8(bc & 0xFF)
	debugLog(fmt.Sprintf("bc is %02x%02x\n", gbcpu.b, gbcpu.c), DEBUG_VAR)
}

// 0x000C
func (gbcpu *cpu) inc_c() {
	gbcpu.c++
}

// 0x000D
func (gbcpu *cpu) dec_c() {
	//set Z flag as appropriate
	if gbcpu.c == 0x01 {
		gbcpu.f = Set(gbcpu.f, Z)
	} else {
		gbcpu.f = Clear(gbcpu.f, Z)
	}
	//todo - implement other flags
	gbcpu.c--
	debugLog(fmt.Sprintf("c is %02x\n", gbcpu.c), DEBUG_VAR)
}

// 0x000E
func (gbcpu *cpu) ld_c_d8() {
	gbcpu.c = gbcpu.fetch()
}

// 0x0011
func (gbcpu *cpu) ld_de_d16() {
	//LSB first
	gbcpu.e = gbcpu.fetch()
	gbcpu.d = gbcpu.fetch()
}

// 0x0012
func (gbcpu *cpu) ld_de_a() {
	var de = makeWord(gbcpu.d, gbcpu.e)

	gbmmu.storeByte(de, gbcpu.a)
}

// 0x0013
func (gbcpu *cpu) inc_de() {
	//var de uint16 = 256*uint16(gbcpu.d) + uint16(gbcpu.e)
	var de = makeWord(gbcpu.d, gbcpu.e)
	de++
	gbcpu.d = uint8(de >> 8)
	gbcpu.e = uint8(de & 0xFF)
	debugLog(fmt.Sprintf("de is %02x%02x\n", gbcpu.d, gbcpu.e), DEBUG_VAR)
}

// 0x0015
func (gbcpu *cpu) dec_d() {
	//set Z flag as appropriate
	if gbcpu.d == 0x01 {
		gbcpu.f = Set(gbcpu.f, Z)
	} else {
		gbcpu.f = Clear(gbcpu.f, Z)
	}
	//todo - implement other flags
	gbcpu.d--
	debugLog(fmt.Sprintf("d is %02x\n", gbcpu.d), DEBUG_VAR)
}

// 0x0016
func (gbcpu *cpu) ld_d_d8() {
	gbcpu.d = gbcpu.fetch()
}

// 0x0017
func (gbcpu *cpu) rla() {
	//perform an RL A but faster with S, Z and P/V flags preserved
	//capture status of C flag
	carry := Has(gbcpu.f, C)

	//reset H
	gbcpu.f = Clear(gbcpu.f, H)
	//reset N
	gbcpu.f = Clear(gbcpu.f, N)
	//set C according to bit 7 of register A before the shift
	if gbcpu.a&0x80 == 0x80 {
		gbcpu.f = Set(gbcpu.f, C)
	} else {
		gbcpu.f = Clear(gbcpu.f, C)
	}

	gbcpu.a = gbcpu.a << 1
	//set bit 0 of A to carry flag
	if carry {
		gbcpu.a = gbcpu.a | 0x01
	}
}

// 0x0018
func (gbcpu *cpu) jr_r8() {
	//fetch relative offset for jump if required
	var rel_offset = gbcpu.fetch()

	if rel_offset > 127 {
		gbcpu.pc = gbcpu.pc - uint16((255 - rel_offset + 1))
	} else {
		gbcpu.pc = gbcpu.pc + uint16(rel_offset)
	}
}

func (gbcpu *cpu) add_hl_de() {
	var hl = makeWord(gbcpu.h, gbcpu.l)
	var de = makeWord(gbcpu.d, gbcpu.e)

	hl = hl + de
	gbcpu.h = uint8(hl >> 8)
	gbcpu.l = uint8(hl & 0xFF)
}

// 0x001A
func (gbcpu *cpu) ld_a_de() {
	var de = makeWord(gbcpu.d, gbcpu.e)
	gbcpu.a = gbmmu.fetchByte(de)
}

func (gbcpu *cpu) inc_e() {
	//set C flag as appropriate
	if gbcpu.e == 0xFF {
		gbcpu.f = Set(gbcpu.f, C)
	} else {
		gbcpu.f = Clear(gbcpu.f, C)
	}
	//todo - implement other flags
	gbcpu.e++
	debugLog(fmt.Sprintf("e is %02x\n", gbcpu.e), DEBUG_VAR)
}

// 0x001D
func (gbcpu *cpu) dec_e() {
	//set Z flag as appropriate
	if gbcpu.e == 0x01 {
		gbcpu.f = Set(gbcpu.f, Z)
	} else {
		gbcpu.f = Clear(gbcpu.f, Z)
	}
	//todo - implement other flags
	gbcpu.e--
	debugLog(fmt.Sprintf("e is %02x\n", gbcpu.e), DEBUG_VAR)
}

// 0x001E
func (gbcpu *cpu) ld_e_d8() {
	gbcpu.e = gbcpu.fetch()
}

// 0x0020
func (gbcpu *cpu) jr_nz_r8() {
	//fetch relative offset for jump if required
	var rel_offset = gbcpu.fetch()

	//check if Z flag not set
	if !Has(gbcpu.f, Z) {
		if rel_offset > 127 {
			debugLog(fmt.Sprintf("JR NZ to %04x\n", gbcpu.pc-uint16((255-rel_offset+1))), DEBUG_JP)
			gbcpu.pc = gbcpu.pc - uint16((255 - rel_offset + 1))
		} else {
			gbcpu.pc = gbcpu.pc + uint16(rel_offset)
		}
	}
}

// 0x0021
func (gbcpu *cpu) ld_hl_d16() {
	//LSB first
	gbcpu.l = gbcpu.fetch()
	gbcpu.h = gbcpu.fetch()
}

// 0x0022
func (gbcpu *cpu) ld_hl_plus_a() {
	var hl = makeWord(gbcpu.h, gbcpu.l)
	gbmmu.storeByte(hl, gbcpu.a)
	gbcpu.inc_hl()
}

// 0x0023
func (gbcpu *cpu) inc_hl() {
	var hl uint16 = 256*uint16(gbcpu.h) + uint16(gbcpu.l)
	hl++
	gbcpu.h = uint8(hl >> 8)
	gbcpu.l = uint8(hl & 0xFF)
	debugLog(fmt.Sprintf("HL is %02x%02x\n", gbcpu.h, gbcpu.l), DEBUG_VAR)
}

// 0x0024
func (gbcpu *cpu) inc_h() {
	gbcpu.h++
}

// 0x0028
func (gbcpu *cpu) jr_z_r8() {
	//fetch relative offset for jump if required
	var rel_offset = gbcpu.fetch()

	//check if Z flag set
	if Has(gbcpu.f, Z) {
		if rel_offset > 127 {
			gbcpu.pc = gbcpu.pc - uint16((255 - rel_offset + 1))
		} else {
			gbcpu.pc = gbcpu.pc + uint16(rel_offset)
		}
	}
}

// 0x002A
func (gbcpu *cpu) ld_a_hl_plus() {
	var hl uint16 = 256*uint16(gbcpu.h) + uint16(gbcpu.l)
	gbcpu.a = gbmmu.fetchByte(hl)
	gbcpu.inc_hl()
}

// 0x002B
func (gbcpu *cpu) dec_hl() {
	var hl uint16 = 256*uint16(gbcpu.h) + uint16(gbcpu.l)
	hl--
	gbcpu.h = uint8(hl >> 8)
	gbcpu.l = uint8(hl & 0xFF)
	debugLog(fmt.Sprintf("h is %02x, l is %02x\n", gbcpu.h, gbcpu.l), DEBUG_VAR)
}

// 0x002E
func (gbcpu *cpu) ld_l_d8() {
	gbcpu.l = gbcpu.fetch()
}

// 0x002F
func (gbcpu *cpu) cpl() {
	gbcpu.a = gbcpu.a ^ 0xFF
	//todo - set H and N flags
}

// 0x0031
func (gbcpu *cpu) ld_sp_d16() {
	var lsb = gbcpu.fetch()
	var msb = gbcpu.fetch()
	var d16 = makeWord(msb, lsb)

	gbcpu.sp = d16
}

// 0x0032
func (gbcpu *cpu) ld_hl_minus_a() {
	var hl = makeWord(gbcpu.h, gbcpu.l)
	gbmmu.storeByte(hl, gbcpu.a)
	gbcpu.dec_hl()
}

// 0x0035
func (gbcpu *cpu) dec__hl() {
	var hl uint16 = 256*uint16(gbcpu.h) + uint16(gbcpu.l)
	data := gbmmu.fetchByte(hl)
	data--
	gbmmu.storeByte(hl, data)
}

// 0x0036
func (gbcpu *cpu) ld_hl_d8() {
	var hl = makeWord(gbcpu.h, gbcpu.l)
	d8 := gbcpu.fetch()
	gbmmu.storeByte(hl, d8)
}

// 0x003D
func (gbcpu *cpu) dec_a() {
	//set Z flag as appropriate
	if gbcpu.a == 0x01 {
		gbcpu.f = Set(gbcpu.f, Z)
	} else {
		gbcpu.f = Clear(gbcpu.f, Z)
	}
	//todo - implement other flags
	gbcpu.a--
	debugLog(fmt.Sprintf("c is %02x\n", gbcpu.a), DEBUG_VAR)
}

// 0x003E
func (gbcpu *cpu) ld_a_d8() {
	data := gbcpu.fetch()
	gbcpu.a = data
}

// 0x003E
func (gbcpu *cpu) ld_c_a() {
	gbcpu.c = gbcpu.a
}

// 0x0047
func (gbcpu *cpu) ld_b_a() {
	gbcpu.b = gbcpu.a
}

func (gbcpu *cpu) ld_d_hl() {
	var hl = makeWord(gbcpu.h, gbcpu.l)

	gbcpu.d = gbmmu.fetchByte(hl)
}

// 0x0057
func (gbcpu *cpu) ld_d_a() {
	gbcpu.d = gbcpu.a
}

// 0x005F
func (gbcpu *cpu) ld_e_a() {
	gbcpu.e = gbcpu.a
}

// 0x005E
func (gbcpu *cpu) ld_e_hl() {
	var hl = makeWord(gbcpu.h, gbcpu.l)

	gbcpu.e = gbmmu.fetchByte(hl)
}

// 0x0067
func (gbcpu *cpu) ld_h_a() {
	gbcpu.h = gbcpu.a
}

// 0x0077
func (gbcpu *cpu) ld_hl_a() {
	var hl = makeWord(gbcpu.h, gbcpu.l)
	gbmmu.storeByte(hl, gbcpu.a)
}

// 0x0078
func (gbcpu *cpu) ld_a_b() {
	gbcpu.a = gbcpu.b
}

// 0x0079
func (gbcpu *cpu) ld_a_c() {
	gbcpu.a = gbcpu.c
}

// 0x007B
func (gbcpu *cpu) ld_a_e() {
	gbcpu.a = gbcpu.e
}

// 0x007C
func (gbcpu *cpu) ld_a_h() {
	gbcpu.a = gbcpu.h
}

// 0x007D
func (gbcpu *cpu) ld_a_l() {
	gbcpu.a = gbcpu.l
}

// 0x0086
func (gbcpu *cpu) add_a_hl() {
	var hl = makeWord(gbcpu.h, gbcpu.l)

	gbcpu.a = gbcpu.a + gbmmu.fetchByte(hl)
}

// 0x0087
func (gbcpu *cpu) add_a_a() {
	gbcpu.a = gbcpu.a + gbcpu.a
}

// 0x0090
func (gbcpu *cpu) sub_b() {
	gbcpu.a = gbcpu.a - gbcpu.b
}

// 0x00A1
func (gbcpu *cpu) and_c() {
	gbcpu.a = gbcpu.a & gbcpu.c
}

// 0x00A7
func (gbcpu *cpu) and_a() {
	gbcpu.a = gbcpu.a & gbcpu.a
	if gbcpu.a == 0 {
		gbcpu.f = Set(gbcpu.f, Z)
	} else {
		gbcpu.f = Clear(gbcpu.f, Z)
	}

}

// 0x00A9
func (gbcpu *cpu) xor_c() {
	gbcpu.a = gbcpu.a ^ gbcpu.c
}

// 0x00AF
func (gbcpu *cpu) xor_a() {
	gbcpu.a = 0 //replaces the "fast" way of setting A to zero
	//gbcpu.a = gbcpu.a ^ gbcpu.a
}

// 0x00B0
func (gbcpu *cpu) or_b() {
	gbcpu.a = gbcpu.a | gbcpu.b
}

// 0x00B1
func (gbcpu *cpu) or_c() {
	gbcpu.a = gbcpu.a | gbcpu.c
}

// 0x00BE
func (gbcpu *cpu) cp_hl() {
	var hl = makeWord(gbcpu.h, gbcpu.l)
	if gbcpu.a-gbmmu.fetchByte(hl) == 0 {
		gbcpu.f = Set(gbcpu.f, Z)
	} else {
		gbcpu.f = Clear(gbcpu.f, Z)
	}
	//todo - implement other flags
	debugLog(fmt.Sprintf("a is %02x, (hl) is %02x\n", gbcpu.a, gbmmu.fetchByte(hl)), DEBUG_VAR)
}

// 0x00C1
func (gbcpu *cpu) pop_bc() {
	gbcpu.sp--
	gbcpu.b = gbmmu.fetchByte(gbcpu.sp)
	gbcpu.sp--
	gbcpu.c = gbmmu.fetchByte(gbcpu.sp)
	debugLog(fmt.Sprintf("popped bc as %02x%02x\n", gbcpu.b, gbcpu.c), DEBUG_VAR)
}

// 0x00C3
func (gbcpu *cpu) jp_a16() {
	var lsb = gbcpu.fetch()
	var msb = gbcpu.fetch()
	var a16 = makeWord(msb, lsb)

	//jump to new location
	gbcpu.pc = a16
	debugLog(fmt.Sprintf("JP to PC: %04x\n", gbcpu.pc), DEBUG_JP)
}

// 0x00C5
func (gbcpu *cpu) push_bc() {
	debugLog(fmt.Sprintf("pushing bc as %02x%02x\n", gbcpu.b, gbcpu.c), DEBUG_PUSHPOP)
	gbmmu.storeByte(gbcpu.sp, gbcpu.c)
	gbcpu.sp++
	gbmmu.storeByte(gbcpu.sp, gbcpu.b)
	gbcpu.sp++
}

// 0x00C9
func (gbcpu *cpu) ret() {
	gbcpu.sp--
	msb := gbmmu.fetchByte(gbcpu.sp)
	gbcpu.sp--
	lsb := gbmmu.fetchByte(gbcpu.sp)
	gbcpu.pc = makeWord(msb, lsb)
	debugLog(fmt.Sprintf("Return popped to PC as %04x\n", gbcpu.pc), DEBUG_PUSHPOP)
}

// 0x00CA
func (gbcpu *cpu) jp_z_a16() {
	//fetch address for jump if required
	var address = makeWord(gbcpu.fetch(), gbcpu.fetch())

	gbcpu.pc = address
}

// 0x00CD
func (gbcpu *cpu) call_a16() {
	var lsb = gbcpu.fetch()
	var msb = gbcpu.fetch()
	var d16 = makeWord(msb, lsb)

	//push current PC onto stack
	debugLog(fmt.Sprintf("PC: %04x LSB %02x MSB %02x\n", gbcpu.pc, getlsb(gbcpu.pc), getmsb(gbcpu.pc)), DEBUG_VAR)
	gbmmu.storeByte(gbcpu.sp, getlsb(gbcpu.pc))
	gbcpu.sp++
	gbmmu.storeByte(gbcpu.sp, getmsb(gbcpu.pc))
	gbcpu.sp++

	//jump to new location
	gbcpu.pc = d16
	debugLog(fmt.Sprintf("Calling to PC: %04x\n", gbcpu.pc), DEBUG_JP)
}

// 0x00D1
func (gbcpu *cpu) pop_de() {
	gbcpu.sp--
	gbcpu.d = gbmmu.fetchByte(gbcpu.sp)
	gbcpu.sp--
	gbcpu.e = gbmmu.fetchByte(gbcpu.sp)
	debugLog(fmt.Sprintf("popped hl as %02x%02x\n", gbcpu.h, gbcpu.l), DEBUG_VAR)
}

// 0x00D5
func (gbcpu *cpu) push_de() {
	debugLog(fmt.Sprintf("pushing de as %02x%02x\n", gbcpu.d, gbcpu.e), DEBUG_VAR)
	gbmmu.storeByte(gbcpu.sp, gbcpu.e)
	gbcpu.sp++
	gbmmu.storeByte(gbcpu.sp, gbcpu.d)
	gbcpu.sp++
}

// 0x00E0
func (gbcpu *cpu) ldh_a8_a() {
	offset := gbcpu.fetch()

	gbmmu.storeByte(0xFF00+uint16(offset), gbcpu.a)
}

// 0x00E1
func (gbcpu *cpu) pop_hl() {
	gbcpu.sp--
	gbcpu.h = gbmmu.fetchByte(gbcpu.sp)
	gbcpu.sp--
	gbcpu.l = gbmmu.fetchByte(gbcpu.sp)
	debugLog(fmt.Sprintf("popped hl as %02x%02x\n", gbcpu.h, gbcpu.l), DEBUG_PUSHPOP)
}

// 0x00E2
func (gbcpu *cpu) ld_dc_a() {
	gbmmu.storeByte(0xFF00+uint16(gbcpu.c), gbcpu.a)
}

// 0x00E5
func (gbcpu *cpu) push_hl() {
	debugLog(fmt.Sprintf("pushing hl as %02x%02x\n", gbcpu.h, gbcpu.l), DEBUG_PUSHPOP)
	gbmmu.storeByte(gbcpu.sp, gbcpu.l)
	gbcpu.sp++
	gbmmu.storeByte(gbcpu.sp, gbcpu.h)
	gbcpu.sp++
}

// 0x00E6
func (gbcpu *cpu) and_d8() {
	d8 := gbcpu.fetch()

	gbcpu.a = gbcpu.a & d8
	if gbcpu.a == 0 {
		gbcpu.f = Set(gbcpu.f, Z)
	} else {
		gbcpu.f = Clear(gbcpu.f, Z)
	}

	//todo - C and N flags cleared, P/V is parity, rest are altered by definition.
}

// 0x00E9
func (gbcpu *cpu) jp_dhl() {
	var hl = makeWord(gbcpu.h, gbcpu.l)

	//jump to new location
	gbcpu.pc = hl
	debugLog(fmt.Sprintf("JP to PC: %04x\n", gbcpu.pc), DEBUG_JP)
	fmt.Printf("JP to PC: %04x\n", gbcpu.pc)
}

// 0x00EA
func (gbcpu *cpu) ld_a16_a() {
	var lsb = gbcpu.fetch()
	var msb = gbcpu.fetch()
	var a16 = makeWord(msb, lsb)

	debugLog(fmt.Sprintf("Loading A into (%04x)\n", a16), DEBUG_VAR)
	gbmmu.storeByte(a16, gbcpu.a)
}

// 0x00EF
func (gbcpu *cpu) rst_28h() {
	//push current PC onto stack
	debugLog(fmt.Sprintf("PC: %04x LSB %02x MSB %02x\n", gbcpu.pc, getlsb(gbcpu.pc), getmsb(gbcpu.pc)), DEBUG_VAR)
	gbmmu.storeByte(gbcpu.sp, getlsb(gbcpu.pc))
	gbcpu.sp++
	gbmmu.storeByte(gbcpu.sp, getmsb(gbcpu.pc))
	gbcpu.sp++

	//jump to new location
	gbcpu.pc = 0x28
	debugLog(fmt.Sprintf("Calling to PC: %04x\n", gbcpu.pc), DEBUG_JP)
}

// 0x00F0
func (gbcpu *cpu) ldh_a_a8() {
	offset := gbcpu.fetch()
	if offset == 0x44 {
		gbcpu.nop()
	}

	gbcpu.a = gbmmu.fetchByte(0xFF00 + uint16(offset))
}

// 0x00F3
func (gbcpu *cpu) di() {
	fmt.Println("DI - disable interrupts <todo>")
}

// 0x00F5
func (gbcpu *cpu) push_af() {
	debugLog(fmt.Sprintf("pushing af as %02x%02x\n", gbcpu.a, gbcpu.f), DEBUG_PUSHPOP)
	gbmmu.storeByte(gbcpu.sp, byte(gbcpu.f))
	gbcpu.sp++
	gbmmu.storeByte(gbcpu.sp, gbcpu.a)
	gbcpu.sp++
}

// 0x00FB
func (gbcpu *cpu) ei() {
	fmt.Println("EI - enable interrupts <todo>")
}

// 0x00FA
func (gbcpu *cpu) ld_a_a16() {
	var lsb = gbcpu.fetch()
	var msb = gbcpu.fetch()
	var a16 = makeWord(msb, lsb)
	var data = gbmmu.fetchByte(a16)

	gbcpu.a = data

	debugLog(fmt.Sprintf("Loading (%04x) into A\n", a16), DEBUG_VAR)
}

// 0x00FE
func (gbcpu *cpu) cp_d8() {

	operand := gbcpu.fetch()
	if gbcpu.a-operand == 0 {
		gbcpu.f = Set(gbcpu.f, Z)
	} else {
		gbcpu.f = Clear(gbcpu.f, Z)
	}
	//todo - implement other flags
	debugLog(fmt.Sprintf("a is %02x, operand is %02x\n", gbcpu.a, operand), DEBUG_VAR)
}

// 0xCB11
func (gbcpu *cpu) rl_c() {
	//capture status of C flag
	carry := Has(gbcpu.f, C)

	//set S if result is negative
	//todo
	//set Z if result is zero
	//todo
	//reset H
	gbcpu.f = Clear(gbcpu.f, H)
	//PV set if parity is even; otherwise, it is reset
	//todo
	//reset N
	gbcpu.f = Clear(gbcpu.f, N)
	//set C according to bit 7 of register A before the shift
	if gbcpu.c&0x80 == 0x80 {
		gbcpu.f = Set(gbcpu.f, C)
	} else {
		gbcpu.f = Clear(gbcpu.f, C)
	}

	gbcpu.c = gbcpu.c << 1
	//set bit 0 of A to carry flag
	if carry {
		gbcpu.c = gbcpu.c | 0x01
	}
}

// 0xCB37
func (gbcpu *cpu) swap_a() {
	copya := gbcpu.a
	gbcpu.a = gbcpu.a << 4
	copya = copya >> 4
	gbcpu.a = gbcpu.a | copya
}

// 0xCB7C
func (gbcpu *cpu) bit_7_h() {
	// check bit 7 of H register
	if gbcpu.h < 128 {
		// set Zero flag
		gbcpu.f = Set(gbcpu.f, Z)
	} else {
		gbcpu.f = Clear(gbcpu.f, Z)
	}
	// todo: leave C unchanged, N reset, H set, P/V undefined
}

// 0xCB87
func (gbcpu *cpu) res_0_a() {
	gbcpu.a = gbcpu.a & 0b11111110
}

func (gbcpu *cpu) status() {
	outlog := fmt.Sprintf("A:%02x ", gbcpu.a)
	outlog += fmt.Sprintf("F:%02x ", gbcpu.f)
	outlog += fmt.Sprintf("B:%02x ", gbcpu.b)
	outlog += fmt.Sprintf("C:%02x ", gbcpu.c)
	outlog += fmt.Sprintf("D:%02x ", gbcpu.d)
	outlog += fmt.Sprintf("E:%02x ", gbcpu.e)
	outlog += fmt.Sprintf("H:%02x ", gbcpu.h)
	outlog += fmt.Sprintf("L:%02x ", gbcpu.l)
	outlog += fmt.Sprintf("SP:%04x ", gbcpu.sp)
	outlog += fmt.Sprintf("PC:%04x ", gbcpu.pc)

	log.Print(outlog)
}

func run() {
	//gbmmu is global
	gbcpu := cpu{}
	gbppu := ppu{}
	gbrom := rom{}

	//initialise cpu, ppu, rom
	gbcpu.initialise()
	gbppu.initialise()
	gbrom.initialise()
	gbrom.load()

	//load boot.rom

	boot, err := hex.DecodeString(boot_rom)
	if err != nil {
		panic(err)
	}
	for i, op := range boot {
		gbmmu.storeByte(uint16(i), byte(op))
	}

	//load ROM into memory
	//todo

	//setup window (GB screen)
	cfg := pixelgl.WindowConfig{
		Title: "Pixel Rocks!",
		//Bounds: pixel.R(0, 0, 256, 256),
		Bounds: pixel.R(0, 0, float64(SCRWIDTH), float64(SCRHEIGHT)),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	//game loop
	//todo

	//execute clock cycle
	gbcpu.a = 0xFF
	for gbcpu.pc <= 65535 {
		gbcpu.tick(gbmmu, gbppu)
		gbcpu.status()

		//gbppu.hblank(win)
		//start := time.Now()
		if tstates >= 48 {
			gbppu.processTileMap(win)
			tstates = 0
		}

		//t := time.Now()
		//elapsed := t.Sub(start)
		//fmt.Printf("%s\n", elapsed)

		if win.Closed() {
			return
		}
	}
}

func main() {
	// log to custom file
	LOG_FILE := "./gbemu_log"
	// open log file
	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer logFile.Close()

	// Set log out put and enjoy :)
	log.SetOutput(logFile)
	log.SetFlags(0)

	pixelgl.Run(run)

	fmt.Printf("Program complete\n")

	//for f := uint16(0x8010); f < 0x9000; f++ {
	//	fmt.Printf("%02x", gbmmu.fetchByte(f))
	//}
}

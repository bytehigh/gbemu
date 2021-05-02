package main

import (
	"encoding/hex"
	"fmt"
	"os"
)

const mem_size = 65536

//const boot_rom string = "31 FE FF AF 21 FF 9F 32 CB 7C 20 FB 21 26 FF 0E 11 3E 80 32 E2 0C 3E F3 E2 32 3E 77 77 3E FC E0 47 11 04 01 21 10 80 1A CD 95 00 CD 96 00 13 7B FE 34 20 F3 11 D8 00 06 08 1A 13 22 23 05 20 F9 3E 19 EA 10 99 21 2F 99 0E 0C 3D 28 08 32 0D 20 F9 2E 0F 18 F3 67 3E 64 57 E0 42 3E 91 E0 40 04 1E 02 0E 0C F0 44 FE 90 20 FA 0D 20 F7 1D 20 F2 0E 13 24 7C 1E 83 FE 62 28 06 1E C1 FE 64 20 06 7B E2 0C 3E 87 E2 F0 42 90 E0 42 15 20 D2 05 20 4F 16 20 18 CB 4F 06 04 C5 CB 11 17 C1 CB 11 17 05 20 F5 22 23 22 23 C9 CE ED 66 66 CC 0D 00 0B 03 73 00 83 00 0C 00 0D 00 08 11 1F 88 89 00 0E DC CC 6E E6 DD DD D9 99 BB BB 67 63 6E 0E EC CC DD DC 99 9F BB B9 33 3E 3C 42 B9 A5 B9 A5 42 3C 21 04 01 11 A8 00 1A 13 BE 20 FE 23 7D FE 34 20 F5 06 19 78 86 23 05 20 FB 86 20 FE 3E 01 E0 50"
const boot_rom string = "31FEFFAF21FF9F32CB7C20FB2126FF0E113E8032E20C3EF3E2323E77773EFCE0471104012110801ACD9500CD9600137BFE3420F311D80006081A1322230520F93E19EA1099212F990E0C3D2808320D20F92E0F18F3673E6457E0423E91E040041E020E0CF044FE9020FA0D20F71D20F20E13247C1E83FE6228061EC1FE6420067BE20C3E87E2F04290E0421520D205204F162018CB4F0604C5CB1117C1CB11170520F522232223C9CEED6666CC0D000B03730083000C000D0008111F8889000EDCCC6EE6DDDDD999BBBB67636E0EECCCDDDC999FBBB9333E3C42B9A5B9A5423C21040111A8001A13BE20FE237DFE3420F506197886230520FB8620FE3E01E050"

var gbmmu mmu

type mmu struct {
	memory [mem_size]byte
}

type cpu struct {
	a, b, c, d, e, h, l byte
	f                   Bits
	pc, sp              uint16
	opcodes             map[uint16]string
	cb_prefix           bool
}

type Bits uint8

const (
	F0 Bits = 1 << iota
	F1
	F2
	F3
	F4
	F5
	F6
	F7
)

func Set(b, flag Bits) Bits    { return b | flag }
func Clear(b, flag Bits) Bits  { return b &^ flag }
func Toggle(b, flag Bits) Bits { return b ^ flag }
func Has(b, flag Bits) bool    { return b&flag != 0 }

//#region
func (gbcpu *cpu) initialise() {
	gbcpu.cb_prefix = false
	gbcpu.opcodes = map[uint16]string{
		// 0x00
		0x0000: "nop",
		/*
			"ld_bc_d16": 0x01, "ld_bc_a": 0x02, "inc_bc": 0x03,
			"inc_b": 0x04,
		*/
		0x0005: "dec_b", 0x0006: "ld_b_d8",
		/*
			"rlca": 0x07,
			"ld_a16_sp": 0x08, "add_hl_bc": 0x09, "ld_a_bc": 0x0A, "dec_bc": 0x0B,
		*/
		0x000C: "inc_c",
		/*
			"dec_c": 0x0D, */
		0x000E: "ld_c_d8",
		/*"rrca": 0x0F,
		// 0x01
		"stop_0": 0x10,*/
		0x0011: "ld_de_d16",
		/* "ld_de_a": 0x12, "inc_de": 0x13,
		"inc_d": 0x14, "dec_d": 0x15, "ld_d_d8": 0x16,
		*/
		0x0017: "rla",
		/*
			"jr_r8": 0x18, "add_hl_de": 0x19,
		*/
		0x001A: "ld_a_de",
		/*"dec_de": 0x1B,
		"inc_e": 0x1C, "dec_e": 0x1D, "ld_e_d8": 0x1E, "rra": 0x1F,
		*/
		// 0x20
		0x0020: "jr_nz_r8", 0x0021: "ld_hl_d16", 0x0022: "ld_hl_plus_a", 0x0023: "inc_hl",
		/*"inc_h": 0x24, "dec_h": 0x25, "ld_h_d8": 0x26, "daa": 0x27,
		"jr_z_r8": 0x28, "add_hl_hl": 0x29, "ld_a_hl_plus": 0x2A, "dec_hl": 0x2B,
		"inc_l": 0x2C, "dec_l": 0x2D, "ld_l_d8": 0x2E, "cpl": 0x2F,
		*/
		// 0x30
		0x0030: "jr_nc_r8", 0x0031: "ld_sp_d16", 0x0032: "ld_hl_minus_a",
		/*:inc_sp, :inc__hl, :dec__hl, :ld_hl_d8, :scf,
		  :jr_c_r8, :add_hl_sp, :ld_a_hl_minus, :dec_sp, :inc_a, :dec_a, */
		0x003E: "ld_a_d8",
		/*
		   // 0x40
		   :ld_b_b, :ld_b_c, :ld_b_d, :ld_b_e, :ld_b_h, :ld_b_l, :ld_b_hl, :ld_b_a, :ld_c_b,
		   :ld_c_c, :ld_c_d, :ld_c_e, :ld_c_h, :ld_c_l, :ld_c_hl,
		*/
		0x004F: "ld_c_a",
		/*
		  	// 0x50
		      :ld_d_b, :ld_d_c, :ld_d_d, :ld_d_e, :ld_d_h, :ld_d_l, :ld_d_hl, :ld_d_a, :ld_e_b,
		      :ld_e_c, :ld_e_d, :ld_e_e, :ld_e_h, :ld_e_l, :ld_e_hl, :ld_e_a,
		      // 0x60
		      :ld_h_b, :ld_h_c, :ld_h_d, :ld_h_e, :ld_h_h, :ld_h_l, :ld_h_hl, :ld_h_a, :ld_l_b,
		      :ld_l_c, :ld_l_d, :ld_l_e, :ld_l_h, :ld_l_l, :ld_l_hl, :ld_l_a,
		      // 0x70
		      :ld_hl_b, :ld_hl_c, :ld_hl_d, :ld_hl_e, :ld_hl_h, :ld_hl_l, :halt,
		*/
		0x0077: "ld_hl_a",
		/*
				:ld_a_b,
			      :ld_a_c, :ld_a_d, :ld_a_e, :ld_a_h, :ld_a_l, :ld_a_hl, :ld_a_a,
			      // 0x80
			      :add_a_b, :add_a_c, :add_a_d, :add_a_e, :add_a_h, :add_a_l, :add_a_hl, :add_a_a, :adc_a_b,
			      :adc_a_c, :adc_a_d, :adc_a_e, :adc_a_h, :adc_a_l, :adc_a_hl, :adc_a_a,
			  	  // 0x90
			      :sub_b, :sub_c, :sub_d, :sub_e, :sub_h, :sub_l, :sub_hl, :sub_a, :sbc_a_b, :sbc_a_c, :sbc_a_d,
			      :sbc_a_e, :sbc_a_h, :sbc_a_l, :sbc_a_hl, :sbc_a_a,
		*/
		// 0xA0
		/*
			"and_b":  0xA0,
			"and_c":  0xA1,
			"and_d":  0xA2,
			"and_e":  0xA3,
			"and_h":  0xA4,
			"and_l":  0xA5,
			"and_hl": 0xA6,
			"and_a":  0xA7,
			"xor_b":  0xA8,
			"xor_c":  0xA9,
			"xor_d":  0xAA,
			"xor_e":  0xAB,
			"xor_h":  0xAC,
			"xor_l":  0xAD,
			"xor_hl": 0xAE,
		*/
		0x00AF: "xor_a",
		/*
		   // 0xB0
		   :or_b, :or_c, :or_d, :or_e, :or_h, :or_l, :or_hl, :or_a, :cp_b, :cp_c, :cp_d, :cp_e, :cp_h, :cp_l,
		   :cp_hl, :cp_a,
		   // 0xC0
		   :ret_nz,
		*/
		0x00C1: "pop_bc",
		/*
		   :jp_nz_a16, :jp_a16, :call_nz_a16,
		*/
		0x00C5: "push_bc",
		/*
			:add_a_d8, :rst_00h, :ret_z,
		*/
		0x00C9: "ret",
		/*	   :jp_z_a16, :prefix_cb, :call_z_a16,
		 */
		0x00CD: "call_a16",
		/*
			:adc_a_d8, :rst_08h,
			   // 0xD0
			   :ret_nc, :pop_de, :jp_nc_a16, :xx, :call_nc_a16, :push_de, :sub_d8, :rst_10h, :ret_c, :reti, :jp_c_a16, :xx,
			   :call_c_a16, :xx, :sbc_a_d8, :rst_18h,
		*/
		// 0xE0
		0x00E0: "ldh_a8_a",
		/*
			:pop_hl,
		*/
		0x00E2: "ld_dc_a",
		/* :xx, :xx, :push_hl, :and_d8, :rst_20h, :add_sp_r8, :jp_dhl, :ld_a16_a,
		   :xx, :xx, :xx, :xor_d8, :rst_28h,
		   // 0xF0
		   :ldh_a_a8, :pop_af, :ld_a_dc, :di, :xx, :push_af, :or_d8, :rst_30h, :ld_hl_sp_r8, :ld_sp_hl, :ld_a_a16, :ei,
		   :xx, :xx, :cp_d8, :rst_38h
		*/
		// 0xCB7C
		0xCB7C: "bit_7_h",
		/*
					   CB_OPCODE = [
			        # 0x00
			        :rlc_b, :rlc_c, :rlc_d, :rlc_e, :rlc_h, :rlc_l, :rlc_hl, :rlc_a, :rrc_b, :rrc_c, :rrc_d, :rrc_e, :rrc_h, :rrc_l, :rrc_hl, :rrc_a,
			        # 0x10
			        :rl_b, :rl_c, :rl_d, :rl_e, :rl_h, :rl_l, :rl_hl, :rl_a, :rr_b, :rr_c, :rr_d, :rr_e, :rr_h, :rr_l, :rr_hl, :rr_a,
			        # 0x20
			        :sla_b, :sla_c, :sla_d, :sla_e, :sla_h, :sla_l, :sla_hl, :sla_a, :sra_b, :sra_c, :sra_d, :sra_e, :sra_h, :sra_l, :sra_hl, :sra_a,
			        # 0x30
			        :swap_b, :swap_c, :swap_d, :swap_e, :swap_h, :swap_l, :swap_hl, :swap_a, :srl_b, :srl_c, :srl_d, :srl_e, :srl_h, :srl_l, :srl_hl, :srl_a,
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
			        :res_0_b, :res_0_c, :res_0_d, :res_0_e, :res_0_h, :res_0_l, :res_0_hl, :res_0_a, :res_1_b, :res_1_c, :res_1_d, :res_1_e, :res_1_h, :res_1_l,
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

//fetch next instruction at the program counter (PC)
func (gbcpu *cpu) fetch() byte {
	var opcode byte = gbmmu.memory[gbcpu.pc]
	gbcpu.pc++

	return opcode
}

//execute a clock cycle
func (gbcpu *cpu) tick(gbmmu mmu) {
	//get the opcode at the current program counter (PC)
	//var opcode byte = gbmmu.memory[gbcpu.pc]
	var opcode = gbcpu.fetch()
	if opcode == 0xCB {
		gbcpu.cb_prefix = true
		opcode = gbcpu.fetch()
	}
	asm := gbcpu.opcodes[uint16(opcode)]
	fmt.Printf("PC: %04x Opcode is %02x %s\n", gbcpu.pc-1, opcode, asm)

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
			/*
				case 0x07:
				case 0x08:
				case 0x09:
				case 0x0A:
				case 0x0B:
			*/
		case 0x0C:
			gbcpu.inc_c()
			/*
				case 0x0D:
			*/
		case 0x0E:
			gbcpu.ld_c_d8()
			/*
				case 0x0F:
				case 0x10:
			*/
		case 0x11:
			gbcpu.ld_de_d16()
			/*
				case 0x12:
				case 0x13:
				case 0x14:
				case 0x15:
				case 0x16:
			*/
		case 0x17:
			gbcpu.rla()
			/*	case 0x18:
				case 0x19:
			*/
		case 0x1A:
			gbcpu.ld_a_de()
			/*	case 0x1B:
				case 0x1C:
				case 0x1D:
				case 0x1E:
				case 0x1F:
			*/
		case 0x20:
			gbcpu.jr_nz_r8()
		case 0x21:
			gbcpu.ld_hl_d16()
		case 0x22:
			gbcpu.ld_hl_plus_a()
		case 0x23:
			gbcpu.inc_hl()
			/*	case 0x24:
				case 0x25:
				case 0x26:
				case 0x27:
				case 0x28:
				case 0x29:
				case 0x2A:
				case 0x2B:
				case 0x2C:
				case 0x2D:
				case 0x2E:
				case 0x2F:
				case 0x30:
			*/
		case 0x31:
			gbcpu.ld_sp_d16()
		case 0x32:
			gbcpu.ld_hl_minus_a()
		case 0x3E:
			gbcpu.ld_a_d8()
		case 0x4F:
			gbcpu.ld_c_a()
		case 0x77:
			gbcpu.ld_hl_a()
		case 0xAF:
			gbcpu.xor_a()
		case 0xC1:
			gbcpu.pop_bc()
		case 0xC5:
			gbcpu.push_bc()
		case 0xC9:
			gbcpu.ret()
		case 0xCD:
			gbcpu.call_a16()
		case 0xE0:
			gbcpu.ldh_a8_a()
		case 0xE2:
			gbcpu.ld_dc_a()
		default:
			fmt.Printf("Opcode not implemented. Exiting\n")
			os.Exit(1)
		}
	} else {
		// CB prefix instructions
		switch opcode {
		case 0x7C:
			gbcpu.bit_7_h()
		default:
		}
		gbcpu.cb_prefix = false
	}
}

// 0x0000
func (gbcpu cpu) nop() {
	//do nothing
}

// 0x0001
func (gbcpu cpu) ld_bc_d16() {
	fmt.Printf("Instruction not yet implemented\n")
}

// 0x0002
func (gbcpu cpu) ld_bc_a() {
	var bc = makeWord(gbcpu.b, gbcpu.c)
	gbcpu.a = gbmmu.memory[bc]
}

// 0x0003
func (gbcpu *cpu) inc_bc() {
	var bc uint16 = 256*uint16(gbcpu.b) + uint16(gbcpu.c)
	bc++
	gbcpu.b = uint8(bc >> 8)
	gbcpu.c = uint8(bc & 0xFF)
	fmt.Printf("b is %b, c is %b", gbcpu.b, gbcpu.c)
}

// 0x0004
func (gbcpu *cpu) inc_b() {
	gbcpu.b++
}

// 0x0005
func (gbcpu *cpu) dec_b() {
	//set Z flag as appropriate
	if gbcpu.b == 0x01 {
		_ = Set(gbcpu.f, F6)
	}
	//todo - implement other flags
	gbcpu.b--
}

// 0x0006
func (gbcpu *cpu) ld_b_d8() {
	gbcpu.b = gbcpu.fetch()
}

// 0x000C
func (gbcpu *cpu) inc_c() {
	gbcpu.c++
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

// 0x0017
func (gbcpu *cpu) rla() {
	//capture status of C flag
	carry := Has(gbcpu.f, F0)

	//set S if result is negative
	//todo
	//set Z if result is zero
	//todo
	//reset H
	_ = Clear(gbcpu.f, F4)
	//PV set if parity is even; otherwise, it is reset
	//todo
	//reset N
	_ = Clear(gbcpu.f, F1)
	//set C according to bit 7 of register A before the shift
	if gbcpu.a&0x80 == 0x80 {
		_ = Set(gbcpu.f, F0)
	} else {
		_ = Clear(gbcpu.f, F0)
	}

	gbcpu.a = gbcpu.a << 1
	//set bit 0 of A to carry flag
	if carry {
		gbcpu.a = gbcpu.a | 0x01
	}
}

// 0x001A
func (gbcpu *cpu) ld_a_de() {
	var de = makeWord(gbcpu.d, gbcpu.e)
	gbcpu.a = gbmmu.memory[de]
}

// 0x0020
func (gbcpu *cpu) jr_nz_r8() {
	//fetch relative offset for jump if required
	var rel_offset = gbcpu.fetch()

	//check if Z flag not set
	if !Has(gbcpu.f, F6) {
		if rel_offset > 127 {
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
	gbmmu.memory[hl] = gbcpu.a
	gbcpu.inc_hl()
}

// 0x0023
func (gbcpu *cpu) inc_hl() {
	var hl uint16 = 256*uint16(gbcpu.h) + uint16(gbcpu.l)
	hl++
	gbcpu.h = uint8(hl >> 8)
	gbcpu.l = uint8(hl & 0xFF)
	fmt.Printf("HL is %02x%02x\n", gbcpu.h, gbcpu.l)
}

// 0x002B
func (gbcpu *cpu) dec_hl() {
	var hl uint16 = 256*uint16(gbcpu.h) + uint16(gbcpu.l)
	hl--
	gbcpu.h = uint8(hl >> 8)
	gbcpu.l = uint8(hl & 0xFF)
	fmt.Printf("h is %02x, l is %02x\n", gbcpu.h, gbcpu.l)
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
	gbmmu.memory[hl] = gbcpu.a
	gbcpu.dec_hl()
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

// 0x0077
func (gbcpu *cpu) ld_hl_a() {
	var hl = makeWord(gbcpu.h, gbcpu.l)
	gbmmu.memory[hl] = gbcpu.a
}

// 0x00AF
func (gbcpu *cpu) xor_a() {
	gbcpu.a = gbcpu.a ^ gbcpu.a
}

// 0x00C5
func (gbcpu *cpu) pop_bc() {
	gbcpu.sp--
	gbcpu.b = gbmmu.memory[gbcpu.sp]
	gbcpu.sp--
	gbcpu.c = gbmmu.memory[gbcpu.sp]
	fmt.Printf("popped bc as %02x%02x\n", gbcpu.b, gbcpu.c)
}

// 0x00C5
func (gbcpu *cpu) push_bc() {
	fmt.Printf("pushing bc as %02x%02x\n", gbcpu.b, gbcpu.c)
	gbmmu.memory[gbcpu.sp] = gbcpu.c
	gbcpu.sp++
	gbmmu.memory[gbcpu.sp] = gbcpu.b
	gbcpu.sp++
}

// 0x00C9
func (gbcpu *cpu) ret() {
	gbcpu.sp--
	msb := gbmmu.memory[gbcpu.sp]
	gbcpu.sp--
	lsb := gbmmu.memory[gbcpu.sp]
	gbcpu.pc = makeWord(msb, lsb)
	fmt.Printf("Return popped to PC as %04x\n", gbcpu.pc)
}

// 0x00CD
func (gbcpu *cpu) call_a16() {
	var lsb = gbcpu.fetch()
	var msb = gbcpu.fetch()
	var d16 = makeWord(msb, lsb)

	//push current PC onto stack
	fmt.Printf("PC: %04x LSB %02x MSB %02x\n", gbcpu.pc, getlsb(gbcpu.pc), getmsb(gbcpu.pc))
	gbmmu.memory[gbcpu.sp] = getlsb(gbcpu.pc)
	gbcpu.sp++
	gbmmu.memory[gbcpu.sp] = getmsb(gbcpu.pc)
	gbcpu.sp++

	//jump to new location
	gbcpu.pc = d16
	fmt.Printf("Calling to PC: %04x\n", gbcpu.pc)
}

// 0x00E0
func (gbcpu *cpu) ldh_a8_a() {
	offset := gbcpu.fetch()

	gbmmu.memory[0xFF00+uint16(offset)] = gbcpu.a
}

// 0x00E2
func (gbcpu *cpu) ld_dc_a() {
	gbmmu.memory[0xFF00+uint16(gbcpu.c)] = gbcpu.a
}

// 0xCB7C
func (gbcpu *cpu) bit_7_h() {
	// check bit 7 of H register
	if gbcpu.h < 128 {
		// set Zero flag
		gbcpu.f = Set(gbcpu.f, F6)
	}
	// todo: leave C unchanged, N reset, H set, P/V undefined
}

func main() {
	//gbmmu := mmu{}
	gbcpu := cpu{}

	//initialise cpu
	gbcpu.initialise()

	//load boot.rom

	boot, err := hex.DecodeString(boot_rom)
	if err != nil {
		panic(err)
	}
	for i, op := range boot {
		gbmmu.memory[i] = byte(op)
	}

	//load ROM into memory
	//todo

	//game loop
	//todo

	//execute clock cycle
	gbcpu.a = 0xFF
	for gbcpu.pc < 256 {
		gbcpu.tick(gbmmu)
	}
	fmt.Printf("Program complete\n")
}

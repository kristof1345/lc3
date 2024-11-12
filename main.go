package main

import (
	"fmt"
	"log"
	"os"
)

// consts
const MEMORY_MAX int = int(1 << 16)

const ( // registers
	R_R0 = iota
	R_R1
	R_R2
	R_R3
	R_R4
	R_R5
	R_R6
	R_R7
	R_PC // program counter
	R_COND
	R_COUNT // the count of registers
)

const (
	OP_BR   = iota // branch
	OP_ADD         // add
	OP_LD          // load
	OP_ST          // store
	OP_JSR         // jump register
	OP_AND         // bitwise and
	OP_LDR         // load register
	OP_STR         // store register
	OP_RTI         // unused
	OP_NOT         // bitwise not
	OP_LDI         // load indirect
	OP_STI         // store indirect
	OP_JMP         // jump
	OP_RES         // reserved(unused)
	OP_LEA         // load effective address
	OP_TRAP        // execute trap
)

const ( // conditional flags
	FL_POS = 1 << 0
	FL_ZRO = 1 << 1
	FL_NEG = 1 << 2
)

var (
	memory = make([]uint16, MEMORY_MAX) //a 65,536 sized empty array
	reg    = [R_COUNT]uint16{}
)

func updateFlags(r uint16) {
	if reg[r] == 0 {
		reg[R_COND] = FL_ZRO
	} else if reg[r]>>15 != 0 { // a '1' in the left-most bit indicates a negative. we get there by bitshiting with 15 becuaes it has 16 bits
		reg[R_COND] = FL_NEG
	} else {
		reg[R_COND] = FL_POS
	}
}

func signExtend(x uint16, bitCount int) uint16 {
	x = x & ((1 << bitCount) - 1)
	if (x>>(bitCount-1))&1 != 0 {
		x |= (0xFFFF) << bitCount
	}
	return x
}

func memRead(address uint16) uint16 {
	if int(address) < len(memory) {
		return memory[address]
	}
	return 0
}

func memWrite(address uint16, value uint16) {
	if address <= 65535 {
		memory[address] = value
	} else {
		log.Fatal("cannot write to memory")
	}
}

func readImage(path string) bool {
	return true
}

func main() {
	args := os.Args
	if len(args) < 2 {
		// show usage string
		fmt.Println("lc3 [image-file1] ...")
		os.Exit(2)
	}

	for i := 0; i < len(args); i++ {
		if !readImage(args[i]) {
			fmt.Printf("failed to load image: %s", args[i])
			os.Exit(1)
		}
	}

	reg[R_COND] = FL_ZRO

	// setting PC to default position
	PC_START := 0x3000
	reg[R_PC] = uint16(PC_START)

	running := true
	for running {
		// fetch
		instr := memRead(reg[R_PC])
		reg[R_PC]++
		op := instr >> 12

		switch op {
		case OP_ADD:
			r0 := (instr >> 9) & 0x7
			r1 := (instr >> 6) & 0x7
			immFlag := (instr >> 5) & 0x1

			if immFlag == 1 {
				imm5 := signExtend(instr&0x1F, 5)
				reg[r0] = reg[r1] + imm5
			} else {
				r2 := instr & 0x7
				reg[r0] = reg[r1] + reg[r2]
			}
			updateFlags(r0)
			break
		case OP_AND:
			r0 := (instr >> 9) & 0x7
			r1 := (instr >> 6) & 0x7
			immFlag := (instr >> 5) & 0x1

			if immFlag == 0 {
				r2 := instr & 0x7
				reg[r0] = reg[r1] & reg[r2]
			} else {
				imm5 := signExtend(instr&0x1F, 5)
				reg[r0] = reg[r1] & imm5
			}
			updateFlags(r0)
			break
		case OP_NOT:
			r0 := (instr >> 9) & 0x7
			r1 := (instr >> 6) & 0x7

			reg[r0] = ^reg[r1] // ^ is the nitwise XOR
			updateFlags(r0)
			break
		case OP_BR:
			pcOffset := signExtend(instr&0x1FF, 9)
			condFlag := (instr >> 9) & 0x7
			if condFlag&reg[R_COND] != 0 {
				reg[R_PC] += pcOffset
			}
			break
		case OP_JMP:
			r1 := (instr >> 6) & 0x7
			reg[R_PC] = reg[r1]
			break
		case OP_JSR:
			reg[R_R7] = reg[R_PC]
			flag := (instr >> 11) & 1
			if flag == 0 {
				r1 := (instr >> 6) & 0x7
				reg[R_PC] = reg[r1]
			} else {
				reg[R_PC] = reg[R_PC] + signExtend(instr&0x7FF, 11)
			}
			break
		case OP_LD:
			r0 := (instr >> 9) & 0x7
			pcOffset := signExtend(instr&0x1FF, 9)
			reg[r0] = memRead(reg[R_PC] + pcOffset)
			updateFlags(r0)
			break
		case OP_LDI:
			r0 := (instr >> 9) & 0x7
			pcOffset := signExtend(instr&0x1FF, 9)
			reg[r0] = memRead(memRead(reg[R_PC] + pcOffset))
			updateFlags(r0)
			break
		case OP_LDR:
			r0 := (instr >> 9) & 0x7
			offset := signExtend(instr&0x3F, 6)
			r1 := (instr >> 6) & 0x7
			reg[r0] = memRead(reg[r1] + offset)
			updateFlags(r0)
			break
		case OP_LEA:
			r0 := (instr >> 9) & 0x7
			pcOffset := signExtend(instr&0x1FF, 9)
			reg[r0] = reg[R_PC] + pcOffset
			updateFlags(r0)
			break
		case OP_ST:
			r0 := (instr >> 9) & 0x7
			pcOffset := signExtend(instr&0x1FF, 9)
			memWrite(reg[R_PC]+pcOffset, reg[r0])
			break
		case OP_STI:
			r0 := (instr >> 9) & 0x7
			pcOffset := signExtend(instr&0x1FF, 9)
			address := memRead(reg[R_PC] + pcOffset)
			memWrite(address, reg[r0])
			break
		case OP_STR:
			r0 := (instr >> 9) & 0x7
			r1 := (instr >> 6) & 0x7
			offset := signExtend(instr&0x3F, 6)
			memWrite(reg[r1]+offset, reg[r0])
			break
		case OP_TRAP:
			// trap
			break
		case OP_RES:
		case OP_RTI:
		default:
			panic("bad opcode")
		}

	}

	fmt.Sprintln("shutty down") // just to make gopls shut up
}

package main

import (
	"fmt"
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

func memRead(address uint16) uint16 {
	if int(address) < len(memory) {
		return memory[address]
	}
	return 0
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
			// add
			break
		case OP_AND:
			// and
			break
		case OP_NOT:
			// not
			break
		case OP_BR:
			// br
			break
		case OP_JMP:
			// jump
			break
		case OP_JSR:
			// jsr
			break
		case OP_LD:
			// ld
			break
		case OP_LDI:
			// ldi
			break
		case OP_LDR:
			// ldr
			break
		case OP_LEA:
			// lea
			break
		case OP_ST:
			// st
			break
		case OP_STI:
			// sti
			break
		case OP_STR:
			// str
			break
		case OP_TRAP:
			// trap
			break
		case OP_RES:
		case OP_RTI:
		default:
			// bad opcode
			break
		}

	}

	fmt.Sprintln("shutty down") // just to make gopls shut up
}

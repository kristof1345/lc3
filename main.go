package main

import "fmt"

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

func main() {
	memory := [MEMORY_MAX]uint16{} //a 65,536 sized empty array
	reg := [R_COUNT]uint16{}

	fmt.Sprintln(memory, reg) // just to make gopls shut up
}

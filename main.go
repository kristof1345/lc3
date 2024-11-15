package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
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

/* trap routines */
const (
	TRAP_GETC  = 0x20 /* get character from keyboard, not echoed onto the terminal */
	TRAP_OUT   = 0x21 /*output a chacarter*/
	TRAP_PUTS  = 0x22 /* output a word string */
	TRAP_IN    = 0x23 /* get a character from keyboard, echoed onto the terminal */
	TRAP_PUTSP = 0x24 /* output a byte string */
	TRAP_HALT  = 0x25 /* halt a program */
)

const ( // memory mapped registers - they allow the system to 'sleep' while waiting for user input from the keyboard
	MR_KBSR = 0xFE00 // 'event listener'
	MR_KBDR = 0xFE02 // data from keyboard
)

var (
	memory = make([]uint16, MEMORY_MAX) // a 65,536 sized empty array
	reg    = [R_COUNT]uint16{}
	mutex  sync.Mutex
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
	mutex.Lock()
	defer mutex.Unlock()

	if address == MR_KBSR {
		if peekChar() {
			reader := bufio.NewReader(os.Stdin)
			char, err := reader.ReadByte()
			if err != nil {
				panic("couldn't read from I/O")
			}

			memory[MR_KBSR] = (1 << 15)
			memory[MR_KBDR] = uint16(char)
		} else {
			memory[MR_KBSR] = 0
		}
	}

	if int(address) <= len(memory) {
		return memory[address]
	} else {
		log.Fatal("unhandled cpu memory read at")
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

func peekChar() bool {
	reader := bufio.NewReader(os.Stdin)
	_, err := reader.Peek(1)
	return err == nil
}

func readImage(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	stats, statsErr := file.Stat()
	if statsErr != nil {
		return false
	}

	// read origin
	var origin uint16

	headerBytes := make([]byte, 2) // 2 cuz one byte is 8 bits long - we need to read 16 bits
	_, err = file.Read(headerBytes)
	if err != nil {
		return false
	}

	headerBuffer := bytes.NewBuffer(headerBytes)
	// convert to big endian
	err = binary.Read(headerBuffer, binary.BigEndian, &origin)
	if err != nil {
		return false
	}
	log.Printf("Origin memory located: 0x%04X", origin)

	var size int64 = stats.Size()
	byteArr := make([]byte, size)

	log.Printf("Creating memory buffer: %d bytes", size)

	_, err = file.Read(byteArr)
	if err != nil {
		return false
	}

	buffer := bytes.NewBuffer(byteArr)

	// read into mem
	for i := origin; i < math.MaxUint16; i++ {
		var val uint16
		binary.Read(buffer, binary.BigEndian, &val)
		memory[i] = val
	}

	return true
}

// main function  
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
		case OP_BR:
			pcOffset := signExtend(instr&0x1FF, 9)
			condFlag := (instr >> 9) & 0x7
			if condFlag&reg[R_COND] != 0 {
				reg[R_PC] += pcOffset
			}
		case OP_JMP:
			r1 := (instr >> 6) & 0x7
			reg[R_PC] = reg[r1]
		case OP_JSR:
			reg[R_R7] = reg[R_PC]
			flag := (instr >> 11) & 1
			if flag == 0 {
				r1 := (instr >> 6) & 0x7
				reg[R_PC] = reg[r1]
			} else {
				reg[R_PC] = reg[R_PC] + signExtend(instr&0x7FF, 11)
			}
		case OP_LD:
			r0 := (instr >> 9) & 0x7
			pcOffset := signExtend(instr&0x1FF, 9)
			reg[r0] = memRead(reg[R_PC] + pcOffset)
			updateFlags(r0)
		case OP_LDI:
			r0 := (instr >> 9) & 0x7
			pcOffset := signExtend(instr&0x1FF, 9)
			reg[r0] = memRead(memRead(reg[R_PC] + pcOffset))
			updateFlags(r0)
		case OP_LDR:
			r0 := (instr >> 9) & 0x7
			offset := signExtend(instr&0x3F, 6)
			r1 := (instr >> 6) & 0x7
			reg[r0] = memRead(reg[r1] + offset)
			updateFlags(r0)
		case OP_LEA:
			r0 := (instr >> 9) & 0x7
			pcOffset := signExtend(instr&0x1FF, 9)
			reg[r0] = reg[R_PC] + pcOffset
			updateFlags(r0)
		case OP_ST:
			r0 := (instr >> 9) & 0x7
			pcOffset := signExtend(instr&0x1FF, 9)
			memWrite(reg[R_PC]+pcOffset, reg[r0])
		case OP_STI:
			r0 := (instr >> 9) & 0x7
			pcOffset := signExtend(instr&0x1FF, 9)
			address := memRead(reg[R_PC] + pcOffset)
			memWrite(address, reg[r0])
		case OP_STR:
			r0 := (instr >> 9) & 0x7
			r1 := (instr >> 6) & 0x7
			offset := signExtend(instr&0x3F, 6)
			memWrite(reg[r1]+offset, reg[r0])
		case OP_TRAP:
			reg[R_R7] = reg[R_PC]

			switch instr & 0xFF {
			case TRAP_GETC:
				reader := bufio.NewReader(os.Stdin)
				char, _, err := reader.ReadRune()
				if err != nil {
					panic("tried reading entered char, failed")
				}
				reg[R_R0] = uint16(char)
				updateFlags(R_R0)
			case TRAP_OUT:
				char := reg[R_R0]
				fmt.Printf("%c", rune(char))
			case TRAP_PUTS:
				address := reg[R_R0]
				var chr uint16
				var i uint16
				for ok := true; ok; ok = (chr != 0x0) {
					chr = memory[address+i] & 0xFFFF
					fmt.Printf("%c", rune(chr))
					i++
				}
			case TRAP_PUTSP:
				address := reg[R_R0]
				for i := uint16(0); ; i++ {
					chr := memory[address+i]
					if chr == 0 {
						break
					}

					char1 := chr & 0xFF
					fmt.Printf("%c", rune(char1))

					char2 := chr >> 8
					if char2 != 0 {
						fmt.Printf("%c", rune(char2))
					}
					i++
				}
			case TRAP_IN:
				fmt.Println("Enter character: ")
				reader := bufio.NewReader(os.Stdin)
				char, _, err := reader.ReadRune()
				if err != nil {
					panic("tried reading entered char, failed")
				}
				reg[R_R0] = uint16(char)
				updateFlags(R_R0)
				/* case TRAP_PUTSP: */
				// trap whatever
			case TRAP_HALT:
				fmt.Println("HALT")
				running = false
			}
		case OP_RES:
		case OP_RTI:
		default:
			panic("bad opcode")
		}

	}

	fmt.Sprintln("shutty down") // just to make gopls shut up
}

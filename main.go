// https://cedriczirtacic.github.io/
package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/arch/x86/x86asm"

	"./create_elf"
)

/*
#include <stdio.h>
#include <termios.h>

int get() {
	struct termios old, new;

	tcgetattr(0,&old);
	new=old;
	new.c_lflag &=~(ECHO | ICANON | IEXTEN);
	tcsetattr(0,TCSANOW,&new);
	char c = getchar();
	tcsetattr(0,TCSANOW,&old);
	return (int)c;
}
*/
import "C"

const (
	// architecture modes
	MOD_16 = (1 << 4)
	MOD_32 = (MOD_16 << 1)
	MOD_64 = (MOD_32 << 1)

	// flavors
	SYN_INTEL = (0 << 1)
	SYN_ATT   = (1 << 1)
	SYN_GO    = (2 << 1)

	// key codes
	KBD_TAB   = 9
	KBD_ENTER = 10
	KBD_LF    = 13

	PROMPT = "> "
)

// main options
var (
	mode        int   = MOD_64 /* default is 64-bit mode */
	syntax      uint8 = SYN_INTEL
	output      *os.File
	out         string
	json_output bool = false
	prettify    bool = false
	custom_elf  *os.File
	rgex        = regexp.MustCompile(`([^\s\t]+?)[\s\t]+(.+)*`)
)

// json output struct
type json_opcode struct {
	Len  int      `json:"length"`
	Inst string   `json:"instruction"`
	Args []string `json:"args"`
}

func output_file(f string) *os.File {
	fd, err := os.OpenFile(f, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return nil
	}

	return fd
}

func pretty_print(s string) {
	//colors
	var (
		Normal = 0
		Red    = 31 // operands
		Blue   = 34 // instruction
	)
	var escape = []byte{0x1b}

	matches := rgex.FindStringSubmatch(s)

	if len(matches) > 0 {
		fmt.Printf("%s[%dm%s ", escape, Blue, matches[1])
		if matches[2] != "" {
			fmt.Printf("%s[%dm%s", escape, Red, matches[2])
		}
		fmt.Printf("%s[%dm\n", escape, Normal)
	}
}

func PrintPrompt() {
	fmt.Printf("%c[2K\r%s", 27, PROMPT)
}

func main() {
	hist, err := NewHist()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up the history: %s\n", err)
	}
	defer hist.Save()
	hist.Populate()

	for {
		PrintPrompt()
		hist.ResetPos()

		var input []byte
		for {
			c := int(C.get())
			if c < 0x20 || c > 0x7e {
				if c == KBD_LF || c == KBD_ENTER {
					println()
					break
				}
				// del
				if c == 0x7f {
					if len(input) == 0 {
						continue
					}
					input = input[:len(input)-1]
					PrintPrompt()
					print(string(input))
				}
				// arrows
				if c == 0x1b {
					_ = int(C.get()) // 0x5b
					c2 := int(C.get())
					var cmd *string
					switch c2 {
					case 0x41:
						if len(hist.cmds) > 0 {
							cmd = hist.Up()
						} else {
							continue
						}
					case 0x42:
						if len(hist.cmds) > 0 {
							cmd = hist.Down()
						} else {
							continue
						}
					}

					PrintPrompt()
					print(*cmd)
					input = []byte(*cmd)
				}
			} else {
				fmt.Printf("%c", c)
				input = append(input, byte(c))
			}
		}

		input_codes := string(input)

		if input_codes == "" {
			continue
		} else if strings.Index(input_codes, "history") == 0 {
			if len(input_codes) == len("history") {
				hist.PrintHistory()
			} else if input_codes == "history clear" {
				hist.Clear()
				hist.Save()
			}
			continue
		} else if input_codes == "quit" || input_codes == "q" {
			return
		}
		hist.AppendCmd(input_codes)
		if len(input_codes) > 6 && input_codes[0:6] == "create" {
			create := strings.Split(input_codes, " ")
			custom_elf, _ = create_elf.Create(create[1])
			defer custom_elf.Close()

			continue
		} else if len(input_codes) > 3 && input_codes[0:3] == "set" {
			set := strings.Split(input_codes, "=")
			switch set[0] {
			case "set mode":
				if len(set) < 2 {
					fmt.Printf("Error: must specify a mode (eg. set mode=32)\n")
				} else {
					tmp_mode, _ := strconv.Atoi(set[1])
					if tmp_mode != MOD_16 && tmp_mode != MOD_32 && tmp_mode != MOD_64 {
						fmt.Printf("Error: invalid mode\n")
					} else {
						mode = tmp_mode
					}
				}
			case "set json":
				fmt.Printf("json = ")
				if !json_output {
					json_output = true
					fmt.Printf("true")
				} else {
					json_output = false
					fmt.Printf("false")
				}
				println()
			case "set flavor":
				if set[1] == "intel" {
					syntax = SYN_INTEL
				} else if set[1] == "att" {
					syntax = SYN_ATT
				} else if set[1] == "go" {
					syntax = SYN_GO
				} else {
					fmt.Fprintf(os.Stderr, "Error: unknown flavor.\n")
				}
				break
			case "set output":
				output = output_file(set[1])
				if output != nil {
					defer output.Close()
				}
			case "set colors":
				fmt.Printf("colors = ")
				if !prettify {
					prettify = true
					fmt.Printf("true")
				} else {
					prettify = false
					fmt.Printf("false")
				}
				println()
			default:
				fmt.Fprintf(os.Stderr, "Error: couldn't set an option.\n")
			}
			continue
		} else if len(input_codes)%2 != 0 {
			fmt.Fprintf(os.Stderr, "Error: unable to parse opcodes.\n")
			continue
		}

		opcodes := make([]byte, len(input_codes)/2)
		for j, i := 0, 0; i < len(input_codes); i = i + 2 {
			o, err := hex.DecodeString(input_codes[i : i+2])
			if err == nil {
				opcodes[j] = o[0]
				j++
			}
		}

		// if we are creating a custom ELF binary
		// then write all opcodes directly to the
		// specified file with create command
		if custom_elf != nil {
			err := create_elf.Write(custom_elf, opcodes)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: unable to write to custom ELF.\n")
			}
		}

		inst, _ := x86asm.Decode(opcodes, mode)

		// inst.Opcode > 0 to be a valid opcode
		if inst.Opcode == 0 {
			println("Wrong or invalid opcode")
			continue
		}

		// get parsed data
		switch syntax {
		case SYN_ATT:
			out = fmt.Sprintf("%s\n", x86asm.GNUSyntax(inst, 0, nil))
		case SYN_INTEL:
			out = fmt.Sprintf("%s\n", x86asm.IntelSyntax(inst, 0, nil))
		case SYN_GO:
			out = fmt.Sprintf("%s\n", x86asm.GoSyntax(inst, 0, nil))
		}

		if json_output {
			// JSON output
			op_data := &json_opcode{
				Len: inst.Len,
			}

			s := rgex.FindAllStringSubmatch(out, -1)
			op_data.Inst = s[0][1]

			if s[0][2] != "" {
				a := strings.Split(s[0][2], ",")
				for _, b := range a {
					if b[0:1] == " " {
						op_data.Args = append(op_data.Args, b[1:])
					} else {
						op_data.Args = append(op_data.Args, b)
					}
				}
			}
			json_data, _ := json.MarshalIndent(op_data, "", "    ")
			println(string(json_data))
		} else {
			// NORMAL output
			if output == nil {
				if !prettify {
					print(out)
				} else {
					pretty_print(out)
				}
			} else {
				output.WriteString(out)
			}
		}
	}

}

// https://cedriczirtacic.github.io/
package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"golang.org/x/arch/x86/x86asm"
	"os"
	"strings"
)

const (
	SYN_INTEL = (0 << 1)
	SYN_ATT   = (1 << 1)
	SYN_GO    = (2 << 1)
)

var (
    syntax uint8 = SYN_INTEL
    output *os.File
    out string
)

func output_file(f string) *os.File {
    fd, err := os.OpenFile(f, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0644)
    if err != nil {
        fmt.Fprintf(os.Stderr, "%s\n", err)
        return nil
    }

    return fd
}

func main() {
	input := bufio.NewScanner(os.Stdin)
	for {
		print("> ")
		input.Scan()
		input_codes := input.Text()
		if input_codes == "" {
			continue
		} else if input_codes == "quit" || input_codes == "q" {
			os.Exit(0)
		} else if len(input_codes) > 3 && input_codes[0:3] == "set" {
			set := strings.Split(input_codes, "=")
			switch set[0] {
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
		inst, _ := x86asm.Decode(opcodes, 64)
		switch syntax {
		case SYN_ATT:
			out = fmt.Sprintf("%s\n", x86asm.GNUSyntax(inst, 0, nil))
		case SYN_INTEL:
			out = fmt.Sprintf("%s\n", x86asm.IntelSyntax(inst, 0, nil))
		case SYN_GO:
			out = fmt.Sprintf("%s\n", x86asm.GoSyntax(inst, 0, nil))
		}
        if output == nil {
            print(out)
        } else {
            output.WriteString(out)
        }
	}

}
package create_elf

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"os"
	"unsafe"
)

const (
	VOffset uint64 = 0x400000
)

type custom_elf_headers struct {
	elf_header elf.Header64
	//section_header elf.Section64
	program_header elf.Prog64
}

func Create(path string) (*os.File, error) {
	fd, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return nil, err
	}

	var headers custom_elf_headers

	// setting default ELF64 header info
	headers.elf_header = elf.Header64{
		Ident:     [16]byte{0x7f, 'E', 'L', 'F', 02, 01, 01},
		Type:      uint16(elf.ET_EXEC),
		Machine:   uint16(elf.EM_X86_64),
		Version:   uint32(elf.EV_CURRENT),
		Flags:     0,
		Ehsize:    uint16(unsafe.Sizeof(headers.elf_header)),
		Phnum:     1, // one program header (LOAD)
		Shentsize: 0,
		Shnum:     0, // no section header
		Shstrndx:  0,
	}

	// setting default Prog64 headers
	headers.program_header = elf.Prog64{
		Type:  uint32(elf.PT_LOAD),
		Flags: uint32(elf.PF_X) | uint32(elf.PF_R),
		Off:   uint64(unsafe.Sizeof(headers.elf_header)),
		Vaddr: uint64(headers.elf_header.Ehsize) + VOffset,
		Paddr: uint64(headers.elf_header.Ehsize) + VOffset,
	}

	// Program header size and offset
	headers.elf_header.Phentsize = uint16(unsafe.Sizeof(headers.program_header))
	headers.elf_header.Phoff = uint64(headers.elf_header.Ehsize)

	// Entry point == (Ehdr + Phdr)+Offset
	headers.elf_header.Entry = uint64(headers.elf_header.Phentsize+headers.elf_header.Ehsize) + VOffset

	// build data buffers
	elf_data := new(bytes.Buffer)
	prg_data := new(bytes.Buffer)

	// write structure data into buffers
	binary.Write(elf_data, binary.LittleEndian, headers.elf_header)
	binary.Write(prg_data, binary.LittleEndian, headers.program_header)

	// now write content to file
	_, _ = fd.Write(elf_data.Bytes())
	_, _ = fd.Write(prg_data.Bytes())

	// xorb %al, %al
	// movb $60, %al
	// syscall
	//var dummy  = []byte{0x30,0xc0,0xb0,0x3c,0x0f,0x05}
	//_, _ = fd.Write(dummy)

	fmt.Printf("Custom ELF file '%s' created. All written opcodes will go there!\n", path)
	return fd, nil
}

func Write(fd *os.File, data []byte) error {
	_, err := fd.Write(data)
	if err != nil {
		return err
	}
	return nil
}

# gopdecode
Machine code decoding :symbols:
You can feed him with opcodes and will decode it using Go's x86asm package or create a custom ELF with them.

### Example
```bash
$ ./gopdecode
> 30c1
xor cl, al
> 90
nop
> history
0       30c1
1       90
> q
$
```
## Options
* `q` or `quit`: quit the application.
* ```set flavor=[att|go|intel]```: change output syntax (AT&T, Intel or Go Assembly).
* ```set output=(filepath)```: save output to file.
* ```set mode=[16|32|64]```: set arch mode (default is 64-bit).
* ```set colors```: toggle "pretty" printing.
* ```set json```: toggle print JSON opcode information.
* ```history```: show history of previous inputs (**.gop_history** file).

JSON output:
```json
> set json
> 8b55fc                                     
{
    "length": 3,
    "instruction": "mov",
    "args": [
        "edx",
        "dword ptr [rbp-0x4]"
    ]
}
>
```

## Create custom ELF executable (experimental)
* ```create [filepath]```: this will create a custom ELF executable with the opcodes that you feed to this tool. All headers are setted by default and with a predefined entry point.

### Example
```bash
$ ./gopdecode
> create elf_test
Custom ELF file 'elf_test' created.
> 30c0
xor al, al
> b03c
mov al, 0x3c
> 0f05
syscall
> q
$ strace ./elf_test
execve("./elf_test", ["./elf_test"], 0x7ffea74e25f0 /* 21 vars */) = 0
exit(0)                                 = ?
+++ exited with 0 +++
$ file elf_test
elf_test: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, corrupted section header size
```

## Credits
Based on [m2elf](https://github.com/XlogicX/m2elf)'s decoding functionality.

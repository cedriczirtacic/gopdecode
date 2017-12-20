# gopdecode
Machine code decoding :symbols:
You can feed him with opcodes and will decode it using Go's x86asm package.

### Example
```bash
$ ./gopdecode
> 30c1
xor cl, al
> 90
nop
> q
$
```
## Options
* `q` or `quit`: quit the application.
* ```set flavor=[att|go|intel]```: change output syntax (AT&T, Intel or Go Assembly).
* ```set output=(filepath)```: save output to file.

## Credits
Based on [m2elf](https://github.com/XlogicX/m2elf)'s decoding functionality.

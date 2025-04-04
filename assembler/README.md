# Hack assembler

The assembler translates assembly Hack programs (`.asm` files) into binary Hack programs (`.hack`
files). Run it like this:

    assembler program.asm

and it will create a file `program.hack` that can be loaded into the NAND2Tetris CPU emulator.

You can build the assembler binary with

    make

and run unit tests with

    make test

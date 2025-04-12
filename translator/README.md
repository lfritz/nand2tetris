# Hack translator

The translator translates Hack VM programs (`.vm` files) into Hack assembly programs (`.asm` files).
Run it like this:

    translator program.vm

and it will create a file `program.asm` that can be used as input to the Hack assembler.

You can build the translator binary with

    make

and run unit tests with

    make test

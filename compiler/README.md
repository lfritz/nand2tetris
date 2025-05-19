# Hack compiler

The compiler will translate Jack programs (`.jack` files) into Hack VM programs (`.vm` files). Right
now it only parses the code and produces a parse tree as XML. Run it on a single file:

    compiler program.jack

to produce a file `program.xml` or on a directory:

    compiler source

to parse all `.jack` files in the directory and produce a matching xml file for each.

You can build the compiler binary with

    make

and run the tests with

    ./test-tokens.sh
    ./test-syntax.sh

The test scripts need xmllint (part of [libxml2](https://gitlab.gnome.org/GNOME/libxml2)).

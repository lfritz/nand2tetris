#!/usr/bin/bash

set -e

compare() {
    echo "Comparing syntax for $1:"
    diff <(cat $1 | sed -e 's/^[ \t]*//' | xmllint --format - ) \
         <(cat $2 | sed -e 's/^[ \t]*//' | xmllint --format - )
    echo "OK"
    echo
}

make -s

./compiler -s test-parser/ArrayTest/Main.jack
compare test-parser/ArrayTest/Main{,Syntax}.xml

./compiler -s test-parser/ExpressionLessSquare
for f in $(ls test-parser/ExpressionLessSquare/*.jack)
do
    name=$(basename -s .jack $f)
    compare test-parser/ExpressionLessSquare/${name}{,Syntax}.xml
done

./compiler -s test-parser/Square
for f in $(ls test-parser/Square/*.jack)
do
    name=$(basename -s .jack $f)
    compare test-parser/Square/${name}{,Syntax}.xml
done

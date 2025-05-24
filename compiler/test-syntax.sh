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

./compiler -s testdata/ArrayTest/Main.jack
compare testdata/ArrayTest/Main{,Syntax}.xml

./compiler -s testdata/ExpressionLessSquare
for f in $(ls testdata/ExpressionLessSquare/*.jack)
do
    name=$(basename -s .jack $f)
    compare testdata/ExpressionLessSquare/${name}{,Syntax}.xml
done

./compiler -s testdata/Square
for f in $(ls testdata/Square/*.jack)
do
    name=$(basename -s .jack $f)
    compare testdata/Square/${name}{,Syntax}.xml
done

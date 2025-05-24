#!/usr/bin/bash

set -e

compare() {
    echo "Comparing tokens for $1:"
    diff <(xmllint --format $1) <(xmllint --format $2)
    echo "OK"
    echo
}

make -s

./compiler -t testdata/ArrayTest/Main.jack
compare testdata/ArrayTest/Main{,Tokens}.xml

./compiler -t testdata/ExpressionLessSquare
for f in $(ls testdata/ExpressionLessSquare/*.jack)
do
    name=$(basename -s .jack $f)
    compare testdata/ExpressionLessSquare/${name}{,Tokens}.xml
done

./compiler -t testdata/Square
for f in $(ls testdata/Square/*.jack)
do
    name=$(basename -s .jack $f)
    compare testdata/Square/${name}{,Tokens}.xml
done

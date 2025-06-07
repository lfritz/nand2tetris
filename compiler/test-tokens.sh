#!/usr/bin/bash

set -e

compare() {
    echo "Comparing tokens for $1:"
    diff <(xmllint --format $1) <(xmllint --format $2)
    echo "OK"
    echo
}

make -s

./compiler -t test-parser/ArrayTest/Main.jack
compare test-parser/ArrayTest/Main{,Tokens}.xml

./compiler -t test-parser/ExpressionLessSquare
for f in $(ls test-parser/ExpressionLessSquare/*.jack)
do
    name=$(basename -s .jack $f)
    compare test-parser/ExpressionLessSquare/${name}{,Tokens}.xml
done

./compiler -t test-parser/Square
for f in $(ls test-parser/Square/*.jack)
do
    name=$(basename -s .jack $f)
    compare test-parser/Square/${name}{,Tokens}.xml
done

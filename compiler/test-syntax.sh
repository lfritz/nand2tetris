#!/usr/bin/bash

set -e

compare() {
    diff <(xmllint --format $1) <(xmllint --format $2)
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

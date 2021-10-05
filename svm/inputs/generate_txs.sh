#!/bin/bash

for FILE in ./spawn/*.json
do
    INPUT=$(basename $FILE)
    OUTPUT=$INPUT.bin

    ../artifacts/svm-cli tx --tx-type=spawn --input=spawn/$INPUT --output=spawn/$OUTPUT
done
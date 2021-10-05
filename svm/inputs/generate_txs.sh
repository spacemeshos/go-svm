#!/bin/bash

for TX_TYPE in spawn call
do
    for FILE in ./$TX_TYPE/*.json
    do
        INPUT=$(basename $FILE)
        OUTPUT=$INPUT.bin
        ../artifacts/svm-cli tx --tx-type=$TX_TYPE --input=$TX_TYPE/$INPUT --output=$TX_TYPE/$OUTPUT
    done
done
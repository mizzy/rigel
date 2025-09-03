#!/bin/bash
export RIGEL_TEST_MODE=1
export PROVIDER=ollama

# Send the exact sequence
printf "aaaa\x0Abbbb\x0Acccc\x0Adddd\x0D" | ./rigel-debug --termflow

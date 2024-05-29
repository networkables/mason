#!/bin/sh

curl -o oui.txt https://standards-oui.ieee.org/oui/oui.txt

grep '(base 16)' oui.txt | awk -F'\t' '{ print $1,$3 }' | sed 's/     (base 16)//' > base16_oui.txt

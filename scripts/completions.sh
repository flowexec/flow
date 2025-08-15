#!/bin/sh
set -e
rm -rf scripts/completions
mkdir scripts/completions
for sh in bash zsh fish; do
	go run main.go completion "$sh" >"scripts/completions/flow.$sh"
done
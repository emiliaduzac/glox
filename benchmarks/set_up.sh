#!/usr/bin/env bash
set -e

mkdir -p benchmarks/otros
cd benchmarks/otros

# 1. clox y jlox -> los del libro
git clone https://github.com/munificent/craftinginterpreters
cd craftinginterpreters
make clox
make jlox
cd ..

# 2. plox -> tree-walk de la cátedra
git clone https://github.com/FdelMazo/plox
cd plox && uv sync && cd ..

# 3. ploxb -> bytecode de la cátedra
git clone -b fns-and-closures https://github.com/FdelMazo/ploxb
cd ploxb && uv sync && cd ..

# 4. golox -> bytecode de otro usuario también en Go
git clone https://github.com/rami3l/golox
cd golox
go build -o golox .
cd ../../..

# 5. glox de la entrega 1 -> tree-walk
git checkout entrega-1
go build -o glox-parcial ./cmd
git checkout -

# 6. glox actual -> tree-walk y bytecode
go build -o glox ./cmd
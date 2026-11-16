#!/usr/bin/env bash
set -e

CLOX="benchmarks/otros/craftinginterpreters/clox"
JLOX="benchmarks/otros/craftinginterpreters/jlox"
PLOX="benchmarks/otros/plox/.venv/bin/plox"
PLOXB="benchmarks/otros/ploxb/.venv/bin/ploxb"
GOLOX="benchmarks/otros/golox/golox"
PYTHON="benchmarks/otros/plox/.venv/bin/python"

RESULTADOS="benchmarks/resultados"
mkdir -p "$RESULTADOS"

# 1. fibonacci -> todas
hyperfine --warmup 1 \
  -n "clox" "$CLOX benchmarks/lox/fibonacci.lox" \
  -n "golox" "$GOLOX benchmarks/lox/fibonacci.lox" \
  -n "glox" "./glox --vm benchmarks/lox/fibonacci.lox" \
  -n "glox (entrega 1)" "./glox-parcial benchmarks/lox/fibonacci.lox" \
  -n "jlox" "$JLOX benchmarks/lox/fibonacci.lox" \
  -n "plox" "$PLOX benchmarks/lox/fibonacci.lox" \
  -n "ploxb" "$PLOXB benchmarks/lox/fibonacci.lox" \
  -n "Go" "benchmarks/go/fibonacci/fibonacci" \
  -n "Python" "$PYTHON benchmarks/python/fibonacci.py" \
  --export-json "$RESULTADOS/fibonacci.json" \
  --export-markdown "$RESULTADOS/fibonacci.md"

# 2. minsky -> todas
hyperfine --warmup 1 \
  -n "clox" "$CLOX benchmarks/lox/minsky.lox" \
  -n "golox" "$GOLOX benchmarks/lox/minsky.lox" \
  -n "glox" "./glox --vm benchmarks/lox/minsky.lox" \
  -n "glox (entrega 1)" "./glox-parcial benchmarks/lox/minsky.lox" \
  -n "jlox" "$JLOX benchmarks/lox/minsky.lox" \
  -n "plox" "$PLOX benchmarks/lox/minsky.lox" \
  -n "ploxb" "$PLOXB benchmarks/lox/minsky.lox" \
  -n "Go" "benchmarks/go/minsky/minsky" \
  -n "Python" "$PYTHON benchmarks/python/minsky.py" \
  --export-json "$RESULTADOS/minsky.json" \
  --export-markdown "$RESULTADOS/minsky.md"

# 3. fizzbuzz -> solo las que soportan %
hyperfine --warmup 1 \
  -n "glox" "./glox --vm benchmarks/lox/fizzbuzz.lox" \
  -n "glox (entrega 1)" "./glox-parcial benchmarks/lox/fizzbuzz.lox" \
  -n "plox" "$PLOX benchmarks/lox/fizzbuzz.lox" \
  -n "ploxb" "$PLOXB benchmarks/lox/fizzbuzz.lox" \
  -n "Go" "benchmarks/go/fizzbuzz/fizzbuzz" \
  -n "Python" "$PYTHON benchmarks/python/fizzbuzz.py" \
  --export-json "$RESULTADOS/fizzbuzz.json" \
  --export-markdown "$RESULTADOS/fizzbuzz.md"

# 4. go run contra el binario
hyperfine --warmup 1 \
  -n "glox (binario)" "./glox --vm benchmarks/lox/fibonacci.lox" \
  -n "glox (go run)" "go run ./cmd --vm benchmarks/lox/fibonacci.lox" \
  --export-json "$RESULTADOS/go-run.json" \
  --export-markdown "$RESULTADOS/go-run.md"
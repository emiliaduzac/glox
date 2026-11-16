#!/usr/bin/env bash
set -e

CLOX="benchmarks/otros/craftinginterpreters/clox"

RESULTADOS="benchmarks/resultados"
mkdir -p "$RESULTADOS"

hyperfine --warmup 1 \
  -n "glox (bound method)" "./glox --vm benchmarks/lox/methods.lox" \
  -n "glox (invoke)" "./glox --vm benchmarks/lox/methods_optimized.lox" \
  -n "clox (bound method)" "$CLOX benchmarks/lox/methods.lox" \
  -n "clox (invoke)" "$CLOX benchmarks/lox/methods_optimized.lox" \
  --export-json "$RESULTADOS/methods.json" \
  --export-markdown "$RESULTADOS/methods.md"
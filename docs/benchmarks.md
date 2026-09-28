# Benchmarks de glox

Para cada benchmark se realizaron 5 iteraciones.

## 1. Mediciones internas con `clock()`

Medición del tiempo neto de ejecución dentro de Lox mediante la función nativa `clock()`. No incluye el arranque de Go ni del sistema operativo.

| Benchmark | Mínimo (ms) | Máximo (ms) | Promedio (ms) | Desvío Estándar ($\sigma$) |
|---|---|---|---|---|
| `fib-clock.lox` | 0.00 | 0.00 | **0.00** | 0.00 ms |
| `minsky-clock.lox` | 0.00 | 0.00 | **0.00** | 0.00 ms |
| `fizzbuzz-clock.lox` | 0.00 | 0.00 | **0.00** | 0.00 ms |

## 2. Mediciones externas con `time go run ./cmd`

Tiempo total medido desde la terminal con `time`. Incluye la compilación al vuelo de Go (`go run`), la inicialización del runtime y la ejecución de Lox.

| Benchmark | Mínimo (ms) | Máximo (ms) | Promedio (ms) | Desvío Estándar ($\sigma$) |
|---|---|---|---|---|
| `fibonacci.lox` | 0.00 | 0.00 | **0.00** | 0.00 ms |
| `minsky.lox` | 0.00 | 0.00 | **0.00** | 0.00 ms |
| `fizzbuzz.lox` | 0.00 | 0.00 | **0.00** | 0.00 ms |

## 3. Mediciones externas con binario precompilado (`time ./glox`)

Tiempo total medido desde la terminal ejecutando el binario compilado con `go build -o glox ./cmd`. Elimina la compilación al vuelo de Go, midiendo únicamente su ejecución.

| Benchmark | Mínimo (ms) | Máximo (ms) | Promedio (ms) | Desvío Estándar ($\sigma$) |
|---|---|---|---|---|
| `fibonacci.lox` | 0.00 | 0.00 | **0.00** | 0.00 ms |
| `minsky.lox` | 0.00 | 0.00 | **0.00** | 0.00 ms |
| `fizzbuzz.lox` | 0.00 | 0.00 | **0.00** | 0.00 ms |

## 4. Comparativa interna glox

| Benchmark | 1. `clock()` interno (ms) | 2. `go run ./cmd` (ms) | 3. Binario `./glox` (ms) |
|---|---|---|---|
| `fibonacci` | 0.00 | 0.00 | 0.00 |
| `minsky` | 0.00 | 0.00 | 0.00 |
| `fizzbuzz` | 0.00 | 0.00 | 0.00 |

---

# Benchmarks de plox

Viendo que en `plox` no tiene la función nativa `clock()` y que no soporta el operador `%`, la medición la hicmos con `time plox benchmarks/<script>.lox` únicamente sobre los compatibles (`fibonacci.lox` y `minsky.lox`).

| Benchmark | Mínimo (ms) | Máximo (ms) | Promedio (ms) | Desvío Estándar ($\sigma$) |
|---|---|---|---|---|
| `fibonacci.lox` | 8688.00 | 8823.00 | **8749.00** | 57.70 ms |
| `minsky.lox` | 1581.00 | 1618.00 | **1595.80** | 14.31 ms |

---

# Comparativa glox vs plox

Teniendo en cuenta que en el uso común es ejecutar directamente desde la terminal `go run ./cmd <archivo>` en `glox` y `plox <archivo>` en `plox`, la comparativa la hacemos midiendo con `time` la ejecución de ambos.

Los resultados son los siguientes:

| Benchmark | `plox` Promedio (ms) | `glox` (`go run`) Promedio (ms) | Speedup (`glox` vs `plox`) |
|---|---|---|:---:|
| `fibonacci` | 8749.00 | 0.00 | - |
| `minsky` | 1595.80 | 0.00 | - |


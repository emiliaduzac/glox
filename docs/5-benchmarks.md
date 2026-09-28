# Benchmarks de glox

Para cada benchmark se realizaron 5 iteraciones.

## 1. Mediciones internas con `clock()`

Medición del tiempo neto de ejecución dentro de Lox mediante la función nativa `clock()`. No incluye el arranque de Go ni del sistema operativo.

Para correr las pruebas:
```bash
go run ./cmd benchmarks/fib-clock.lox
```
```bash
go run ./cmd benchmarks/minsky-clock.lox
```
```bash
go run ./cmd benchmarks/fizzbuzz-clock.lox
```

| Benchmark | Mínimo (ms) | Máximo (ms) | Promedio (ms) | Desvío Estándar ($\sigma$) |
|---|---|---|---|---|
| `fib-clock.lox` | 42.674 | 46.998 | **45.699** | 1.547 ms |
| `minsky-clock.lox` | 4.812 | 6.103 | **5.816** | 0.503 ms |
| `fizzbuzz-clock.lox` | 3.346 | 3.898 | **3.663** | 0.178 ms |

## 2. Mediciones externas con `time go run ./cmd`

Tiempo total medido desde la terminal con `time`. Incluye la compilación al vuelo de Go (`go run`), la inicialización del runtime y la ejecución de Lox.

Para correr las pruebas:
```bash
time go run ./cmd benchmarks/fibonacci.lox
```
```bash
time go run ./cmd benchmarks/minsky.lox
```
```bash
time go run ./cmd benchmarks/fizzbuzz.lox
```


| Benchmark | Mínimo (ms) | Máximo (ms) | Promedio (ms) | Desvío Estándar ($\sigma$) |
|---|---|---|---|---|
| `fibonacci.lox` | 108.00 | 203.00 | **127.80** | 42.04 ms |
| `minsky.lox` | 47.00 | 67.00 | **62.60** | 8.73 ms |
| `fizzbuzz.lox` | 62.00 | 79.00 | **67.00** | 6.86 ms |

## 3. Mediciones externas con binario precompilado (`time ./glox`)

Tiempo total medido desde la terminal ejecutando el binario compilado con `go build -o glox ./cmd`. Elimina la compilación al vuelo de Go, midiendo únicamente su ejecución.

Para correr las pruebas primero:
```bash
go build -o glox ./cmd
```
Después:
```bash
time ./glox benchmarks/fibonacci.lox
```
```bash
time ./glox benchmarks/minsky.lox
```
```bash
time ./glox benchmarks/fizzbuzz.lox
```

| Benchmark | Mínimo (ms) | Máximo (ms) | Promedio (ms) | Desvío Estándar ($\sigma$) |
|---|---|---|---|---|
| `fibonacci.lox` | 73.00 | 78.00 | **76.20** | 1.92 ms |
| `minsky.lox` | 9.00 | 37.00 | **22.40** | 9.91 ms |
| `fizzbuzz.lox` | 12.00 | 17.00 | **16.00** | 2.24 ms |

## 4. Comparativa interna glox

| Benchmark | 1. `clock()` interno (ms) | 2. `go run ./cmd` (ms) | 3. Binario `./glox` (ms) |
|---|---|---|---|
| `fibonacci` | 45.70 | 127.80 | 76.20 |
| `minsky` | 5.82 | 62.60 | 22.40 |
| `fizzbuzz` | 3.66 | 67.00 | 16.00 |

---

# Benchmarks de plox

Viendo que en `plox` no tiene la función nativa `clock()` y que no soporta el operador `%`, la medición la hicmos con `time plox benchmarks/<script>.lox` únicamente sobre los compatibles (`fibonacci.lox` y `minsky.lox`).

Para correr las pruebas:
```bash
time plox benchmarks/fibonacci.lox
```
```bash
time plox benchmarks/minsky.lox
```

| Benchmark | Mínimo (ms) | Máximo (ms) | Promedio (ms) | Desvío Estándar ($\sigma$) |
|---|---|---|---|---|
| `fibonacci.lox` | 1816.00 | 1837.00 | **1824.40** | 9.50 ms |
| `minsky.lox` | 345.00 | 372.00 | **361.80** | 10.33 ms |

---

# Comparativa glox vs plox

Teniendo en cuenta que en el uso común es ejecutar directamente desde la terminal `go run ./cmd <archivo>` en `glox` y `plox <archivo>` en `plox`, la comparativa la hacemos midiendo con `time` la ejecución de ambos.

Los resultados son los siguientes:

| Benchmark | `plox` Promedio (ms) | `glox` (`go run`) Promedio (ms) | Speedup (`glox` vs `plox`) |
|---|---|---|:---:|
| `fibonacci` | 1824.40 | 127.80 | **14.3x más rápido** |
| `minsky` | 361.80 | 62.60 | **5.8x más rápido** |

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
| `fib-clock.lox` | 146.92 | 199.11 | **166.63** | 21.73 ms |
| `minsky-clock.lox` | 19.52 | 20.64 | **20.24** | 0.46 ms |
| `fizzbuzz-clock.lox` | 10.51 | 13.72 | **11.86** | 1.17 ms |

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
| `fibonacci.lox` | 230.00 | 316.00 | **253.40** | 35.54 ms |
| `minsky.lox` | 91.00 | 98.00 | **95.00** | 2.74 ms |
| `fizzbuzz.lox` | 83.00 | 90.00 | **85.60** | 2.70 ms |

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
| `fibonacci.lox` | 151.00 | 308.00 | **213.40** | 57.75 ms |
| `minsky.lox` | 24.00 | 27.00 | **25.20** | 1.30 ms |
| `fizzbuzz.lox` | 15.00 | 17.00 | **15.60** | 0.89 ms |

## 4. Comparativa interna glox

| Benchmark | 1. `clock()` interno (ms) | 2. `go run ./cmd` (ms) | 3. Binario `./glox` (ms) |
|---|---|---|---|
| `fibonacci` | 166.63 | 253.40 | 213.40 |
| `minsky` | 20.24 | 95.00 | 25.20 |
| `fizzbuzz` | 11.86 | 85.60 | 15.60 |

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
| `fibonacci.lox` | 8688.00 | 8823.00 | **8749.00** | 57.70 ms |
| `minsky.lox` | 1581.00 | 1618.00 | **1595.80** | 14.31 ms |

---

# Comparativa glox vs plox

Teniendo en cuenta que en el uso común es ejecutar directamente desde la terminal `go run ./cmd <archivo>` en `glox` y `plox <archivo>` en `plox`, la comparativa la hacemos midiendo con `time` la ejecución de ambos.

Los resultados son los siguientes:

| Benchmark | `plox` Promedio (ms) | `glox` (`go run`) Promedio (ms) | Speedup (`glox` vs `plox`) |
|---|---|---|:---:|
| `fibonacci` | 8749.00 | 253.40 | **34.5x más rápido** |
| `minsky` | 1595.80 | 95.00 | **16.8x más rápido** |

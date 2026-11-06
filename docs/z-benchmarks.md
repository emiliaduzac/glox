# Benchmarks

## Introducción

El objetivo de este documento es evaluar el rendimiento de _glox_, nuestra implementación de Lox en Go, y compararlo con _plox_, la implementación en Python de la cátedra.
Primero realizamos las mediciones en _glox_, luego en _plox_ y por último comparamos.

> TO DO: agregar comparaciones con otras implementaciones de lox en go, quizá contra go.

## Entorno utilizado

- Máquina: Apple M4, 16GB RAM
- SO: macOS 15.3
- Go: 1.25.0
- Python (plox): 3.14.7

## Metodología y programas

Utilizamos tres programas, realizando 5 iteraciones en cada caso y reportando mínimo, máximo, promedio y desvío estándar:

- `fibonacci.lox`: fibonacci recursivo (n = 25). Pone a prueba llamadas a funciones y recursión.
- `minsky.lox`: simula una máquina de registros con un `while` (n = 20000). Pone a prueba ciclos, asignaciones y condicionales.
- `fizzbuzz.lox`: FizzBuzz acumulando en un string (n = 5000). Pone a prueba ciclos, condicionales, el operador `%` y concatenación de strings.

→ **glox** se mide de tres formas, para separar el tiempo de ejecución de Lox del de compilar y arrancar Go: con `clock()` interno, con `time go run ./cmd` y con `time ./glox`.

→ **plox** se mide solo con `time`, porque no tiene `clock()`, y solo con `fibonacci` y `minsky`, porque no soporta el operador `%`.

Los programas se encuentran en la carpeta **benchmarks**.∫

## Resultados de glox

### 1. Mediciones internas con `clock()`

En primer lugar, tomamos la medición del tiempo neto de ejecución dentro de Lox mediante la función nativa `clock()` (que devuelve segundos, pero los convertimos a ms para la comparación). Esto no incluye el arranque de Go ni del sistema operativo.

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

### 2. Mediciones externas con `time go run ./cmd`

Luego, tomamos el tiempo total medido desde la terminal con `time`. Incluye la compilación al vuelo de Go (`go run`), la inicialización del runtime y la ejecución de Lox.

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

### 3. Mediciones externas con binario precompilado (`time ./glox`)

Por último, medimos el tiempo total desde la terminal ejecutando el binario compilado con `go build -o glox ./cmd`. Esto elimina la compilación al vuelo de Go, midiendo únicamente la ejecución del binario.

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

### Comparativa interna glox

En esta sección, comparamos el promedio de las 3 mediciones anteriores, para cada programa utilizado:

| Benchmark | 1. `clock()` interno (ms) | 2. `go run ./cmd` (ms) | 3. Binario `./glox` (ms) |
|---|---|---|---|
| `fibonacci` | 45.70 | 127.80 | 76.20 |
| `minsky` | 5.82 | 62.60 | 22.40 |
| `fizzbuzz` | 3.66 | 67.00 | 16.00 |

Como podemos ver, en los 3 programas la medición mediante `go run ./cmd` fue la más larga. La diferencia entre esta y el binario es de unos 40-50 ms en los tres casos, que lo atribuimos a la compilación. Por otro lado, `clock()` es siempre la más corta porque mide solo la ejecución de Lox.

## Resultados de plox

Como se mencionó al inicio, para plox la medición fue realizada únicamente con `time plox benchmarks/<script>.lox` sobre los programas compatibles (`fibonacci.lox` y `minsky.lox`).

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

## Comparativa glox vs plox

Dado que el uso más común es ejecutar directamente desde la terminal `go run ./cmd <archivo>` en glox y `plox <archivo>` en plox, la comparativa se hace midiendo la ejecución de ambos con `time`, que es la única medición posible para plox.

| Benchmark | `plox` Promedio (ms) | `glox` (`go run`) Promedio (ms) | Speedup (`glox` vs `plox`) |
|---|---|---|:---:|
| `fibonacci` | 1824.40 | 127.80 | **14.3x más rápido** |
| `minsky` | 361.80 | 62.60 | **5.8x más rápido** |

## Análisis y conclusiones

Como se ve en las secciones 2 y 3, correr glox con `go run` trae un costo fijo: **compilar el código**, que suma unos 40-50 ms. En programas cortos como `fizzbuzz` o `minsky`, esto es lo que termina dominando: casi todo el tiempo se va en la compilación y no en la ejecución de Lox. Con `clock()` la diferencia es todavía más drástica: `fizzbuzz` tarda unos 3.6 ms y `fibonacci` de 25, pese a su recursión, unos 45 ms.

En cuanto a la comparativa de glox y plox, es notoria la diferencia del tiempo, particularmente en `fibonacci`, donde el orden de magnitud cambia de los milisegundos a segundos. Una de las causas es que Go es **compilado**, mientras que Python es **interpretado**. Entonces, como plox corre sobre Python, termina siendo un intérprete corriendo sobre otro intérprete, y eso lo hace más lento frente a nuestro glox.

En resumen, glox resulta entre 5.8 y 14.3 veces más rápido que plox en los programas medidos, aun incluyendo el costo de compilar con `go run`. Si se lo ejecuta como binario precompilado, la diferencia es todavía mayor ya que se elimina ese costo fijo de compilar. Además, la diferencia depende del tipo de programa: es más marcada en `fibonacci`, que hace mucha recursión y es el de mayor duración, que en `minsky`, donde el costo de compilar pesa más sobre un programa corto.
# Benchmarks

## Introducción

El objetivo de este documento es evaluar el rendimiento de _glox_, nuestra implementación de Lox en Go, y compararlo con otras implementaciones del mismo lenguaje:

| N | Implementación | Lenguaje | Técnica | Origen |
|---|---|---|---|---|
| 1 | **glox** | Go | bytecode (`--vm`) | nuestra versión final |
| 2 | **glox (entrega 1)** | Go | tree-walk | nuestra entrega parcial ([`entrega-1`](https://github.com/emiliaduzac/glox/releases/tag/entrega-1)) |
| 3 | **clox** | C | bytecode | [Crafting Interpreters](https://github.com/munificent/craftinginterpreters) |
| 4 | **jlox** | Java | tree-walk | [Crafting Interpreters](https://github.com/munificent/craftinginterpreters) |
| 5 | **plox** | Python | tree-walk | de la cátedra ([plox](https://github.com/FdelMazo/plox)) |
| 6 | **ploxb** | Python | bytecode | de la cátedra ([ploxb](https://github.com/FdelMazo/ploxb), rama `fns-and-closures`) |
| 7 | **golox** | Go | bytecode | de otro autor ([rami3l/golox](https://github.com/rami3l/golox)) |

Además, para comparar contra otros lenguajes, escribimos los mismos programas en:

| N | Lenguaje | Cómo se ejecuta | Por qué lo elegimos |
|---|---|---|---|
| 8 | **Go** | compilado a código de máquina (`go build`) | es el lenguaje en el que están hechos glox y golox |
| 9 | **Python** | CPython | es el lenguaje en el que están hechos plox y ploxb |

## Requisitos y set up

Para correr los benchmarks hace falta tener instalado:

- **Go**, para glox, golox y comparar contra Go mismo
- **un compilador de C y `make`**, para clox
- **Java (JDK)**, para jlox
- **[uv](https://docs.astral.sh/uv/)** y Python 3.12 o superior, para plox, ploxb y Python mismo
- **hyperfine**

El script de set up descarga y compila todas las implementaciones en `benchmarks/otros`, y compila glox y su versión de la entrega 1 (`glox-parcial`) en la raíz del repo. Se corre desde la raíz:

```bash
./benchmarks/set_up.sh
```

Para borrar todo lo que descargó y compiló:

```bash
./benchmarks/clean_up.sh
```

## Entorno

Las mediciones y resultados mostrados se obtuvieron en el siguiente entorno:

- **Máquina**: MacBook Air, Apple M4, 16 GB de RAM
- **Sistema operativo**: macOS 15.3

| Herramienta | Versión | Usada por |
|---|---|---|
| Go | 1.25.0 | glox, glox (entrega 1), golox y los programas en Go |
| Apple clang | 17.0.0 | clox |
| OpenJDK (Temurin) | 21.0.10 | jlox |
| CPython | 3.12.14 | plox, ploxb y los programas en Python |
| hyperfine | 2.0.0 | — |

## Metodología

Para medir, utilizamos la herramienta recomendada por la cátedra: [hyperfine](https://github.com/sharkdp/hyperfine), que ejecuta cada comando varias veces y reporta el **promedio**, el **desvío estándar**, el **mínimo** y el **máximo** del tiempo y también la **memoria máxima** que usó el proceso.

hyperfine mide el **tiempo total del proceso**, desde que arranca hasta que termina, incluyendo el arranque de cada implementación (el runtime de Go, la JVM o el intérprete de Python). Por eso medimos binarios ya compilados y no `go run`, que además compilaría glox en cada ejecución.

- **Corridas**: cada comando se ejecuta una vez sin medir (`--warmup 1`), y después hyperfine decide cuántas veces medirlo: al menos 10 corridas y al menos 3 segundos en total. Así, las implementaciones rápidas se miden cientos de veces, y las más lentas, 10.
- **Salida**: hyperfine descarta lo que imprime cada programa, así que no se mide el tiempo de escribir en la consola.

Los resultados se exportan a JSON y markdown (`--export-json` y `--export-markdown`) en `benchmarks/resultados/`, y los gráficos se generan a partir del JSON.

Hacemos cuatro comparaciones:

1. **glox contra otras implementaciones de Lox**: los mismos programas con las 7 implementaciones.
2. **La evolución de glox**: la versión tree-walk de la entrega 1 contra la versión final con bytecode. Además, comparamos ejecutar el binario compilado contra `go run`, que compila glox en cada ejecución. Vale la pena mencionar que la versión final sin `--vm` también es tree-walk, pero usamos el tag de la entrega 1 para medir exactamente lo que habíamos entregado
3. **Lox contra otros lenguajes**: los mismos programas escritos en **Go** y en **Python**. Elegimos esos dos porque son los lenguajes en los que están hechas glox, golox, plox y ploxb.
4. **Optimización de métodos**: TO DO.

Para correr **todas** las mediciones:

```bash
./benchmarks/run.sh
```

## Programas

Usamos tres programas, cada uno enfocado en una parte distinta del lenguaje. `fibonacci` y `minsky` corren en las 7 implementaciones, así que solo usan lo que todas soportan: no usan `%`, `break` ni `continue`, ni clases.

- `fibonacci.lox`: fibonacci recursivo (n = 30). Pone a prueba llamadas a funciones y recursión.
- `minsky.lox`: simula una máquina de registros con un `while` (n = 200000). Pone a prueba ciclos, asignaciones y condicionales.
- `fizzbuzz.lox`: FizzBuzz acumulando en un string (n = 100000). Pone a prueba ciclos, condicionales, el operador `%` y concatenación de strings.

El operador `%` es una extensión de la cátedra: lo tienen glox, plox y ploxb, pero no existe en el Lox del libro, así que clox y jlox no lo soportan. Por eso `fizzbuzz` solo se compara entre las implementaciones que lo tienen.

## Hipótesis

Antes de medir, definimos algunas hipótesis que esperamos ver en los resultados:

1. **clox va a ser la más rápida entre las implementaciones de Lox**: es bytecode (gana a las versiones AST) y está escrita en C (gana a las demás versiones de bytecode en otros lenguajes)
2. **Bytecode le va a ganar a tree-walk** en los dos lenguajes en los que tenemos ambas versiones: glox VM contra glox TW (entrega 1) en Go, y ploxb contra plox en Python. La razón para implementar esta segunda versión fue para mejorar la velocidad, por lo que sería esperable que se cumpla.
3. **Las implementaciones en Python van a ser las más lentas**, porque son un intérprete corriendo sobre otro intérprete.
4. **glox VM y golox van a estar cerca**: ambos son bytecode compilers de Lox. Nos resulta interesante ver que diferencias podrían mostrar. Si las hay, deberían ser optimizaciones que alguno de los dos agregó, o decisiones de diseño.
5. **Compilar Go en cada corrida (`go run`) suma un costo fijo**. Ya lo vimos en los benchmarks de la primer entrega, donde sumaba unos 40 a 50 ms. Esperamos que se mantenga igual, porque ese costo depende de compilar glox y no del programa.

## Resultados

### Distintas implementaciones de Lox

**Tiempo** es el promedio ± el desvío estándar de las corridas y **Memoria**, el promedio de la memoria máxima que usó el proceso. Los datos completos de cada corrida (incluyendo el uso de CPU y las instrucciones ejecutadas) están en `benchmarks/resultados/`.

**1. Fibonacci**

| Implementación | Tiempo (ms) | Factor vs glox | Memoria (MB) |
|---|---|---|---|
| clox | **57.67** ± 15.83 | 4.40x más rápido | 1.14 |
| golox | **202.09** ± 5.17 | 1.26x más rápido | 10.53 |
| glox | **253.83** ± 1.44 | — | 9.45 |
| glox (entrega 1) | **497.30** ± 4.01 | 1.96x más lento | 11.90 |
| jlox | **360.88** ± 149.12 | 1.42x más lento | 198.25 |
| plox | **27720.15** ± 177.25 | 109.21x más lento | 29.03 |
| ploxb | **17300.09** ± 233.48 | 68.16x más lento | 29.14 |

**2. Minsky**

| Implementación | Tiempo (ms) | Factor vs glox | Memoria (MB) |
|---|---|---|---|
| clox | **5.83** ± 2.10 | 5.61x más rápido | 1.13 |
| golox | **29.09** ± 3.50 | 1.12x más rápido | 6.41 |
| glox | **32.72** ± 0.27 | — | 5.45 |
| glox (entrega 1) | **52.57** ± 5.05 | 1.61x más lento | 9.21 |
| jlox | **109.76** ± 14.67 | 3.35x más lento | 69.00 |
| plox | **3992.82** ± 37.90 | 122.03x más lento | 28.82 |
| ploxb | **2788.06** ± 417.13 | 85.21x más lento | 29.05 |

**3. Fizzbuzz**

| Implementación | Tiempo (ms) | Factor vs glox | Memoria (MB) |
|---|---|---|---|
| glox | **341.11** ± 5.01 | — | 19.40 |
| glox (entrega 1) | **297.47** ± 6.68 | 1.15x más rápido | 36.95 |
| plox | **2701.40** ± 20.10 | 7.92x más lento | 32.29 |
| ploxb | **1893.95** ± 28.26 | 5.55x más lento | 32.49 |

**Análisis**

Tanto en `fibonacci` como en `minsky`, clox fue la más rápida, lo cual era esperado. Otra cosa que destaca es que golox es siempre más rápido que nuestro glox (entre 1.12x y 1.26x), aunque usa un poco más de memoria. Como ambos son bytecode, están hechos en Go y se basan en el mismo libro, la diferencia podría venir de decisiones de implementación, por ejemplo nosotros nos separamos varias veces del libro apoyándonos en utilidades de Go.

También se cumple que ploxb es más rápido que plox, igual que glox VM es más rápido que glox TW: el bytecode ayuda a agilizar, tal como anticipamos al implementar [bytecode](../docs/5-bytecode_and_vm.md). La excepción es `fizzbuzz`, donde la entrega 1 es más rápida, lo cual nos sorprendió mucho y no identificamos bien su causa.

Las implementaciones en Python fueron las más lentas por mucho, entre 68x y 122x más lentas que glox. En cuanto a jlox, en `minsky` que es corto, es 3.35x más lento que glox, mientras que en `fibonacci` que es más largo, su mediana (248.5 ms) es prácticamente igual a la de glox. Es algo parecido a lo que vimos en los benchmarks de la entrega 1 con `go run`: un costo fijo que pesa más cuanto más corto es el programa. En ese caso era compilar glox; acá creemos que es el arranque de la JVM. Además, el JIT de Java que optimiza el código mientras corre, puede ser una razón que ayude.

### Evolución de glox

### Lox contra otros lenguajes

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
chmod +x ./benchmarks/set_up.sh
./benchmarks/set_up.sh
```

Para borrar todo lo que descargó y compiló:

```bash
chmod +x ./benchmarks/clean_up.sh
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

Para medir, utilizamos la herramienta recomendada por la cátedra: [hyperfine](https://github.com/sharkdp/hyperfine), que ejecuta cada comando varias veces y reporta el **promedio**, el **desvío estándar**, el **mínimo** y el **máximo** del tiempo, y también la **memoria máxima** que usó el proceso.

hyperfine mide el **tiempo total del proceso**, desde que arranca hasta que termina, incluyendo el arranque de cada implementación (el runtime de Go, la JVM o el intérprete de Python). Por eso medimos binarios ya compilados y no `go run`, que además compilaría glox en cada ejecución.

- **Corridas**: cada comando se ejecuta una vez sin medir (`--warmup 1`), y después hyperfine decide cuántas veces medirlo: al menos 10 corridas y al menos 3 segundos en total. Así, las implementaciones rápidas se miden cientos de veces, y las más lentas, 10.
- **Salida**: hyperfine descarta lo que imprime cada programa, así que no se mide el tiempo de escribir en la consola.

Los resultados se exportan a JSON y markdown (`--export-json` y `--export-markdown`) en `benchmarks/resultados/`, y los gráficos se generan a partir del JSON.

Hacemos cuatro comparaciones:

1. **glox contra otras implementaciones de Lox**: los mismos programas con las 7 implementaciones.
2. **La evolución de glox**: la versión tree-walk de la entrega 1 contra la versión final con bytecode. Además, comparamos ejecutar el binario compilado contra `go run`, que compila glox en cada ejecución. Vale la pena mencionar que la versión final sin `--vm` también es tree-walk, pero usamos el tag de la entrega 1 para medir exactamente lo que habíamos entregado
3. **Lox contra otros lenguajes**: los mismos programas escritos en **Go** y en **Python**. Elegimos esos dos porque son los lenguajes en los que están hechas glox, golox, plox y ploxb.
4. **Optimización de métodos**: TO DO.

## Programas

Usamos tres programas, cada uno enfocado en una parte distinta del lenguaje. Como tienen que correr en las 7 implementaciones, solo usan lo que todas soportan: no usamos `%`, `break` ni `continue`, ni clases.

- `fibonacci.lox`: fibonacci recursivo (n = 30). Pone a prueba llamadas a funciones y recursión.
- `minsky.lox`: simula una máquina de registros con un `while` (n = <n>). Pone a prueba ciclos, asignaciones y condicionales.
- `fizzbuzz.lox`: FizzBuzz acumulando en un string (n = <n>). Pone a prueba ciclos, condicionales y concatenación de strings.

Los tamaños los elegimos para que la implementación más rápida (clox) tarde lo suficiente como para que su arranque pese poco y la más lenta (plox) no tarde tanto como para que medirla sea ilógico. <algunos a definir todavia>
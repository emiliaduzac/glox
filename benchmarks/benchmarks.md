# Benchmarks

## Introducción

El objetivo de este documento es evaluar el rendimiento de _glox_, nuestra implementación de Lox en Go, y compararlo con otras implementaciones del mismo lenguaje:

| Implementación | Lenguaje | Técnica | Origen |
|---|---|---|---|
| **glox** | Go | tree-walk y bytecode (`--vm`) | nuestra versión final |
| **glox (entrega 1)** | Go | tree-walk | nuestra entrega parcial ([`entrega-1`](https://github.com/emiliaduzac/glox/releases/tag/entrega-1)) |
| **clox** | C | bytecode | [Crafting Interpreters](https://github.com/munificent/craftinginterpreters) |
| **jlox** | Java | tree-walk | [Crafting Interpreters](https://github.com/munificent/craftinginterpreters) |
| **plox** | Python | tree-walk | de la cátedra ([plox](https://github.com/FdelMazo/plox)) |
| **ploxb** | Python | bytecode | de la cátedra ([ploxb](https://github.com/FdelMazo/ploxb), rama `fns-and-closures`) |
| **golox** | Go | bytecode | de otro autor ([rami3l/golox](https://github.com/rami3l/golox)) |

## Requisitos y set up

Para correr los benchmarks hace falta tener instalado:

- **Go**, para glox y golox
- **un compilador de C y `make`**, para clox
- **Java (JDK)**, para jlox
- **[uv](https://docs.astral.sh/uv/)** y Python 3.12 o superior, para plox y ploxb
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
| Go | 1.25.0 | glox, glox (entrega 1), golox |
| Apple clang | 17.0.0 | clox |
| OpenJDK (Temurin) | 21.0.10 | jlox |
| CPython | 3.12.14 | plox, ploxb |
| hyperfine | 2.0.0 | — |

## Metodología

Para medir, utilizamos la herramienta recomendada por la cátedra: [hyperfine](https://github.com/sharkdp/hyperfine), que ejecuta cada comando indicado varias veces y reporta el **promedio**, el **desvío estándar**, el **mínimo** y el **máximo**.

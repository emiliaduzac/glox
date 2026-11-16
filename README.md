# Lenguajes y compiladores - TP Lox

Repositorio para el trabajo práctico de la materia Lenguajes y Compiladores.

Implementación de una variante propia del lenguaje [Lox](https://craftinginterpreters.com/), combinando las dos aproximaciones vistas en clase ([tree-walk interpreter](https://github.com/FdelMazo/plox) y [bytecode compiler](https://github.com/FdelMazo/ploxb)). Desarrollado en [Go](https://go.dev/).

## Tabla de contenidos

- [Integrantes](#integrantes)
- [Documentación de la implementación](#documentación-de-la-implementación)
- [Cómo correr lox](#cómo-correr-lox)
- [Tests unitarios](#tests-unitarios)
- [Tests de la cátedra](#tests-de-la-cátedra)
- [To Do List](#to-do-list)

## Integrantes

- [Emilia Duzac](https://github.com/emiliaduzac) - 110307
- [Marcos Navarro](https://github.com/NVRar) - 106974

## Documentación de la implementación

Como la implementación será desarrollada en etapas, decidimos separar la documentación en un documento por etapa, para facilitar su lectura y comprensión. Toda la documentación técnica se encuentra bajo la carpeta **docs**, y la de los **extras** y **benchmarks** en sus propias carpetas.

> Indice de documentación

- [Scanner](docs/1-scanner.md)
- [Parser](docs/2-parser.md)
- [Interpreter](docs/3-interpreter.md)
- [Resolver](docs/4-resolver.md)
- [Bytecode & Virtual Machine](docs/5-bytecode_and_vm.md)
- [Compiler](docs/6-compiler.md)
- [Extras propios](docs/7-extra.md)
- [Benchmarks](benchmarks/benchmarks.md)
- [Uso Real de Glox](real-use/uso_real.md)

## Cómo correr lox

Lox se puede ejecutar de dos formas:

1. Ejecutando un archivo

```bash
go run ./cmd <archivo.lox>
```

2. De forma interactiva (REPL)

```bash
go run ./cmd
```

Si se quiere especificar el modo (como `--scanning`, `--parsing`, etc):

```bash
go run ./cmd --<mode>
```

Por default se ejecuta el tree-walk de la primera parte. Para usar el compilador a bytecode y la VM se agrega la flag `--vm`:

```bash
go run ./cmd --vm <archivo.lox>
```

Combinaciones de flags permitidas:

- Ninguna flag: tree-walk.
- Una sola entre `--scanning`, `--parsing` o `--resolve`. No se pueden combinar entre sí ni con `--vm` o `--debug`.
- `--vm`, solo o seguido de `--debug`. `--debug` no se puede usar sin `--vm`.

Para ver todos los modos:

```bash
go run ./cmd --help
```

## Tests unitarios

Durante el desarrollo, fuimos agregando tests unitarios de cada componente de la implementación, para validar su funcionamiento. Es decir: validar que el scanner escanee como se espera, que el parser se comporte como se espera, etc.

> Para ejecutar **todos** los tests del proyecto:

```bash
go test ./...
```

> Para ejecutar los tests de **un paquete en particular** (por ejemplo, el scanner):

```bash
go test ./internal/scanner
```

> Para ver el detalle de cada test ejecutado (nombre y resultado individual):

```bash
go test -v ./...
```

## Tests de la cátedra

Para ejecutar los tests de la cátedra en el modo **tree walking interpreter**:

```bash
python3 real-tests/script.py "go run ./cmd"
```

Para la segunda implementación: **bytecode compiler**:

```bash
python3 real-tests/script.py "go run ./cmd --vm"
```

En cualquier caso, se mostrará por cada archivo su estado general (`OK` / `ERROR`). En caso de haber errores, aparecerá el error obtenido por Glox.

## To Do List

- [X] Chunk y VM
- [X] Compilador base, expresiones
- [X] statements, variables globales 
- [X] bloques y variables locales
- [X] Control de flujo
- [X] Funciones
- [X] Closures
- [X] Clases
- [ ] Herencia
- [x] Extra propio
- [X] Uso real del lenguaje
- [ ] Benchmarks contra la entrega parcial, mejorar la que teniamos y documentación final

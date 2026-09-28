# Lenguajes y compiladores - TP Lox

Repositorio para el trabajo práctico de la materia Lenguajes y Compiladores.

Implementación de una variante propia del lenguaje [Lox](https://craftinginterpreters.com/), combinando las dos aproximaciones vistas en clase. Desarrollado en [Go](https://go.dev/).

## Tabla de contenidos

- [Integrantes](#integrantes)
- [Documentación de la implementación](#documentación-de-la-implementación)
- [Cómo correr lox](#cómo-correr-lox)
- [Tests unitarios](#tests-unitarios)
- [Tests de la cátedra](#tests-de-la-cátedra)
- [Benchmarks](#benchmarks)
- [To Dos](#to-dos)

## Integrantes

- [Emilia Duzac](https://github.com/emiliaduzac) - 110307
- [Marcos Navarro](https://github.com/NVRar) - 106974

## Documentación de la implementación

Como la implementación será desarrollada en etapas, decidimos separar la documentación en un documento por etapa, para facilitar su lectura y comprensión. Toda la documentación técnica se encuentra bajo la carpeta **docs**.

> Indice de documentación

- [Scanner](docs/scanner.md)
- [Parser](docs/parser.md)
- [Interpreter](docs/interpreter.md)

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

Si se quiere especificar el modo (como `--scanning` o `--parsing`):

```bash
go run ./cmd --<mode>
```

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

Para ejecutar los tests de la cátedra:

```bash
python3 real-tests/script.py "go run ./cmd"
```

Se mostrará por cada archivo su estado general (`OK` / `ERROR`). En caso de haber errores, aparecerá el error obtenido por Glox.

## Benchmarks

TBD

## To Dos

- [ ] Manejar errores: la verdad que me confunde un poco como hace el libro y como hace la versión de clase. No me termina de quedar claro, por ejemplo en el scanner, si tenemos que leer todos los tokens, mostrar el error y seguir leyendo o cortar con el primer error. **Ahora con los tests de la catedra se nota mas, muestra solo el 1er error que se encuentra**
- [X] Parser - expressions
- [X] Interpreter - expressions
- [X] Parser - statements
- [X] Interpreter - statements
- [X] Estado
- [X] Control de flujo
- [ ] Funciones

(esto lo borramos despues para la entrega)
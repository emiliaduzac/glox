# Lenguajes y compiladores - TP Lox
Repositorio para el trabajo práctico de la materia Lenguajes y Compiladores.

Implementación de una variante propia del lenguaje [Lox](https://craftinginterpreters.com/), combinando las dos
aproximaciones vistas en clase. Desarrollado en [Go](https://go.dev/).

## Tabla de contenidos
- [Integrantes](#integrantes)
- [Documentación de la implementación](#documentación-de-la-implementación)

## Integrantes
- [Emilia Duzac](https://github.com/emiliaduzac) - 110307
- [Marcos Navarro](https://github.com/NVRar) - 106974
- [Comandos y ejecución](#comandos-y-ejecución)

## Documentación de la implementación

Como la implementación será desarrollada en etapas, decidimos separar la documentación en un documento por etapa,
para facilitar su lectura y comprensión. Toda la documentación técnica se encuentra bajo la carpeta **docs**.

> Indice de documentación

- [Scanner](docs/scanner.md)

## Comandos y ejecución

### Tests

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
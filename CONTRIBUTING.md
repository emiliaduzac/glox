# Contributing

## Configuración del entorno de desarrollo

1. Clonar el repositorio:

```bash
git clone git@github.com:emiliaduzac/glox.git
cd glox
```

2. Instalar las herramientas de desarrollo:

```bash
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
go install github.com/evilmartians/lefthook/v2@latest
```

3. Instalar los git hooks:

```bash
lefthook install
```

Las dependencias del módulo de Go no necesitan un paso de instalación aparte: se descargan automáticamente la primera vez que corras `go build`, `go run` o `go test`.

## Branching

Usar nombres de rama convencionales:

| Prefijo               | Propósito                                 |
|-----------------------|-------------------------------------------|
| `feature/` o `feat/`  | Nueva funcionalidad                       |
| `fix/`                | Corrección de errores                     |
| `chore/`              | Mantenimiento, dependencias, herramientas |
| `refactor/`           | Reestructuración de código                |
| `docs/`               | Cambios de documentación                  |

## Commits

Seguir [conventional commits](https://www.conventionalcommits.org/).

```
<tipo>(<ámbito>): <descripción>

feat: tokenize string literals
fix(parser): recover from panic on unexpected token
chore: add lefthook and setup
```

Tipos comunes: `feat`, `fix`, `chore`, `docs`, `refactor`.

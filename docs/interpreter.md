## Interpreter

El interpreter tiene la responsabilidad de recorrer el **AST** generado por el parser. La idea es que evalúe las expresiones de forma recursiva. Según el libro se espera que lo haga en un **recorrido post-order**, es decir primero los hijos para luego operar sobre los resultados de estos.

### Mapeo de tipos en Glox

| Tipo en Lox | Representación en Go |
| --- | --- |
| NUMBER | `float64` |
| STRING | `string` |
| BOOL | `bool` |
| NIL | `nil` |

Particularmente en la parte de nil, nos encontramos que por cómo trabaja Go terminaba saliendo `<nil>`, entonces ahí fue que implementamos la función de `Stringify`.

### Tipos

Para saber el tipo usamos un `switch e := expr.(type)`, a diferencia de, por ejemplo el libro, donde proponen su implementación en Java utilizando el patrón Visitor. 

Acá es que también debemos manejar "a mano" el caso `default` y estar atentos para un error explícito de tipo no soportado.

### Operadores

1. **Truthiness**: Es importante respetar Truthiness para Lox. En este caso se define que solo `false` y `nil` son falsos, y que todo lo demás es verdadero, comportamiento que quedó declarado en la función de `isTruthy`.

2. En glox manejamos la no coerción implícita de tipo. Es decir, `3 == "3"` es `false`, termina comparando tipos distintos de datos y es manejado simplemente con un `false` y no con un error. 

3. **División por cero**: Al dividir por cero, se retorna `+Inf` o `-Inf` según el caso. Y en el caso de `0 / 0`, se retorna `NaN`.

4. Sobre el operador módulo (`%`): Al manejarlo Go solo lo aplica a enteros, por lo que usamos `math.Mod` para preservar los flotantes y no caer en panics. 

### Errores en tiempo de ejecución

Si por ejemplo llega un `true < false`, no se crashea con `panic`, sino que se retorna un error con el número de línea informativo. 

### Tests

Quisimos dejar plasmado comportamiento en test de las diferentes expresiones, y los casos de cada una de ellas.

```bash
go test ./internal/interpreter
```

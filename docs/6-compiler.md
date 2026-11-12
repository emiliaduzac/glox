## Compiler

Siguiendo la clase y el libro, nuestro compilador tendrá dos responsabilidades:

- **Parsear** los tokens para darles estructura y entender qué significan.
- **Emitir** las instrucciones que generan, en el orden en que tienen que ejecutarse.

Es decir que reemplazará a nuestro **parser** y a nuestro **resolver**. Esto se define como un **Single Pass Compiler**: de los tokens pasamos directamente a bytecode, emitiendo a medida que leemos (la entrada son los _tokens_ y la salida son los _chunks_).
Otras implementaciones de compiladores aprovechan el AST parseado, pero el libro elige un single pass compiler por su simplicidad. Además, indica que originalmente muchos compiladores se hacían de esta forma por razones de memoria y que este enfoque podría no servir para cualquier lenguaje, pero Lox lo permite por ser de tipado dinámico: el compilador no necesita conocer los tipos de lo que todavía no leyó, porque se resuelven en tiempo de ejecución.

### Comparación con la primera entrega

**Tree-walk**: scanner → parser → resolver → interpreter

**Bytecode**: scanner → compiler → VM

Para mantener ambos enfoques y evitar duplicar código, se agregó un `TokenCursor` ([`cursor.go`](../internal/token/cursor.go)), que incluye el listado de tokens y el índice del token actual. De esta forma, centralizamos la lógica de recorrer los tokens del `parser` y del `compiler`.

### Pratt Parser

Para la etapa de parseo de expresiones, se utiliza el algoritmo _Pratt Parser_. A diferencia del _Recursive Descent Parser_ de nuestro [parser](2-parser.md), donde cada nivel de precedencia tiene su propia regla y su propia función, el Pratt Parser usa una **tabla** y una única función: a medida que recorre los tokens, obtiene la información de cada uno desde la tabla y ejecuta lo que corresponda.

La tabla ([`precedence.go`](../internal/compiler/precedence.go)) indica para cada tipo de token:

- **prefix**: dice que hacer si el token empieza una expresión
- **infix**: que hacer si el token aparece después de un operando. Por ejemplo, el `+` en `1 + 2`.
- **precedence**: qué tan fuerte es el token como operador infijo (es lo que nos dice precedencia. Por ejemplo, `*` tiene un nivel mayor de precedencia que `+`)

#### Expresiones

Para compilar expresiones, la función `parsePrecedence(precedence)` recibe una precedencia, que indica la mínima que puede aceptar. Funciona así:

1. Obtiene el token actual y, si tiene un _prefix_ asociado, lo ejecuta. Si no lo tiene, es un error: ese token no debería estar empezando una expresión.
2. Itera los tokens siguientes, obteniendo sus _infix_ y ejecutándolos.
3. Si encuentra un token de menor precedencia a la recibida, deja de iterar, porque quiere decir que ese token no pertenece a la expresión actual.

Un truco que aplica el libro es no generar instrucciones nuevas para `!=`, `>=` y `<=`, sino aprovechar las existentes y negarlas. Por ejemplo, `a != b` sería `!(a == b)`.

#### Literales

Los números y los strings se compilan de la misma forma, se guarda el valor en el pool de constantes del chunk y se crea un OP_CONSTANT con su índice. Es por esto mismo que en la tabla los números y strings apuntan a la misma funcion value que agarra el valor que el scanner dejó en el token y lo deja como constante.
Un detalle de diferencia con el libro acá es que para los strings el compilador tiene que sacar las comillas pero para nuestra implementación eso no es necesario porque ya lo hace a eso el scanner. Y otra cosa interesante es que en el libro necesita copiar los caracteres al heap para que el mismo string sea dueño de su memoria, pero nosotros en glox no hacemos eso porque en go los strings son inmutables asi que se pueden compartir sin copiarlos. 

#### Asociatividad

A diferencia de la precedencia, la asociatividad no está en la tabla, sino que sale de cómo cada función compila sus operandos.
En el caso de los binarios, `binary()` compila su operando **derecho** llamando a `parsePrecedence` con la precedencia del operador más uno. Así, el operando derecho no puede quedarse con otro operador del mismo nivel, y este queda para la expresión de afuera.

Paso a paso para el ejemplo `1 - 2 - 3`:

1. `parsePrecedence(PREC_ASSIGNMENT)` acepta cualquier operador. Compila el `1`.
2. Encuentra el primer `-` (`PREC_TERM`). Como es mayor o igual a `PREC_ASSIGNMENT`, lo toma y llama a `binary()`.
3. `binary()` compila el lado derecho con `parsePrecedence(PREC_FACTOR)` (suma 1 a la precedencia actual, que era `PREC_TERM`), que solo acepta operadores de `PREC_FACTOR` o mayor. Compila el `2`.
4. Encuentra el segundo `-` (`PREC_TERM`). Como es menor a `PREC_FACTOR`, corta. Se emite la resta: `1 - 2`.
5. De vuelta en el primer nivel, encuentra el segundo `-`. Como es mayor o igual a `PREC_ASSIGNMENT`, lo toma y llama a `binary()`.
6. `binary()` compila el `3` y se emite la segunda resta: `(1 - 2) - 3`.

### Tests

Los tests unitarios del compilador se pueden ejecutar de la siguiente forma:

```bash
go test ./internal/compiler
```

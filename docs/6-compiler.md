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

Los números y los strings se compilan de la misma forma, se guarda el valor en el pool de constantes del chunk y se crea un `OP_CONSTANT` con su índice. Es por esto mismo que en la tabla los números y strings apuntan a la misma funcion `value()` que agarra el valor que el scanner dejó en el token y lo deja como constante.
Un detalle de diferencia con el libro es que para los strings el compilador tiene que sacar las comillas, pero para nuestra implementación eso no es necesario porque eso ya lo hace el scanner. Y otra cosa interesante es que en el libro necesita copiar los caracteres al heap para que el mismo string sea dueño de su memoria, pero nosotros en glox no hacemos eso porque en Go los strings son inmutables asi que se pueden compartir sin copiarlos.

#### Asociatividad

Como vimos, cada operador tiene en la tabla una función que se ejecuta cuando el compilador lo encuentra. Las que nos interesan para la asociatividad son las de los operadores: `binary()` para los binarios, `unary()` para los unarios y `variable()`, que maneja la asignación cuando después de un nombre encuentra un `=`.

Cuando se llama a una de esas funciones, lo que está a la izquierda del operador ya se compiló (o no existe, en el caso de los unarios). Queda compilar lo que viene después del operador y emitir la instrucción del operador. Para eso, cada función llama a `parsePrecedence` con una precedencia mínima, que definirá la asociatividad. Si aparece otro operador del mismo nivel (con la misma precedencia), puede pasar una de dos cosas:

- Si la función pide una precedencia **mayor** que la del operador actual, un operador que tenga el mismo nivel no alcanzaría el mínimo, así que no entra a la expresión: queda para la de afuera. La asociatividad es **a izquierda**.
- Si pide la **misma o una menor**, el siguiente operador sí alcanza el mínimo, así que entra y se resuelve primero. La asociatividad es **a derecha**.

| Función | Precedencia que pide | Asociatividad | Ejemplo |
| --- | --- | --- | --- |
| `binary()` | La del operador + 1 | A izquierda | `1 - 2 - 3` → `(1 - 2) - 3` |
| `unary()` | `PREC_UNARY`, la de su operador | A derecha | `- -1` → `-(-1)` |
| `variable()` | `PREC_ASSIGNMENT`, la más baja | A derecha | `a = b = 1` → `a = (b = 1)` |

### Tests

Los tests unitarios del compilador se pueden ejecutar de la siguiente forma:

```bash
go test ./internal/compiler
```

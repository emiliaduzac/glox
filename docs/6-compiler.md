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

### Expresiones

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

### Variables locales y scopes

En nuestra implementación anterior, en el manejo de scopes tenían un rol importante el resolver y el intérprete. El resolver ayudaba a dejar definido antes del runtime a qué distancia estaba cada variable definida, para que luego el intérprete usara eso en vez de calcularlo en ejecución. Con esa distancia, el intérprete subía hasta el entorno correspondiente y ahí buscaba la variable por nombre.

En esta nueva versión, el compilador es quien se encarga de calcular antes de la ejecución, en qué índice del stack de la VM va a estar cada variable local. Para eso, el compilador guarda un `scopeDepth`, que indica para cada momento la profundidad actual. Entonces, cuando se encuentra con un `block statement`, inicia un nuevo **scope**, que es simplemente sumar uno al contador de profundidad. Luego, sigue compilando todo lo que está dentro del bloque. Si lo que hay dentro del bloque son variables locales, las va a guardar en su **"stack falso"** `locals`. Cada elemento de este stack es una `Local`, compuesta por su nombre y su profundidad. La idea es que el índice de cada variable en este nuevo stack coincide con el índice que tendrá en el stack de la VM. Esto funciona porque cada statement deja el stack como lo encontró: los valores temporales de una expresión se apilan arriba de las locales y se sacan antes del siguiente statement.

Cuando encuentra una declaración de una variable local, recorre las locales del scope actual y valida que no haya otra con ese nombre, ya que redeclarar en el mismo scope es un error.
Si no hay, la declara con profundidad `-1`, indicando que aún no fue inicializada (para evitar los `var a = a;`). Luego de ser inicializada la marca con su profundidad correspondiente.

De esta forma, no necesita generar una instrucción como `OP_DEFINE_GLOBAL`, ya que el valor de la variable queda directamente en el stack. Luego, cuando quiera usarla (con `OP_GET_LOCAL` o `OP_SET_LOCAL`), usa como operando su índice en el stack. No necesitamos guardar el nombre en el bytecode, porque la VM accede a las locales por su índice, sin importar su nombre. Además, para saber qué índice usar, el compilador busca el nombre en las locales de atrás para adelante, así siempre encuentra la más interna. Esto permite que una variable de un scope interno tape a otra de afuera con el mismo nombre (*shadowing*).

Al finalizar el bloque, resta uno a la profundidad y elimina de su stack falso las variables locales de ese bloque. Además, emite un `OP_POP` por cada una, para asegurarnos de que al terminar un bloque, el stack de la VM quede sin sus variables.

Un detalle más sobre las variables locales es que puede haber como máximo 256 a la vez, ya que el operando de `OP_GET_LOCAL` y `OP_SET_LOCAL` ocupa un solo byte, que va de 0 a 255.

### Control de flujo

#### If - Else

En el tree-walk interpreter, un `if` de Lox se ejecutaba directamente con un if de Go. En el bytecode compiler, en cambio, lo implementamos como lo haría una CPU: con saltos.

Hasta ahora todas las instrucciones venían siendo secuenciales, pero para los `if-else`, vamos a querer que saltee instrucciones según la rama en la que cae. Es decir: si la condición del `if` es verdadera, seguimos secuencial, pero salteamos toda la parte del `else` (si la hay). Lo mismo si la condición no se cumple: queremos saltear el cuerpo del `if`.

Para eso, tenemos dos nuevas instrucciones:

- `OP_JUMP`: salta la cantidad de bytes que indica su operando. Lo usamos luego del cuerpo del `if`, para que se saltee toda la parte del `else` y su cuerpo.
- `OP_JUMP_IF_FALSE`: mira el tope del stack, donde quedó la condición, y solo si es falsa salta la cantidad de bytes que indica su operando. Lo usamos si la condición del `if` no se cumple y no queremos ejecutar su cuerpo.

En ambos casos, cuando el compilador emite el salto, todavía no sabe cuántos bytes tiene que saltear, ya que, como mencionamos, vamos leyendo secuencialmente. Al emitir el `OP_JUMP_IF_FALSE` todavía no compiló el cuerpo del `if`, así que no sabe cuánto ocupa. Lo mismo pasa con el `OP_JUMP` y el cuerpo del `else`. Para resolverlo se usa el truco de **backpatching**: ponemos un placeholder de dos bytes (para abarcar cuerpos posiblemente grandes) y una vez compilado el cuerpo, volvemos y escribimos el valor real.

Por último, se debe dejar el stack limpio, como cualquier statement. Cuando hay un `if`, la condición puede ser verdadera o falsa, pero en ambos casos queda en el tope del stack. Además, el compilador no sabe como se evalúa la condición, entonces tenemos que cubrir ambos casos. El orden quedaría:

```
<condición>
OP_JUMP_IF_FALSE <hasta el POP del else, antes de ejecutar el cuerpo del else>
OP_POP // solo se va a ejecutar si la condición dio verdadera
<cuerpo del if>
OP_JUMP <hasta después del else>
OP_POP // a partir de acá, solo si la condición dio falsa, aunque no haya cuerpo de else
<cuerpo del else (opcional)>
<lo que sigue>
```

#### Operadores lógicos

Para los operadores lógicos tampoco vamos a crear nuevas instrucciones: vamos a aprovechar los saltos.

- **and**: si el primer operando es falso, no hace falta evaluar el segundo: el resultado es el primero. Entonces compilamos el primer operando y emitimos un `OP_JUMP_IF_FALSE`. Justo después emitimos un `OP_POP` y compilamos el segundo operando: si el primero era verdadero, la VM lo descarta y el resultado pasa a ser el segundo. Una vez compilado el segundo operando, ya podemos completar cuántos bytes tiene que saltear el `OP_JUMP_IF_FALSE`.
- **or**: este caso es al revés: si el primer operando es verdadero, no hace falta evaluar el segundo, y si es falso, sí. Como no existe una instrucción para saltar si es verdadero, usamos dos saltos: después del primer operando emitimos un `OP_JUMP_IF_FALSE` y seguido un `OP_JUMP`. Si el primero dio falso, el `OP_JUMP_IF_FALSE` saltea el `OP_JUMP`, se descarta el primer operando con un `OP_POP` y se evalúa el segundo. Si era verdadero, no salta nada, pero se encuentra con el `OP_JUMP`, que saltea el segundo operando. Del `OP_JUMP_IF_FALSE` ya sabemos cuánto tiene que saltear: solo el `OP_JUMP` que le sigue. El `OP_JUMP` lo completamos una vez compilado el segundo operando.

Algo interesante, igual que en la implementación anterior, es que el resultado no es necesariamente un booleano, sino uno de los dos operandos tal cual. Cuál de los dos depende de si el primero es *truthy* (en Lox, todo menos `false` y `nil`). Por ejemplo, `"hola" or "chau"` da `"hola"`, porque `"hola"` es *truthy*.

#### Iteraciones

Para los ciclos necesitamos saltar hacia **atrás**, para volver a ejecutar una parte del código. Para eso agregamos una nueva instrucción: `OP_LOOP`. Funciona igual que `OP_JUMP` pero resta en vez de sumar. A diferencia de los otros saltos, no necesita backpatching: cuando se emite, ya sabemos a dónde tiene que volver: el inicio del ciclo, que ya fue compilado.

- **while**: apenas nos encontramos con un `while`, guardamos la posición actual del chunk, que es donde va a empezar la condición. Luego compilamos la condición y emitimos un `OP_JUMP_IF_FALSE`, que actúa igual que en el `if`: si la condición es falsa, saltea todo el cuerpo. Le sigue un `OP_POP` para eliminar la condición del stack, que solo se ejecuta si la condición fue verdadera (sino, se hubiese salteado). Luego se compila el cuerpo y el `OP_LOOP` que vuelve a la posición que habíamos guardado, para evaluar la condición otra vez. Recién ahí terminan las instrucciones del ciclo y podemos completar el `OP_JUMP_IF_FALSE`, que salta hasta un último `OP_POP`, que se ejecuta una vez que la condición es falsa.
- **for**: este tipo de ciclo tiene 3 partes y todas son opcionales. Todo el `for` está dentro de un scope propio, así que una variable declarada en el inicializador es local al ciclo.
  En primer lugar, el _inicializador_, que puede ser la declaración de una variable o una expresión, y **se debe ejecutar una única vez**. Recién después de eso marcamos el inicio del ciclo, guardando la posición actual del chunk.
  A eso le sigue la _condición_, que, igual que en el `while`, emite un `OP_JUMP_IF_FALSE` para salir del ciclo cuando no se cumple. Si no hay condición, no hay salto y el ciclo sería infinito.
  Finalmente, el _incremento_, que a pesar de estar escrito antes del cuerpo, debe ejecutarse **después** de cada iteración. Como el compilador lo emite en el orden en que lo lee, el incremento queda en el bytecode entre la condición y el cuerpo, y lo resolvemos con saltos para que la VM los recorra en este orden:
  - Después de la condición, un `OP_JUMP` saltea el incremento y va directo al cuerpo.
  - Al final del cuerpo, un `OP_LOOP` vuelve al incremento.
  - Al final del incremento, otro `OP_LOOP` vuelve a la condición.
  Si no hay incremento, el cuerpo vuelve directamente a la condición, como en el `while`. Por último, se completa el salto de salida de la condición.

### Tests

Los tests unitarios del compilador se pueden ejecutar de la siguiente forma:

```bash
go test ./internal/compiler
```

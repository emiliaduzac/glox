# Compiler

Siguiendo la clase y el libro, nuestro compilador tendrá dos responsabilidades:

- **Parsear** los tokens para darles estructura y entender qué significan.
- **Emitir** las instrucciones que generan, en el orden en que tienen que ejecutarse.

Es decir que reemplazará a nuestro **parser** y a nuestro **resolver**. Esto se define como un **Single Pass Compiler**: de los tokens pasamos directamente a bytecode, emitiendo a medida que leemos (la entrada son los _tokens_ y la salida son los _chunks_).
Otras implementaciones de compiladores aprovechan el AST parseado, pero el libro elige un single pass compiler por su simplicidad. Además, indica que originalmente muchos compiladores se hacían de esta forma por razones de memoria y que este enfoque podría no servir para cualquier lenguaje, pero Lox lo permite por ser de tipado dinámico: el compilador no necesita conocer los tipos de lo que todavía no leyó, porque se resuelven en tiempo de ejecución.

## Comparación con la primera entrega

**Tree-walk**: scanner → parser → resolver → interpreter

**Bytecode**: scanner → compiler → VM

Para mantener ambos enfoques y evitar duplicar código, se agregó un `TokenCursor` ([`cursor.go`](../internal/token/cursor.go)), que incluye el listado de tokens y el índice del token actual. De esta forma, centralizamos la lógica de recorrer los tokens del `parser` y del `compiler`.

Además nos guardamos el anterior, que el compilador usa para saber en qué línea está cada instrucción que emite. Al avanzar se guarda lo que se consumió como anterior, así que siempre existe uno y eso cubre un caso más borde que nos importa, el caso de que llega solo un `EOF`, que sería un programa vacío, ahí no se rompe, como sí pasaría si calculáramos el anterior como `actual - 1`, porque ahí no habría nada en el anterior.

## Pratt Parser

Para la etapa de parseo de expresiones, se utiliza el algoritmo _Pratt Parser_. A diferencia del _Recursive Descent Parser_ de nuestro [parser anterior](2-parser.md), donde cada nivel de precedencia tiene su propia regla y su propia función, el Pratt Parser usa una **tabla** y una única función: a medida que recorre los tokens, obtiene la información de cada uno desde la tabla y ejecuta lo que corresponda.

La tabla ([`precedence.go`](../internal/compiler/precedence.go)) indica para cada tipo de token:

- **prefix**: que hacer si el token empieza una expresión. Por ejemplo, el `-` en `-1`.
- **infix**: que hacer si el token aparece después de un operando. Por ejemplo, el `+` en `1 + 2`.
- **precedence**: qué tan fuerte es el token como operador infijo. Por ejemplo, `*` tiene un nivel mayor de precedencia que `+`.

Una diferencia con el libro es cómo armamos la tabla. En clox es un array con una entrada para todos los tokens, incluso los que no participan en expresiones. Nosotros usamos un `map`, y como un token que no está devuelve el valor cero (sin funciones y con `PREC_NONE`), solo listamos los que importan: los que no están obtienen el comportamiento esperado igual.

Además, para las funciones hicimos un tipo `parseFn func(c *Compiler, canAssign bool) error`, y la tabla guarda métodos del compilador, como `(*Compiler).binary`, que al llamarlos reciben el compilador como primer parámetro y devuelven un `error`.

Por último, la tabla y las funciones de parseo dependen una de la otra: la tabla referencia a las funciones, y las funciones usan la tabla a través de `getRule()`. En C, el libro lo resuelve declarando `getRule()` antes de las funciones y definiéndola después de la tabla. En Go, inicializar la tabla directamente da un error de _initialization cycle_, así que la armamos dentro de `init()`.

## Expresiones

Para compilar expresiones, la función `parsePrecedence(precedence)` recibe una precedencia, que indica la mínima que puede aceptar. Funciona así:

1. Obtiene el token actual y, si tiene un _prefix_ asociado, lo ejecuta. Si no lo tiene, es un error: ese token no debería estar empezando una expresión.
2. Itera los tokens siguientes, obteniendo sus _infix_ y ejecutándolos.
3. Si encuentra un token de menor precedencia a la recibida, deja de iterar, porque quiere decir que ese token no pertenece a la expresión actual.

Un truco que aplica el libro es no generar instrucciones nuevas para `!=`, `>=` y `<=`, sino aprovechar las existentes y negarlas. Por ejemplo, `a != b` sería `!(a == b)`.

### Literales

Los números y los strings se compilan de la misma forma, se guarda el valor entre las constantes del chunk y se crea un `OP_CONSTANT` con su índice. Es por esto mismo que en la tabla los números y strings apuntan a la misma función `value()`, que agarra el valor que el scanner dejó en el token y lo deja como constante.
Una diferencia con el libro es que para los strings, el compilador tiene que sacar las comillas, pero para nuestra implementación eso no es necesario porque eso ya lo hace el scanner. Y otra cosa interesante es que en el libro necesita copiar los caracteres al heap para que el mismo string sea dueño de su memoria, pero nosotros en glox no hacemos eso porque en Go los strings son inmutables así que se pueden compartir sin copiarlos.

Tenemos el límite que el mismo libro habla, sobre que la posición de cada constante ocupa un solo byte así que un chunk puede tener como máximo 256 constantes. Quizás acá es un punto a mejorar, pero por ahora lo que tenemos es que para gastar menos lugar los nombres de las variables no se repiten y eso lo explicamos más a detalle en la sección de [nombres repetidos](#nombres-repetidos).

### Asociatividad

Como vimos, cada operador tiene en la tabla una función que se ejecuta cuando el compilador lo encuentra. Las que nos interesan para la asociatividad son las de los operadores: `binary()` para los binarios, `unary()` para los unarios y `variable()`, que maneja la asignación cuando después de un nombre encuentra un `=`.

Cuando se llama a una de esas funciones, lo que está a la izquierda del operador ya se compiló (o no existe, en el caso de los unarios). Queda compilar lo que viene después del operador y emitir la instrucción del operador. Para eso, cada función llama a `parsePrecedence` con una precedencia mínima, que definirá la asociatividad. Si aparece otro operador del mismo nivel (con la misma precedencia), puede pasar una de dos cosas:

- Si la función pide una precedencia **mayor** que la del operador actual, un operador que tenga el mismo nivel no alcanzaría el mínimo, así que no entra a la expresión: queda para la de afuera. La asociatividad es **a izquierda**.
- Si pide la **misma o una menor**, el siguiente operador sí alcanza el mínimo, así que entra y se resuelve primero. La asociatividad es **a derecha**.

## Statements

Con los statements, un programa pasa a ser una lista de declaraciones que se compilan hasta llegar a un `EOF`. Acá aparecen dos funciones: `declaration`, que si lo que viene empieza con `var` compila la declaración de la variable, y si no se lo pasa a `statement`. Esta última compila el resto de las sentencias: `print`, bloques, `if`, ciclos y las expresiones que terminan en `;`. Las separamos así porque, como explica el libro, en Lox una declaración no puede ser el cuerpo de un `if` o un `while`, ya que no quedaría claro hasta dónde llega la variable.

Es importante ver el efecto que tiene cada statement sobre el stack de la VM (su _stack effect_). La idea de fondo es que una expresión siempre deja un valor más en el stack, y un statement tiene que dejarlo como estaba. Por ejemplo, en `print 1 + 2;` se apila el 1, después el 2, y `OP_ADD` los saca y apila el 3: la expresión dejó un valor más. Después, `OP_PRINT` saca ese resultado y lo imprime, así que el statement termina dejando el stack como estaba. Si en cambio solo se escribe `1 + 2;`, el compilador emite un `OP_POP` para tirar el valor que dejó la expresión. En el tree-walk, ese descarte lo hacía Go, ignorando el valor devuelto; ahora es una instrucción más.

Esto es importante porque, si cada statement dejara algo en el stack, con muchos statements seguidos el stack crecería constantemente. Además, es lo que permite que los índices de las variables locales coincidan con su posición en el stack (ver [variables locales y scopes](#variables-locales-y-scopes)). La única excepción es la declaración de una variable local, que deja su valor en el stack a propósito, porque ese valor es la variable, y se saca recién al cerrar su bloque.

Para asegurarnos de que esto se cumpla, cuando termina el script la VM verifica que el stack haya quedado vacío y si no levanta un error, porque significa que el compilador dejó algo de más.

## Manejo de errores

Acá llegó un gran cambio frente al tree-walk de la primera parte: al detectar un error, cortaba y mostraba solo ese. Si había varios, el programador corregía el primero, volvía a correr y recién ahí veía el siguiente, que ya estaba desde antes. En esta segunda parte la idea es seguir compilando después de un error, para mostrarlos todos juntos y poder corregirlos de una vez.

Seguir después de un error trae un detalle: el compilador queda parado en una línea rota, y si sigue leyendo desde ahí, levantaría errores falsos. Para eso está la **sincronización**: avanza hasta que termina el statement (un `;`) o empieza uno nuevo (por ejemplo, un `var` o un `print`), y recién desde ahí sigue compilando.

Al principio pensamos en aprovechar algo típico de Go: cortar el statement con el error y devolverlo hacia arriba. El problema apareció con las variables locales: al sincronizar, lo hacíamos dos veces, una dentro del scope y otra al salir, y eso generaba un error falso de que faltaba un `}`. Ahí decidimos pegarnos más al libro con la idea del _flag_ de modo pánico, pero en vez de mostrar el error en el momento, lo guardamos. El flag se prende cuando hay un error, mientras está prendido los errores nuevos se ignoran, y se apaga cuando se sincroniza y arranca un statement nuevo. Así cada error aparece una sola vez y donde corresponde.

Otra diferencia con el libro es que, en C, reportar un error no corta la función: sigue leyendo tokens. En Go, en cambio, devolver un error corta la función en ese punto. Por eso tenemos dos caminos:
  - cuando el código está mal escrito y no se puede seguir leyendo (por ejemplo, falta un `)`), devolvemos el error y se sincroniza.
  - cuando el código está bien escrito pero rompe una regla del lenguaje (como redeclarar una variable local en el mismo scope), lo reportamos con `report` y seguimos compilando normalmente.

Al terminar de compilar, si hubieron errores, se devuelven todos juntos en un solo error. Además, en ese caso no se devuelve el chunk: el bytecode generado puede estar incompleto (por ejemplo, con saltos sin completar), así que la VM no llega a ejecutar nada.

## Variables globales

En la primera parte una variable global era una entrada más del `Environment` y el intérprete la buscaba por nombre cuando la necesitaba. Acá el compilador no se guarda el valor de la variable, lo que hace es emitir las instrucciones para que la VM se encargue cuando se ejecuta. Las globales en Lox se resuelven tarde (_late bound_), es decir que se buscan por nombre recién en la ejecución. El tema es que un nombre es un string y no entra como operando de una instrucción, que es de un solo byte. Para resolver esto, el nombre se guarda entre las constantes del chunk y la instrucción lleva la posición donde quedó. Y para saber si una variable es global pasa lo mismo que con el resolver de la primera parte, si no se encuentra en ningún scope local se considera global. Al agregar closures, primero se chequea que no sea un **upvalue**. 

**Declaración**

Cuando el compilador se encuentra con `var a = 1;` primero guarda el nombre `a` entre las constantes del chunk, después compila el valor, y recién al final emite `OP_DEFINE_GLOBAL` con la posición del nombre. Va al final porque cuando la VM ejecute esa instrucción necesita que el valor ya esté en el stack para poder guardarlo. Si no tiene valor, como en `var a;`, se emite un `OP_NIL` en su lugar, así que toda variable declarada existe y vale `nil` hasta que se le asigne algo, igual que en el tree-walk. Otro detalle que también es igual es que declarar de nuevo una global no da error, solo pisa el valor que tenía.

**Lectura**

Cuando aparece un identificador el compilador emite `OP_GET_GLOBAL` con la posición del nombre y nada más. Importante es ver que acá no se fija si la variable existe, eso se sabe recién cuando la VM ejecuta la instrucción y busca el nombre en su tabla de globales, y si no está levanta un error de runtime. Es decir que `print a;` compila aunque `a` no esté declarada, y eso es justamente lo que significa que sea _late bound_. El libro lo deja así porque más adelante una función puede usar una global que se declara más abajo en el código, y lo único que importa es que exista cuando esa función se ejecute.

**Asignación**

Si después del nombre viene un `=`, en vez de leer la variable el compilador compila el valor nuevo y emite `OP_SET_GLOBAL`. Como la asignación es una expresión su valor queda disponible, y por eso se puede encadenar `a = b = 5`, que se resuelve de derecha a izquierda como en el parser de la primera parte.

El tema acá es que como compilamos en una sola pasada, cuando se lee `a` todavía no se sabe si es una lectura o el lugar donde se asigna algo. En el parser de la primera parte podíamos armar el lado izquierdo y después fijarnos si venía un `=`, pero acá ya se emitió el código de lo que está a la izquierda, entonces el libro lo resuelve mirando si viene un `=` justo después del nombre. Pero eso solo no alcanza, en `a * b = c;` la `b` está adentro de la multiplicación, vería el `=` y se lo quedaría, y terminaría armando `a * (b = c)` cuando en realidad es `(a * b) = c`, que no es válido. Para evitarlo, `parsePrecedence()` calcula un `canAssign`, que es verdadero solo si estamos en el nivel más bajo de la expresión, y se lo pasa a cada función de la tabla. Entonces la `b` ahí sabe que no puede agarrar el `=`, y cuando termina la expresión y queda un `=` sin agarrar, el compilador levanta `invalid assignment target`, que es el mismo error del parser de la primera parte.

### Nombres repetidos

En el libro cada vez que aparece una variable se guarda su nombre de nuevo entre las constantes, y deja como ejercicio evitarlo. Sabemos que con el límite de 256 constantes esto es un problema si en un programa se usa muchas veces la misma variable, porque cada uso gasta un lugar, por lo tanto lo quisimos resolver. En el compilador decidimos crear un `map` de nombre a posición para saber si una variable ya apareció, porque nos pareció la forma más simple y que no toca nada de la VM. Si ya apareció, reusamos esa posición en vez de guardarla otra vez, así que a la VM no le cambia nada porque sigue recibiendo una posición que apunta al nombre correcto. El costo que nos queda es una consulta al `map` cada vez que se usa una variable al compilar. Los números y strings literales no los tocamos, esos se siguen guardando cada vez que aparecen.

## Variables locales y scopes

En nuestra implementación anterior, en el manejo de scopes tenían un rol importante el resolver y el intérprete. El resolver ayudaba a dejar definido antes del runtime a qué distancia estaba cada variable definida, para que luego el intérprete usara eso en vez de calcularlo en ejecución. Con esa distancia, el intérprete subía hasta el entorno correspondiente y ahí buscaba la variable por nombre.

En esta nueva versión, el compilador es quien se encarga de calcular antes de la ejecución, en qué índice del stack de la VM va a estar cada variable local. Para eso, el compilador guarda un `scopeDepth`, que indica para cada momento la profundidad actual. Entonces, cuando se encuentra con un `block statement`, inicia un nuevo **scope**, que es simplemente sumar uno al contador de profundidad. Luego, sigue compilando todo lo que está dentro del bloque. Si lo que hay dentro del bloque son variables locales, las va a guardar en su **"stack falso"** `locals`. Cada elemento de este stack es una `Local`, compuesta por su nombre y su profundidad. La idea es que el índice de cada variable en este nuevo stack coincide con el índice que tendrá en el stack de la VM. Esto funciona porque cada statement deja el stack como lo encontró: los valores temporales de una expresión se apilan arriba de las locales y se sacan antes del siguiente statement.

Cuando encuentra una declaración de una variable local, recorre las locales del scope actual y valida que no haya otra con ese nombre, ya que redeclarar en el mismo scope es un error.
Si no hay, la declara con profundidad `-1`, indicando que aún no fue inicializada (para evitar los `var a = a;`). Luego de ser inicializada la marca con su profundidad correspondiente.

De esta forma, no necesita generar una instrucción como `OP_DEFINE_GLOBAL`, ya que el valor de la variable queda directamente en el stack. Luego, cuando quiera usarla (con `OP_GET_LOCAL` o `OP_SET_LOCAL`), usa como operando su índice en el stack. No necesitamos guardar el nombre en el bytecode, porque la VM accede a las locales por su índice, sin importar su nombre. Además, para saber qué índice usar, el compilador busca el nombre en las locales de atrás para adelante, así siempre encuentra la más interna. Esto permite que una variable de un scope interno tape a otra de afuera con el mismo nombre (*shadowing*).

Al finalizar el bloque, resta uno a la profundidad y elimina de su stack falso las variables locales de ese bloque. Además, emite un `OP_POP` por cada una, para asegurarnos de que al terminar un bloque, el stack de la VM quede sin sus variables.

Un detalle más sobre las variables locales es que puede haber como máximo 255 a la vez, ya que el operando de `OP_GET_LOCAL` y `OP_SET_LOCAL` ocupa un solo byte, que va de 0 a 255, y el slot 0 queda reservado para la función.

## Control de flujo

### If - Else

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

### Operadores lógicos

Para los operadores lógicos tampoco vamos a crear nuevas instrucciones: vamos a aprovechar los saltos.

- **and**: si el primer operando es falso, no hace falta evaluar el segundo: el resultado es el primero. Entonces compilamos el primer operando y emitimos un `OP_JUMP_IF_FALSE`. Justo después emitimos un `OP_POP` y compilamos el segundo operando: si el primero era verdadero, la VM lo descarta y el resultado pasa a ser el segundo. Una vez compilado el segundo operando, ya podemos completar cuántos bytes tiene que saltear el `OP_JUMP_IF_FALSE`.
- **or**: este caso es al revés: si el primer operando es verdadero, no hace falta evaluar el segundo, y si es falso, sí. Como no existe una instrucción para saltar si es verdadero, usamos dos saltos: después del primer operando emitimos un `OP_JUMP_IF_FALSE` y seguido un `OP_JUMP`. Si el primero dio falso, el `OP_JUMP_IF_FALSE` saltea el `OP_JUMP`, se descarta el primer operando con un `OP_POP` y se evalúa el segundo. Si era verdadero, no salta nada, pero se encuentra con el `OP_JUMP`, que saltea el segundo operando. Del `OP_JUMP_IF_FALSE` ya sabemos cuánto tiene que saltear: solo el `OP_JUMP` que le sigue. El `OP_JUMP` lo completamos una vez compilado el segundo operando.

Algo interesante, igual que en la implementación anterior, es que el resultado no es necesariamente un booleano, sino uno de los dos operandos tal cual. Cuál de los dos depende de si el primero es *truthy* (en Lox, todo menos `false` y `nil`). Por ejemplo, `"hola" or "chau"` da `"hola"`, porque `"hola"` es *truthy*.

### Iteraciones

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

## Funciones

### El script como función

Un cambio importante al llegar a funciones es que el compilador ya no escribe en un chunk suelto, sino en **el chunk de una función**. Por eso `Compile` pasa a devolver una función. El programa completo se trata como una función implícita: el script, así que el compilador siempre está compilando dentro de una función y la VM siempre ejecuta una. Esto no cambia a las globales: aunque el script sea una función, las variables declaradas en el nivel de arriba siguen siendo globales y no locales del script. Además, el compilador guarda si lo que está compilando es el script o una función, lo que permite, por ejemplo, dar error con un `return` fuera de una función.

Algo importante que se agrega con todo esto es que el índice 0 de la ventana de cada función queda reservado para que la VM ponga la función que se está ejecutando (desde closures, su closure). Como el operando de una local es de un byte, hay 256 lugares, y al ocupar el 0 quedan 255 para variables locales. Esto pasa a ser una **limitación conocida** de esta versión, porque el tree-walk no tiene este límite.

### Una funcion con su propio chunk

Con funciones, cada una tiene su propio código, así que cada una tiene su propio chunk, y la VM tiene que poder pasar de uno a otro cuando se llama a una función. El libro marca que en Lox una función es un valor, como un número o un string, así que se puede guardar en una variable. Además de su **chunk**, guarda su **nombre** para identificarla y su **aridad**, cuántos parámetros recibe, que se usa para chequear que una llamada reciba la cantidad de argumentos esperada.

Una diferencia importante con el libro es que, como clox está hecho en C, la función es un objeto del heap y hay que manejar su memoria a mano. En glox es un struct común, y el garbage collector de Go la libera cuando ya no tiene referencias.

### Declaracion de fucniones

Cuando el compilador está compilando el script y se encuentra con la declaración de una función, tiene que dejar lo que estaba haciendo y pasar a compilar el cuerpo de esa función, porque va a tener su propio chunk y sus propias variables locales. Cuando termina, tiene que volver a donde había quedado. Para resolver esto, el libro plantea tener un compilador nuevo por cada función, donde cada uno referencia al compilador de la función que lo contiene, el `enclosing`, así sabe a dónde volver. Esta referencia también es la que se usa después para buscar variables en las funciones de afuera.

Acá surge una diferencia con el libro, que guarda lo que comparten todos los compiladores en dos variables globales: `parser` (por dónde va leyendo los tokens, los errores que encontró y si está en modo pánico) y `current` (el compilador de la función que se está compilando), que es un puntero que hay que ir cambiando a mano: al entrar a una función apunta a su compilador, y al terminarla hay que volver a ponerlo en el de afuera. Es algo fácil de olvidar, y que después rompe todo.

En Go no hace falta, porque cada compilador es un objeto con sus propios métodos, y un método trabaja sobre el compilador con el que lo llamamos. Por ejemplo, para compilar el cuerpo de una función se llama a `fc.block()`, y todo lo que pasa ahí adentro es sobre `fc`, el compilador de esa función. Cuando termina, seguimos con el compilador de afuera sin haber cambiado nada, así que nunca hay que acordarse de volver para atrás. Con el parser pasa algo parecido: como cada línea del REPL es una compilación nueva, una variable global habría que reiniciarla cada vez.

Entonces lo que se comparte lo juntamos en `parserState`, que usan todos los compiladores por puntero, y lo propio de cada función queda en su `Compiler`. Se comparte por puntero porque tiene que haber un solo `cursor`, una sola lista de errores y un solo modo pánico. Si cada función tuviera una copia, cuando termina el cuerpo de una función el compilador de afuera no sabría por dónde seguir leyendo.

### Parametros

Los parámetros de una funcion para nosotros son varibles locales de la propia función, y cada uno suma a la aridad de la misma. Asi que los parámetros ocupan las primeras posiciones que siguen depues de la funcion misma, que es donde la vm deja los agumento cuando se llama: no hace falta copiar nada ahi. El maximo sigue siendo de 255 parametro igual que la primera parte del tree-walk.

### Llamadas

Para compilar una llamada como `add(1, 2)`, primero se compila lo que se quiere llamar (`add`, en este caso), después los argumentos de izquierda a derecha, y al final se emite un `OP_CALL` con la cantidad de argumentos como operando. Como esa cantidad va en un byte, el máximo es de 255 argumentos, igual que en el tree-walk.

Para que el compilador pueda reconocer la llamada, el `(` pasa a ser un operador infijo en el Pratt parser, con la precedencia de llamada.

### Return

El return es un statement que, si tiene una expresión, la compila y emite un `OP_RETURN`, y si no, devuelve `nil`: emite un `OP_NIL` y un `OP_RETURN`, igual que lo que se agrega al final de cada función. Acá hay una diferencia importante con el tree-walk: allá el recorrido era recursivo, y para salir de varias llamadas anidadas usamos el truco de que el `ReturnValue` suba como si fuera un error. En la VM eso no hace falta, porque el bytecode es una secuencia plana y el return es una instrucción más, que corta la llamada y vuelve al frame anterior.

Un return fuera de una función no tiene sentido, porque no hay nada de dónde salir. Para detectarlo, el compilador se fija si lo que está compilando es el script, y si es así, da un error. Ese error no corta la compilación, sino que sigue leyendo para avisar si hay más errores, porque el código está bien escrito y solo rompe una regla del lenguaje. Con esto nos acoplamos al libro, que dice que se podría permitir, pero elige que sea un error.

## Closures

### Variables de otra función

Al definir las funciones, permitíamos que estas usaran las variables declaradas dentro de su propio cuerpo (locales) o las globales, pero como las funciones se pueden anidar, podrían usar variables que son locales de otra función que las envuelve, y ahí no calificarían ni como globales ni como locales de la función más interna. Además, la función interna se podría devolver o guardar en una variable de afuera y seguir viva después de que la función que la envuelve retornó y su ventana se sacó del stack, junto con la variable que la interna necesitaba.

Otro problema es que la declaración de una función interna se ejecuta una vez por cada llamada a la de afuera, y cada una de esas llamadas tiene sus propias variables. Entonces la función interna tiene siempre el mismo código, pero cada vez tiene que usar variables distintas, y no alcanza con guardarlas en la función compilada porque hay una sola, y terminarían usando las mismas variables.

En el tree-walk interpreter esto no pasaba porque cada scope era un `Environment` en el heap y cada función guardaba una referencia al suyo, así que seguía vivo mientras se lo usara.

### Function y Closure

Para resolverlo el libro separa la función en dos:

- **`Function`:** la arma el compilador, una sola por cada declaración. Tiene el código de la función y queda como constante en el chunk de la función que la contiene.
- **`Closure`:** lo crea la VM cada vez que se ejecuta la declaración. Tiene un puntero a la `Function` y las variables de afuera que esa función usa (**upvalues**).

Así el código de la función es el mismo, pero cada closure tiene sus propias variables.

### Upvalues

Ahora, cuando compilamos el cuerpo de una función y encontramos una variable, buscamos en este orden: es local? → es un upvalue? → es global?. Por este orden, si una local tiene el mismo nombre que una variable de afuera, gana la local (como pasaba antes entre locales y globales).

Un **upvalue** es una variable local pero externa: la función la usa, pero no la declaró ella, sino otra función que la envuelve.
Para validar si es un upvalue, primero se busca en las locales del `enclosing` y, si no está, se ve si el `enclosing` la tiene como upvalue propio. Para que esto funcione, una función captura no solo las variables de afuera que usa ella, sino también las que usan las funciones declaradas dentro suyo. Esto es porque cuando se crea el closure de una función interna, solo se está ejecutando la que la envuelve directamente, y las de más arriba podrían haber retornado, así que la del medio es la única que puede pasarle la variable.

### Cambios en el compilador

Lo primero que tuvimos que hacer es dejar de emitir `OP_CONSTANT` para las funciones y emitir `OP_CLOSURE`, para que la VM distinga y sepa que tiene que crear el closure.

Luego, al compilar las funciones, cada vez que detectamos un upvalue lo guardamos en un nuevo slice **`upvalues`**. Por cada upvalue, nos guardamos un booleano que indica si es una local de la función que la envuelve o un upvalue de ella (la variable viene de más afuera), y su índice en las locales o en los upvalues de esa función, según el caso. Al igual que con las locales, no solo las guardamos sino que emitimos `OP_GET_UPVALUE` / `OP_SET_UPVALUE`, con la posición en ese slice como operando, para que la VM sepa dónde ir a buscarla.

Como la VM no tiene acceso al `Compiler`, después del `OP_CLOSURE` el compilador escribe la lista de upvalues de la función como pares de bytes `<isLocal> <index>`, que la VM usa para capturar las variables al crear el closure. Y como el `Compiler` desaparece al terminar de compilar, cada `Function` guarda la cantidad de upvalues que tiene, así la VM sabe cuántos pares tiene que leer luego de un `OP_CLOSURE`.

Si una función usa la misma variable varias veces, se guarda un solo upvalue para no ocupar lugares de más en el slice. Como su posición se emite en un byte, una función puede tener hasta 256 upvalues, y esto queda como **limitación conocida**.

### Cerrar upvalues

El último problema a resolver era qué pasa cuando un bloque termina o una función retorna y sus variables salen del stack: si alguna estaba capturada como upvalue, el closure quedaría apuntando a un lugar inválido. El libro trata de aprovechar lo máximo posible la velocidad del stack, entonces usa la variable ahí mientras pueda y recién cuando está por salir la pasa al heap.

Para eso, cuando el compilador encuentra un upvalue que es local de la función que la envuelve, marca esa local como `IsCaptured`. La marca queda en las locales de la función que declaró la variable (el enclosing de la interna), que es quien luego la va a sacar del stack. Si el upvalue hallado no es local de la de afuera, no hay que marcarlo porque ya se marcó un nivel más arriba.

Entonces al terminar un bloque, por cada local emite `OP_CLOSE_UPVALUE` si fue capturada por una función más interna y `OP_POP` si no, indicándole a la VM que tiene que moverla al heap antes de sacarla, para que luego pueda ser usada. Cuando una función retorna, el compilador no tiene que hacer nada distinto, porque `OP_RETURN` saca todas sus locales de una vez moviendo la ventana y no hay una instrucción por variable que reemplazar (eso lo resuelve la VM cuando lee el `OP_RETURN`).
Esto también se aplicó a `break` y `continue`, que al sacar las locales del ciclo antes de saltar también emiten `OP_CLOSE_UPVALUE` para las que estén capturadas.

## Clases

### Declaracion de clases

Compilar una clase se hace como una declaración de una variable pero usando `OP_CLASS`: crea la clase y después se define como una variable global o queda como local en su slot. Si queda como local, las reglas que ya tenemos siguen valiendo para las clases, como el _shadowing_ para una variable de afuera, y que es un error repetir el nombre en el mismo scope local. Y en cuanto a si queda global, vale la regla de que sí se puede redeclarar como pasa cuando usamos `var`.

El nombre de la clase va como constante aunque sea una local, porque `OP_CLASS` lo necesita para que la clase sepa cómo se llama. Con `var` o `fun`, si la variable es local, `parseVariable` no agrega su nombre a las constantes, porque se lee por su posición en el stack. Para la clase eso no sirve, así que, como hace el libro, `classDeclaration` usa `identifierConstant` para guardar el nombre y después la declara con `declareLocalVariable`.

Una diferencia con el libro es que el `consume` del libro reporta el error y la función sigue, pero en Go devuelve el error y hay que cortar, es por eso que `classDeclaration` devuelve error y en el libro tienen un `void`.

### Get y set de propiedades

Agregamos el **punto**, que es un operador infijo: a la izquierda tiene una expresión, la instancia, pero a la derecha no tiene una expresión sino un nombre. En `x.a`, la `a` no es una variable, sino un nombre que guardamos entre las constantes para que después la VM pueda buscarlo en la instancia.

El punto tiene la misma precedencia que los paréntesis de una llamada, así que se lee de izquierda a derecha: en algo como `a.f().x`, se aplican en el orden en que aparecen. Además, tiene más precedencia que el `-`, así que en `-a.x` primero se lee `a.x` y después se niega.

Para distinguir el set del get, `dot()` se fija si se puede asignar y si viene un `=`. Si es así, compila el valor y emite `OP_SET_PROPERTY`. Si no, es un get y emite `OP_GET_PROPERTY`. Acá también se usa `canAssign`, y lo necesitamos por el ejemplo del libro: sin el `a + b.c = 3` se leería como `a + (b.c = 3)`, que la gramática no permite. Con `canAssign`, `dot()` deja el `=` sin consumir, y el error lo termina dando `parsePrecedence` cuando lo encuentra. En cambio, en `f().x = 1` `canAssign` es verdadero y la asignación funciona. Como la asignación deja el valor en el stack, también se puede encadenar, como en `a.x = a.y = 1`.

Una diferencia con el libro que ya mencionamos en [Nombres repetidos](#nombres-repetidos) es que, cuando aparece un nombre, el libro agrega una constante nueva y nosotros reutilizamos la que ya existe, y acá se vuelve a aprovechar. Esto ayuda a no llegar tan rápido al límite de 256 constantes por función. Pasa tanto en `a.x + a.x`, que usa la misma constante para la `x`, como en `x.x` si `x` es global: la variable y el campo comparten la constante, porque se guarda por nombre.

### Métodos

Al agregar los métodos nos aparece el problema de que una clase puede tener varios métodos y eso tiene que llegar a la vm como una serie de instrucciones de bytecode y no como un nodo del árbol que recorría antes el tree-walk. En el libro habla de que para el programador declarar una clase es una sola cosa, pero para nosotros el compilador lo emite como una serie de instrucciones que la vm ejecuta de a una.

Primero tenemos ya el OP_CLASS que crea la clase vacía, después se guarda en su variable y se vuelve a agregar en el stack porque la vm la necesita ahí para agregarle los métodos. Si la clase es global al guardarla con OP_DEFINE_GLOBAL sale del stack y por eso hay que volver a cargarla, y si es local queda en su slot. Después por cada método se emite el OP_CLOSURE que deja el closure del método arriba de la clase en el stack y le sigue el OP_METHOD que lo guarda en la clase con su nombre y lo saca del stack, así la clase sigue quedando arriba para el siguiente. Y al final un OP_POP que saca a la clase, que ya quedó guardada en su variable con sus métodos.

Además para agregarle métodos a la clase la vm necesita tener la clase a mano pero el compilador no puede saber si es global o local y por ende dónde queda, por eso usa el nombre para volver a cargarla con el namedVariable, que es el mismo que para una variable, y eso genera la instrucción que corresponda según el caso.

Un detalle sobre métodos es que Lox no tiene declaración de campos para la clase así que todo lo que está entre {} es un método, así que si aparece otra cosa es un error. Se termina compilando con la misma función que para la fun así que termina quedando como un closure y puede usar las variables de afuera.

### This

Con los métodos ya andando faltaba poder usar `this` adentro y en el libro lo resuelve aprovechando el slot 0 de cada función. En una función común ahí queda la función que se está ejecutando pero en un método eso no hace falta, así que la vm pone en ese lugar la instancia sobre la que se llamó y el compilador le pone de nombre `this` a ese slot. Entonces `this` termina siendo una variable local más que se compila con `OP_GET_LOCAL 0`, y como se compila sin poder asignarse, surge que algo como `this = 1` da error igual que pasa con `a + b.c = 3`. Para saber cuándo hacer esto agregamos un tipo de función nuevo, `TYPE_METHOD` que es el que usa `method()` al compilar cada método.

El slot 0 se llama `this` en todo lo que no sea una función común y esto tiene sentido cuando hay una función adentro de un método que usa `this` porque ahí se espera que sea el del método. Como la función no tiene uno el compilador no lo encuentra y lo sigue buscando afuera hasta encontrarlo en el método que lo termina encontrando como un upvalue con lo mismo que ya teníamos de [Closures](#closures). Si la función también tuviera su propio `this`, taparía al del método y se rompería ese caso.

Siguiendo la idea del libro tenemos el error de compilacion cuando se usa this fuera de una clase, y un detalle con eso es que fijarse si estamos en un método no alcanza por lo de la función adentro de un método, entonces lleva una pila con las clases que se están compilando donde cada una apunta a la de afuera, el libro dice que no es comun pero Lox deja declarar una clase adentro de un método de otra. Y bueno la idea es que si aparece un `this` con la pila vacía, es un error.

### Init

Con la vm llamando al `init` nos aparecía un problema porque el `init` se compila como cualquier método y termina con el `OP_NIL` y el `OP_RETURN` que pone el compilador al final, y el `OP_RETURN` deja lo que devuelve donde el que llamó espera el resultado, entonces `var a = Auto("rojo");` terminaba guardando `nil` en `a` y no la instancia. Lo que hace el libro es resolverlo en el compilador haciendo que el `init` siempre devuelva `this` que ya está en el slot 0, así la vm no tiene que hacer nada distinto.

Teniendo en cuenta lo que hicimos con `this` esto fue sencillo porque agregamos otro tipo de función `TYPE_INITIALIZER` que lo usa `method()` cuando el método se llama `init`, y como no es una función común el slot 0 también se llama `this`. Lo que sí cambia es que `emitReturn()` en vez de emitir el `OP_NIL` emite un `OP_GET_LOCAL 0`, y es por esto que el `return;` sin valor también devuelve `this` porque usa el mismo `emitReturn()` que el final de la función entonces se puede salir antes de un `init` y la instancia no se pierde.

Lo que no se puede es hacer un `return` con un valor adentro del `init` porque se devolvería ese valor y no la instancia, así que igual que en el libro es un error de compilación. Lo levantamos con `report` y no devolviendo el error como pasa con el `return` en el script, porque el código está bien escrito y solo rompe una regla del lenguaje así que sigue compilando y muestra los demás errores.

Un detalle es que cada compilador mira su tipo entonces si hay una función adentro del `init` esa sigue siendo `TYPE_FUNCTION` y puede hacer `return 1;` sin problema porque ese return es de la función y no del `init`. Y otra cosa es que el `init` se puede llamar como cualquier método.

Una diferencia con el libro es que para saber si el método se llama `init` en C tienen que comparar el largo y los bytes porque el token apunta al código y nosotros como el lexema ya es un `string` lo comparamos con `==`. 

Una diferencia con el libro es que guarda la clase actual en una variable global y nosotros en `parserState` que es lo que ya comparten por puntero todos los compiladores (ver [Funciones](#funciones)), así un método que tiene su propio compilador ve la misma clase que la declaración de afuera. Pero la diferencia más grande es cuándo se saca la clase de la pila porque el libro lo hace al final de la declaración y le funciona, porque en C cuando hay un error de sintaxis lo reporta y la función sigue de largo hasta el final. En glox un error de sintaxis corta la función con `return err`, así que con una clase con error nunca se llegaría a esa última línea entoences la clase quedaría en la pila y un `this` suelto más abajo ya no daría error. Por eso la sacamos con un `defer`, que se ejecuta siempre al salir de la función sea como sea y así el resultado es el mismo que en el libro aunque lo hayamos hecho de otra forma.

## Tests

> Los tests unitarios del compilador se pueden ejecutar de la siguiente forma:

```bash
go test ./internal/compiler
```

> Para las clases agregamos los tests ([6-classes](../real-tests/6-classes.lox)) que solo corren con la VM, porque el tree-walk no tiene clases.

```bash
python3 real-tests/script.py "go run ./cmd --vm" 6-classes
```

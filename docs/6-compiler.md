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

Además nos guardamos el anterior, que el compilador usa para saber en qué línea está cada instrucción que emite. Al avanzar se guarda lo que se consumió como anterior, así que siempre existe uno y eso cubre un caso más borde que nos importa, el caso de que llega solo un `EOF`, que sería un programa vacío, ahí no se rompe, como sí pasaría si calculáramos el anterior como `actual - 1`, porque ahí no habría nada en el anterior.

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

Los números y los strings se compilan de la misma forma, se guarda el valor entre las constantes del chunk y se crea un `OP_CONSTANT` con su índice. Es por esto mismo que en la tabla los números y strings apuntan a la misma función `value()`, que agarra el valor que el scanner dejó en el token y lo deja como constante.
Una diferencia con el libro es que para los strings, el compilador tiene que sacar las comillas, pero para nuestra implementación eso no es necesario porque eso ya lo hace el scanner. Y otra cosa interesante es que en el libro necesita copiar los caracteres al heap para que el mismo string sea dueño de su memoria, pero nosotros en glox no hacemos eso porque en Go los strings son inmutables así que se pueden compartir sin copiarlos.

Tenemos el límite que el mismo libro habla, sobre que la posición de cada constante ocupa un solo byte así que un chunk puede tener como máximo 256 constantes. Quizás acá es un punto a mejorar, pero por ahora lo que tenemos es que para gastar menos lugar los nombres de las variables no se repiten y eso lo explicamos más a detalle en la sección de [nombres repetidos](#nombres-repetidos).

#### Asociatividad

Como vimos, cada operador tiene en la tabla una función que se ejecuta cuando el compilador lo encuentra. Las que nos interesan para la asociatividad son las de los operadores: `binary()` para los binarios, `unary()` para los unarios y `variable()`, que maneja la asignación cuando después de un nombre encuentra un `=`.

Cuando se llama a una de esas funciones, lo que está a la izquierda del operador ya se compiló (o no existe, en el caso de los unarios). Queda compilar lo que viene después del operador y emitir la instrucción del operador. Para eso, cada función llama a `parsePrecedence` con una precedencia mínima, que definirá la asociatividad. Si aparece otro operador del mismo nivel (con la misma precedencia), puede pasar una de dos cosas:

- Si la función pide una precedencia **mayor** que la del operador actual, un operador que tenga el mismo nivel no alcanzaría el mínimo, así que no entra a la expresión: queda para la de afuera. La asociatividad es **a izquierda**.
- Si pide la **misma o una menor**, el siguiente operador sí alcanza el mínimo, así que entra y se resuelve primero. La asociatividad es **a derecha**.

### Statements

Con los statements, un programa pasa a ser una lista de declaraciones que se compilan hasta llegar a un `EOF`. Acá aparecen dos funciones: `declaration`, que si lo que viene empieza con `var` compila la declaración de la variable, y si no se lo pasa a `statement`. Esta última compila el resto de las sentencias: `print`, bloques, `if`, ciclos y las expresiones que terminan en `;`. Las separamos así porque, como explica el libro, en Lox una declaración no puede ser el cuerpo de un `if` o un `while`, ya que no quedaría claro hasta dónde llega la variable.

Es importante ver el efecto que tiene cada statement sobre el stack de la VM (su _stack effect_). La idea de fondo es que una expresión siempre deja un valor más en el stack, y un statement tiene que dejarlo como estaba. Por ejemplo, en `print 1 + 2;` se apila el 1, después el 2, y `OP_ADD` los saca y apila el 3: la expresión dejó un valor más. Después, `OP_PRINT` saca ese resultado y lo imprime, así que el statement termina dejando el stack como estaba. Si en cambio solo se escribe `1 + 2;`, el compilador emite un `OP_POP` para tirar el valor que dejó la expresión. En el tree-walk, ese descarte lo hacía Go, ignorando el valor devuelto; ahora es una instrucción más.

Esto es importante porque, si cada statement dejara algo en el stack, con muchos statements seguidos el stack crecería constantemente. Además, es lo que permite que los índices de las variables locales coincidan con su posición en el stack (ver [variables locales y scopes](#variables-locales-y-scopes)). La única excepción es la declaración de una variable local, que deja su valor en el stack a propósito, porque ese valor es la variable, y se saca recién al cerrar su bloque.

Para asegurarnos de que esto se cumpla, cuando termina el script la VM verifica que el stack haya quedado vacío y si no levanta un error, porque significa que el compilador dejó algo de más.

### Manejo de errores

Acá llegó un gran cambio frente al tree-walk de la primera parte: este, al detectar un error, cortaba y solo mostraba ese, pero no seguía para ver si habían más errores y devolverlos todos a la vez. Si bien funciona, acá creemos que llega un poco más de salto de calidad en cuanto a la experiencia que se puede percibir de esta segunda parte, porque la idea es seguir así se muestran todos los errores y se pueden corregir todos juntos. Si no, en un caso de muchos errores, te mostraría el primer error, el programador tendría que corregirlo, y luego vería otro error que ya existía pero venía después del primero, así que no se había enterado.
Pero seguir después del primero trae un detalle a tener en cuenta, el compilador queda parado en una línea rota, que si sigue leyendo levantaría errores falsos propios de donde retoma, entonces para eso entra la **sincronización** que es avanzar hasta llegar a que empiece un statement nuevo o un `var`, `print` y desde ahí recién seguir compilando.

Una situación que nos surgió es primero pensar en que para devolver errores nos podemos aprovechar de cortar el statement con el error y subirlo, algo típico para Go, el tema es que cuando llegaron las variables locales por los scopes esto nos trajo un problema: al querer sincronizar lo hacíamos dos veces, una dentro del scope local y otra al salir. Esto introducía un error falso de que faltaba un `}`. Ahí, decidimos pegarnos un poco más al libro con la idea del _flag_, pero con la diferencia de que en vez de mostrar ahí el error nos vamos guardando los errores. Entonces el flag se prende cuando hay un error, mientras está prendida los errores nuevos se ignoran, y cuando se sincroniza y arranca el statement nuevo se apaga, así cada error solo lo vemos una vez y donde corresponde. 
Una diferencia más con el libro es que, en C, reportar un error no corta la ejecución: la función sigue leyendo tokens. En Go, en cambio, devolver un error corta la función en ese punto. Por eso tenemos dos caminos: cuando el código está mal escrito y no se puede seguir leyendo (por ejemplo, falta un `)`), devolvemos el error y se sincroniza. Cuando el código está bien escrito pero rompe una regla (como redeclarar una variable local en el mismo scope), lo reportamos con `report` y seguimos compilando normalmente. Es lo que hacemos con los errores de las variables locales.

Al terminar de compilar, si hubo errores, se devuelven todos juntos en un solo error, así se pueden ver todos de una vez en lugar de ir arreglándolos de a uno. Además, en ese caso no se devuelve el chunk: el bytecode generado puede estar incompleto (por ejemplo, con saltos sin completar), así que la VM no llega a ejecutar nada.

### Variables globales

En la primera parte una variable global era una entrada más del `Environment` y el intérprete la buscaba por nombre cuando la necesitaba. Acá el compilador no se guarda el valor de la variable, lo que hace es emitir las instrucciones para que la VM se encargue cuando se ejecuta. Las globales en Lox se resuelven tarde (_late bound_), es decir que se buscan por nombre recién en la ejecución. El tema es que un nombre es un string y no entra como operando de una instrucción, que es de un solo byte. Para resolver esto, el nombre se guarda entre las constantes del chunk y la instrucción lleva la posición donde quedó. Y para saber si una variable es global pasa lo mismo que con el resolver de la primera parte, si no se encuentra en ningún scope local se considera global.

#### Declaración

Cuando el compilador se encuentra con `var a = 1;` primero guarda el nombre `a` entre las constantes del chunk, después compila el valor, y recién al final emite `OP_DEFINE_GLOBAL` con la posición del nombre. Va al final porque cuando la VM ejecute esa instrucción necesita que el valor ya esté en el stack para poder guardarlo. Si no tiene valor, como en `var a;`, se emite un `OP_NIL` en su lugar, así que toda variable declarada existe y vale `nil` hasta que se le asigne algo, igual que en el tree-walk. Otro detalle que también es igual es que declarar de nuevo una global no da error, solo pisa el valor que tenía.

#### Lectura

Cuando aparece un identificador el compilador emite `OP_GET_GLOBAL` con la posición del nombre y nada más. Importante es ver que acá no se fija si la variable existe, eso se sabe recién cuando la VM ejecuta la instrucción y busca el nombre en su tabla de globales, y si no está levanta un error de runtime. Es decir que `print a;` compila aunque `a` no esté declarada, y eso es justamente lo que significa que sea _late bound_. El libro lo deja así porque más adelante una función puede usar una global que se declara más abajo en el código, y lo único que importa es que exista cuando esa función se ejecute.

#### Asignación

Si después del nombre viene un `=`, en vez de leer la variable el compilador compila el valor nuevo y emite `OP_SET_GLOBAL`. Como la asignación es una expresión su valor queda disponible, y por eso se puede encadenar `a = b = 5`, que se resuelve de derecha a izquierda como en el parser de la primera parte.

El tema acá es que como compilamos en una sola pasada, cuando se lee `a` todavía no se sabe si es una lectura o el lugar donde se asigna algo. En el parser de la primera parte podíamos armar el lado izquierdo y después fijarnos si venía un `=`, pero acá ya se emitió el código de lo que está a la izquierda, entonces el libro lo resuelve mirando si viene un `=` justo después del nombre. Pero eso solo no alcanza, en `a * b = c;` la `b` está adentro de la multiplicación, vería el `=` y se lo quedaría, y terminaría armando `a * (b = c)` cuando en realidad es `(a * b) = c`, que no es válido. Para evitarlo, `parsePrecedence()` calcula un `canAssign`, que es verdadero solo si estamos en el nivel más bajo de la expresión, y se lo pasa a cada función de la tabla. Entonces la `b` ahí sabe que no puede agarrar el `=`, y cuando termina la expresión y queda un `=` sin agarrar, el compilador levanta `invalid assignment target`, que es el mismo error del parser de la primera parte.

#### Nombres repetidos

En el libro cada vez que aparece una variable se guarda su nombre de nuevo entre las constantes, y deja como ejercicio evitarlo. Sabemos que con el límite de 256 constantes esto es un problema si en un programa se usa muchas veces la misma variable, porque cada uso gasta un lugar, por lo tanto lo quisimos resolver. En el compilador decidimos crear un `map` de nombre a posición para saber si una variable ya apareció, porque nos pareció la forma más simple y que no toca nada de la VM. Si ya apareció, reusamos esa posición en vez de guardarla otra vez, así que a la VM no le cambia nada porque sigue recibiendo una posición que apunta al nombre correcto. El costo que nos queda es una consulta al `map` cada vez que se usa una variable al compilar. Los números y strings literales no los tocamos, esos se siguen guardando cada vez que aparecen.

### Variables locales y scopes

En nuestra implementación anterior, en el manejo de scopes tenían un rol importante el resolver y el intérprete. El resolver ayudaba a dejar definido antes del runtime a qué distancia estaba cada variable definida, para que luego el intérprete usara eso en vez de calcularlo en ejecución. Con esa distancia, el intérprete subía hasta el entorno correspondiente y ahí buscaba la variable por nombre.

En esta nueva versión, el compilador es quien se encarga de calcular antes de la ejecución, en qué índice del stack de la VM va a estar cada variable local. Para eso, el compilador guarda un `scopeDepth`, que indica para cada momento la profundidad actual. Entonces, cuando se encuentra con un `block statement`, inicia un nuevo **scope**, que es simplemente sumar uno al contador de profundidad. Luego, sigue compilando todo lo que está dentro del bloque. Si lo que hay dentro del bloque son variables locales, las va a guardar en su **"stack falso"** `locals`. Cada elemento de este stack es una `Local`, compuesta por su nombre y su profundidad. La idea es que el índice de cada variable en este nuevo stack coincide con el índice que tendrá en el stack de la VM. Esto funciona porque cada statement deja el stack como lo encontró: los valores temporales de una expresión se apilan arriba de las locales y se sacan antes del siguiente statement.

Cuando encuentra una declaración de una variable local, recorre las locales del scope actual y valida que no haya otra con ese nombre, ya que redeclarar en el mismo scope es un error.
Si no hay, la declara con profundidad `-1`, indicando que aún no fue inicializada (para evitar los `var a = a;`). Luego de ser inicializada la marca con su profundidad correspondiente.

De esta forma, no necesita generar una instrucción como `OP_DEFINE_GLOBAL`, ya que el valor de la variable queda directamente en el stack. Luego, cuando quiera usarla (con `OP_GET_LOCAL` o `OP_SET_LOCAL`), usa como operando su índice en el stack. No necesitamos guardar el nombre en el bytecode, porque la VM accede a las locales por su índice, sin importar su nombre. Además, para saber qué índice usar, el compilador busca el nombre en las locales de atrás para adelante, así siempre encuentra la más interna. Esto permite que una variable de un scope interno tape a otra de afuera con el mismo nombre (*shadowing*).

Al finalizar el bloque, resta uno a la profundidad y elimina de su stack falso las variables locales de ese bloque. Además, emite un `OP_POP` por cada una, para asegurarnos de que al terminar un bloque, el stack de la VM quede sin sus variables.

Un detalle más sobre las variables locales es que puede haber como máximo 255 a la vez, ya que el operando de `OP_GET_LOCAL` y `OP_SET_LOCAL` ocupa un solo byte, que va de 0 a 255, y el slot 0 queda reservado para la función.

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

### Funciones

#### El script como función

Un salto que se da importante al momento de llegar a funciones es que el compilador ya no escribe un chunk ahí suelto, sino que ahora lo hace en una función. Por lo que por este motivo ahora compile pasa a devolver una función. Un detalle es que ahora el programa completo sería como una función implícita. Es por esto que el compilador siempre está dentro de una función y la VM se ejecuta en una. Pero sí que tenemos la observación que esto no es para las globales que tienen un alcance propio. Otro punto a tener en cuenta es que ahora sí distingue entre una función y el script.

Algo importante que se agrega con todo esto es lo del slot 0 en el stack, reservado para que la VM ponga la función que está ejecutando, y por ende, como el operando es de un byte hay 256 lugares y el slot 0 ocupa uno, ahora pasa a soportar hasta 255 variables locales, a diferencia del tree-walk, por lo que esto pasa a ser una **limitación conocida que tenemos en esta versión**.

#### Una funcion con su propio chunk
Agregar funciones hizo que pasemos de que el compilador escriba todo en un único chunk, a que teniendo funciones cada una tiene su propio código entonces la VM tiene que poder pasar de uno a otro cuando se llama a una función. En el libro marca que una función en lox es un valor como un número o string, así que tiene que poder guardarse como una variable. Además tiene un nombre para identificarla, su propio chunk y los parámetros, los cuales chequeamos para saber si recibe los esperados.

En cuanto a una diferencia importante en general con el libro es que en estos capítulos usa C para implementar clox, por lo que tiene que hacer manejo de memoria, pero nosotros en glox lo usamos como un struct común que luego se libera por el garbage collector propio de Go cuando no tiene más referencias.

#### Declaracion de fucniones
Cuando el compilador esta compilando el script y se topa con una funcoin tiene que dejar ahi y pasar a compialr el cuempo de esa funcion porqu eva a tener su chunk, su variables locales yc uando termine tiene que vovler al scrip donde habia quedado. Pra resolver esto el libro plantea tener un compilador neuvo por cad afuncion y cad auno cuarda una referecia al compilador de la funcion que lo continee que eso es el enclosing., asi soabe donde vovler. 

Aca suge una diferencia con el libro, porque ahhi propone guardar lo que comparten todos los compiladores en dos variables globales, parser (por dónde va leyendo los tokens, los errores que encontró y si está en modo pánico) y current (el compilador de la función que se está compilando, que se cambia al entrar a una función y se vuelve a poner al salir).

Nosotros lo hicimos distinto al libro donde current es un puntero que hay que ir cambiando a mano, apunta al compilador de la función hasta que la termina y después hay que volver a ponerlo en el de afuera. Es algo fácil de olvidarse y que después rompe todo. En Go no hace falta, porque cada compilador es un objeto con sus propios métodos, y cuando llamamos a un método, ese método ya trabaja sobre el compilador con el que lo llamamos. Por ejemplo para compilar el cuerpo de una función se llama a fc.block(), y ahí adentro todo lo que se haga es sobre fc, que es el compilador de esa función. Y cuando termina seguimos con el compilador de afuera sin haber cambiado nada. Así nunca hay que acordarse de volver para atrás. Con el parser es algo parecido, porque como cada linea del REPL es una compialcion neuva una variable global habrai que reiniciarla cada vez.

Entonces lo que se comparte lo juntamos en parserState, que usan todos los compiladores por puntero, y lo propio de cada función queda en su Compiler. Se comparte por puntero porque tiene que haber un solo cursor, una sola lista de errores y un solo modo pánico. Si cada función tuviera una copia, cuando termina el cuerpo de una función el compilador de afuera no sabría por dónde seguir leyendo.

#### Parametros

Los parametros de una funcion para nosotros son varibles locales de la funcion, y cada uno suma a la aridad de la misma. Asi que los aprametros ocupan las primeras posiciones que siguen depues de la funcion misma que es donde la vm deja los agumento cuando se llma asi que no hace falta copiar anda ahi. El maximo sigue siendo de 255 parametro igual que la primera parte del tree-walk.

#### Llamadas
Una llamada a la funcion como por ejemplo add(1,2) para compilarla, primero se compila lo que se queire llamar que es add en este caso, despues los argumentos de izquierda a derecha y al final se emite un OP_CALL que lleva al cantidad de argumentos. Como esa cantidad va en un byte enteonces la maxima cantidad es de 255 igual que en el tree walk. 
Para qu el compilador puede reconocer la llmada , el '(' pasa a ser un operador infijo en el pratt parser con la precedencia de llamada.  

#### Return

El return es un statement que si tiene una expresión la compila y emite un OP_RETURN, y si no tiene devuelve nil, es decir que emite un OP_NIL y un OP_RETURN igual que lo que se agrega al final de cada función. Acá hay una diferencia importante con el tree-walk, porque allá el recorrido era recursivo y para salir de varias llamadas anidadas usamos como un truco que era que el ReturnValue sube como si fuera un error. En la VM eso no hace falta porque el bytecode es una secuencia plana y el return es una instrucción más, que corta la llamada y vuelve al frame anterior como vimos en return de la VM.

Un return fuera de una función no tiene sentido porque no hay nada de donde salir. Para detectarlo el compilador se fija si lo que está compilando es el script y si es así da un error. Ese error no corta la compilación si no que sigue leyendo para avisar si hay más errores porque el código está bien escrito y solo rompe una regla del lenguaje. Con esto nos acomplamso a lo que dice el libro que se podria permitir esto pero elige que sea un error. 

### Closures

#### Variables de otra función

Al definir las funciones, permitíamos que estas usaran las variables declaradas dentro de su propio cuerpo (locales) o las globales, pero como las funciones se pueden anidar, podrían usar variables que son locales de otra función que las envuelve, y ahí no calificarían ni como globales ni como locales de la función más interna. Además, la función interna se podría devolver o guardar en una variable de afuera y seguir viva después de que la función que la envuelve retornó y su ventana se sacó del stack, junto con la variable que la interna necesitaba.

Otro problema es que la declaración de una función interna se ejecuta una vez por cada llamada a la de afuera, y cada una de esas llamadas tiene sus propias variables. Entonces la función interna tiene siempre el mismo código, pero cada vez tiene que usar variables distintas, y no alcanza con guardarlas en la función compilada porque hay una sola, y terminarían usando las mismas variables.

En el tree-walk interpreter esto no pasaba porque cada scope era un `Environment` en el heap y cada función guardaba una referencia al suyo, así que seguía vivo mientras se lo usara.

#### Function y Closure

Para resolverlo el libro separa la función en dos:

- **`Function`:** la arma el compilador, una sola por cada declaración. Tiene el código de la función y queda como constante en el chunk de la función que la contiene.
- **`Closure`:** lo crea la VM cada vez que se ejecuta la declaración. Tiene un puntero a la `Function` y las variables de afuera que esa función usa (**upvalues**).

Así el código de la función es el mismo, pero cada closure tiene sus propias variables.

#### Upvalues

Ahora, cuando compilamos el cuerpo de una función y encontramos una variable, buscamos en este orden: es local? → es un upvalue? → es global?. Por este orden, si una local tiene el mismo nombre que una variable de afuera, gana la local (como pasaba antes entre locales y globales).

Un **upvalue** es una variable local pero externa: la función la usa, pero no la declaró ella, sino otra función que la envuelve.
Para validar si es un upvalue, primero se busca en las locales del `enclosing` y, si no está, se ve si el `enclosing` la tiene como upvalue propio. Para que esto funcione, una función captura no solo las variables de afuera que usa ella, sino también las que usan las funciones declaradas dentro suyo. Esto es porque cuando se crea el closure de una función interna, solo se está ejecutando la que la envuelve directamente, y las de más arriba podrían haber retornado, así que la del medio es la única que puede pasarle la variable.

#### Cambios en el compilador

Lo primero que tuvimos que hacer es dejar de emitir `OP_CONSTANT` para las funciones y emitir `OP_CLOSURE`, para que la VM distinga y sepa que tiene que crear el closure.

Luego, al compilar las funciones, cada vez que detectamos un upvalue lo guardamos en un nuevo slice **`upvalues`**. Por cada upvalue, nos guardamos un booleano que indica si es una local de la función que la envuelve o un upvalue de ella (la variable viene de más afuera), y su índice en las locales o en los upvalues de esa función, según el caso. Al igual que con las locales, no solo las guardamos sino que emitimos `OP_GET_UPVALUE` / `OP_SET_UPVALUE`, con la posición en ese slice como operando, para que la VM sepa dónde ir a buscarla.

Como la VM no tiene acceso al `Compiler`, después del `OP_CLOSURE` el compilador escribe la lista de upvalues de la función como pares de bytes `<isLocal> <index>`, que la VM usa para capturar las variables al crear el closure. Y como el `Compiler` desaparece al terminar de compilar, cada `Function` guarda la cantidad de upvalues que tiene, así la VM sabe cuántos pares tiene que leer luego de un `OP_CLOSURE`.

Si una función usa la misma variable varias veces, se guarda un solo upvalue para no ocupar lugares de más en el slice. Como su posición se emite en un byte, una función puede tener hasta 256 upvalues, y esto queda como **limitación conocida**.

#### Cerrar upvalues

El último problema a resolver era qué pasa cuando un bloque termina o una función retorna y sus variables salen del stack: si alguna estaba capturada como upvalue, el closure quedaría apuntando a un lugar inválido. El libro trata de aprovechar lo máximo posible la velocidad del stack, entonces usa la variable ahí mientras pueda y recién cuando está por salir la pasa al heap.

Para eso, cuando el compilador encuentra un upvalue que es local de la función que la envuelve, marca esa local como IsCaptured. La marca queda en las locales de la función que declaró la variable (el enclosing de la interna), que es quien luego la va a sacar del stack. Si el upvalue hallado no es local de la de afuera, no hay que marcarlo porque ya se marcó un nivel más arriba.

Entonces al terminar un bloque, por cada local emite `OP_CLOSE_UPVALUE` si fue capturada por una función más interna y `OP_POP` si no, indicándole a la VM que tiene que moverla al heap antes de sacarla, para que luego pueda ser usada. Cuando una función retorna, el compilador no tiene que hacer nada distinto, porque `OP_RETURN` saca todas sus locales de una vez moviendo la ventana y no hay una instrucción por variable que reemplazar (eso lo resuelve la VM cuando lee el `OP_RETURN`).
Esto también se aplicó a `break` y `continue`, que al sacar las locales del ciclo antes de saltar también emiten `OP_CLOSE_UPVALUE` para las que estén capturadas.

### Tests

Los tests unitarios del compilador se pueden ejecutar de la siguiente forma:

```bash
go test ./internal/compiler
```

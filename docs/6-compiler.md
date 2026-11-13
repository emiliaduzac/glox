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

Los números y los strings se compilan de la misma forma, se guarda el valor entre las constantes del chunk y se crea un `OP_CONSTANT` con su índice. Es por esto mismo que en la tabla los números y strings apuntan a la misma función `value()` que agarra el valor que el scanner dejó en el token y lo deja como constante.
Un detalle de diferencia con el libro es que para los strings el compilador tiene que sacar las comillas, pero para nuestra implementación eso no es necesario porque eso ya lo hace el scanner. Y otra cosa interesante es que en el libro necesita copiar los caracteres al heap para que el mismo string sea dueño de su memoria, pero nosotros en glox no hacemos eso porque en Go los strings son inmutables así que se pueden compartir sin copiarlos.

Tenemos el límite que el mismo libro habla, sobre que la posición de cada constante ocupa un solo byte así que un chunk puede tener como máximo 256 constantes. Quizás acá es un punto a mejorar, pero por ahora lo que tenemos es que para gastar menos lugar los nombres de las variables no se repiten y eso lo explicamos más a detalle en la sección de nombres repetidos.

#### Asociatividad

Como vimos, cada operador tiene en la tabla una función que se ejecuta cuando el compilador lo encuentra. Las que nos interesan para la asociatividad son las de los operadores: `binary()` para los binarios, `unary()` para los unarios y `variable()`, que maneja la asignación cuando después de un nombre encuentra un `=`.

Cuando se llama a una de esas funciones, lo que está a la izquierda del operador ya se compiló (o no existe, en el caso de los unarios). Queda compilar lo que viene después del operador y emitir la instrucción del operador. Para eso, cada función llama a `parsePrecedence` con una precedencia mínima, que definirá la asociatividad. Si aparece otro operador del mismo nivel (con la misma precedencia), puede pasar una de dos cosas:

- Si la función pide una precedencia **mayor** que la del operador actual, un operador que tenga el mismo nivel no alcanzaría el mínimo, así que no entra a la expresión: queda para la de afuera. La asociatividad es **a izquierda**.
- Si pide la **misma o una menor**, el siguiente operador sí alcanza el mínimo, así que entra y se resuelve primero. La asociatividad es **a derecha**.

### Statements

Con los statements ahora pasamos a tener una lista de declaraciones que se compilan hasta llegar a un `EOF`. Acá aparecen dos funciones, una que es `declaration` que si lo que viene empieza con `var` compila la declaración de la variable y si no se lo pasa a `statement`. Esta última básicamente compila el resto, como es `print`, los bloques y las expresiones que terminan en `;`. La separamos así porque en el libro habla de que en Lox una declaración no puede ser el cuerpo de un `if` o `while` porque no quedaría claro hasta dónde llega la variable.

Importante es ver el efecto que generan los statements agregados al stack de la VM, porque cada una cambia esa cantidad del stack. La idea de fondo es que una expresión siempre deja un valor más y un statement tiene que dejar el stack como estaba. Un ejemplo de esto es con `print 1 + 2;` ahí se apila el 1, después el 2, y `OP_ADD` los saca y apila el 3, entonces queda con un valor más a como originalmente estaba, pero cuando llega `OP_PRINT` saca el resultado y lo imprime, entonces el statement termina dejando el stack como estaba. O si solo se hace `1 + 2;` entonces `OP_POP` se emite para tirar el valor que deja la expresión y queda de nuevo como estaba el stack. En la primera parte del tree-walk eso era evaluar y descartar, pero ahora ese descarte lo tenemos como una instrucción. Y lo importante de esto es que al tener muchos statements seguidos si cada uno sigue sumando al stack crecería constantemente y de esta manera no es así.

Un detalle acá es que controlamos que si cuando llega `OP_RETURN` el stack no está vacío entonces levantamos un error.

### Manejo de errores

Acá llegó un gran cambio frente al tree-walk de la primera parte porque este al detectar un error cortaba y solo mostraba ese, pero no seguía para ver si había más errores y devolverlos todos a la vez. Si bien funciona, acá creemos que llega un poco más de salto de calidad en cuanto a la experiencia que se puede percibir de esta segunda parte, porque la idea es seguir así se muestran todos los errores y se pueden corregir todos juntos, si no caería en que te muestra el primer error, el programador lo corrige, y luego le muestra un error que ya existía pero venía después del primero. Pero seguir después del primero trae un detalle a tener en cuenta, el compilador queda parado en una línea rota, que si sigue leyendo levantaría errores falsos propios de donde retoma, entonces para eso entra la sincronización que es avanzar hasta llegar a que empiece un statement nuevo o un `var`, `print` y desde ahí recién seguir compilando.

Una situación que nos surgió es primero pensar en que para devolver errores nos podemos aprovechar de cortar el statement con el error y subirlo, algo típico para Go, el tema es que cuando llegaron las variables locales por los scopes esto nos trajo problema en que al querer sincronizar lo hacíamos dos veces, una dentro del scope local y otra al salir, por lo que introdujo un error falso de que faltaba un `}`. Ahí entonces decidimos pegarnos un poco más al libro con la idea de la flag, pero con la diferencia de que en vez de mostrar ahí el error nos vamos guardando los errores. Entonces la flag se prende cuando hay un error, mientras está prendida los errores nuevos se ignoran, y cuando se sincroniza y arranca el statement nuevo se apaga, así cada error solo lo vemos una vez y donde corresponde. Una diferencia más con el libro es que en C no cortaba el error y sigue leyendo y en Go sí. Entonces tenemos dos caminos, uno de retornar error cuando no se puede seguir leyendo y otro de reportarlo cuando sí, que es lo que hacemos con los errores de las variables locales.

Al final los errores se guardan y se devuelven todos juntos y además no se genera el chunk.

### Variables globales

En la primera parte una variable global era una entrada más del `Environment` y el intérprete la buscaba por nombre cuando la necesitaba. Acá el compilador no se guarda el valor de la variable, lo que hace es emitir las instrucciones para que la VM se encargue cuando se ejecuta. Las globales en Lox se resuelven tarde (_late bound_), es decir que se buscan por nombre recién en la ejecución. El tema es que un nombre es un string y no entra como operando de una instrucción, que es de un solo byte, entonces el nombre se guarda entre las constantes del chunk y la instrucción lleva la posición donde quedó. Y para saber si una variable es global pasa lo mismo que con el resolver de la primera parte, si no se encuentra en ningún scope local se considera global.

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

Un detalle más sobre las variables locales es que puede haber como máximo 256 a la vez, ya que el operando de `OP_GET_LOCAL` y `OP_SET_LOCAL` ocupa un solo byte, que va de 0 a 255.

### Tests

Los tests unitarios del compilador se pueden ejecutar de la siguiente forma:

```bash
go test ./internal/compiler
```

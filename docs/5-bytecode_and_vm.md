# Bytecode y Virtual Machine

En nuestra versión anterior, teníamos un **tree walking interpreter**, pero vimos que esto hacía que no fuera veloz. Para mejorar nuestro lenguaje, implementaremos un **bytecode compiler**.

Nuestro nuevo diseño se basará en:

1. Convertir código de Lox en bytecode.
2. Ejecutar ese bytecode en una VM.

Como _IR_ (Representación Intermedia), utilizamos _bytecode_, con la intención de desacoplar el almacenamiento del código de su ejecución. Esto no sucedía en nuestro tree walking interpreter ya que el AST no emitía código: se armaba y se ejecutaba en la misma corrida. Además, el AST es un árbol de objetos dispersos en memoria que se recorre de forma recursiva, mientras que el bytecode es una secuencia lineal de bytes, lo cual lo hace más rápido de recorrer. El bytecode nos permite tener esta flexibilidad sin tener que llegar a código máquina, que requeriría adaptarse a cada arquitectura posible.

Estos bytecodes formarán un _instruction set ficticio_, un conjunto secuencial de instrucciones. Nos alcanza con un único conjunto, ya que funcionará donde sea que funcione nuestra VM, sin importar la arquitectura.

## Chunks

Una instrucción tiene:

1. **OpCode** → primer byte, indica qué operación es.
2. **Operandos** → la información que necesita esa operación. Puede ser uno o más bytes, dependiendo de la instrucción.

Los valores que usa el programa se guardan aparte, en un **pool de constantes**, y algunas instrucciones (como `OP_CONSTANT`) usan su operando como índice de ese pool. Por lo que finalmente un **chunk** se compone de este pool y del set de instrucciones.

# Virtual Machine

La VM simula el comportamiento de una CPU, y será la encargada de ejecutar nuestro código. Se compone de:

- **instruction pointer**: indica la siguiente instrucción a ejecutar. Como es una secuencia lineal, simplemente va avanzando, salvo cuando una instrucción de salto lo mueve. Con las funciones cada llamada tiene el suyo, y la VM usa el de la llamada que se está ejecutando.
- **stack**: guarda los valores con los que trabaja la VM. Inicialmente eran los valores temporales que producen las instrucciones en runtime, como el resultado de `1 + 2`. Esos valores son usados por instrucciones que trabajan siempre sobre el tope del stack (por ejemplo `OP_ADD` o `OP_PRINT`). Luego, incorporamos las variables locales, cuyos valores también viven en el stack. A diferencia de los anteriores, no se acceden desde el tope, sino por su índice (con `OP_GET_LOCAL` y `OP_SET_LOCAL`). El compilador conoce ese índice porque lleva la cuenta de en qué posición queda cada variable, y lo emite como operando de esas instrucciones. Con las funciones, en el stack también viven la función que se está ejecutando (desde closures, el closure que la envuelve) y sus argumentos. Y si una función interna usa una variable local de la función que la envuelve, esa variable sigue viviendo en el stack mientras la función de afuera está activa, y recién cuando sale de él se mueve al heap (ver [Closures](#closures)).

Al igual que el intérprete, la VM es la que ejecuta el programa. La diferencia está en cómo lo recorre: el intérprete recorría el árbol de forma recursiva, mientras que la VM ejecuta las instrucciones en un único ciclo, ya que son un array lineal.

Si la vm detecta un error de runtime, como puede ser que se recibe un tipo incorrecto, entonces corta y devuelve el error con la linea. Pero como la vm mira con `peek()` antes de sacar del stack, si falla quedaría basura en el stack. Por este motivo, cuando el `run()` de la VM falla, invocamos a `resetStack()` usando el `defer` de Go. Además de vaciar el stack, ahora también saca los frames y vacía la lista de upvalues abiertos, para que en el REPL la próxima línea no reutilice upvalues que apuntaban a posiciones del stack que ya no existen.

### Instrucciones de statements y variables globales

Con los statements aparecen dos instrucciones. `OP_PRINT` saca el valor del tope del stack y lo imprime, usando el mismo `Stringify` que el tree-walk así los dos motores imprimen igual, por ejemplo un `nil` se muestra como `nil`. `OP_POP` simplemente saca el valor del tope y lo tira.

Para las globales hay tres:

1. `OP_DEFINE_GLOBAL` lee el nombre entre las constantes del chunk, saca el valor del stack y lo guarda en la tabla de globales, que es un `map[string]any`, y si la variable ya existía la pisa.
2. `OP_GET_GLOBAL` busca el nombre en esa tabla y apila el valor, y si no está levanta un error de runtime. Es importante ver que, para saber si está en la tabla usamos `valor, bool`, porque el `map` de Go devuelve `nil` tanto si la variable _no existe_ como si _existe y vale `nil`_, por ejemplo con `var a;`. Queremos esos casos porque uno es un error de que no existe la variable, y el otro no, porque Lox tolera que una variable sea `nil`.
3. `OP_SET_GLOBAL` primero se fija si la variable existe, y si no existe levanta el error y no la crea, porque en Lox una variable no se puede declarar asignándola, tiene que estar declarada antes con `var`. Si existe le guarda el valor del tope, pero sin sacarlo del stack, porque la asignación es una expresión y su valor se puede seguir usando, como por ejemplo en `a = b = 5`.

### Variables globales vs locales

La diferencia principal es que las globales son _late bound_: se resuelven en tiempo de ejecución, buscándolas por nombre en una tabla de globales, lo que permite que el código compile aunque se use una variable antes de ser declarada, lo único que importa es que exista cuando se ejecute.

En cambio las locales se resuelven en tiempo de compilación: el compilador ya sabe en qué posición del stack va a estar cada una, y emite esa posición como operando. Esto las hace más rápidas: la VM no tiene que buscar la variable, recibe directamente su posición. Y acceder a una posición del stack es más rápido que buscar en un `map`, que tiene que calcular el hash del nombre.

Otra diferencia es lo que sucede al declarar de nuevo una variable. Una global se puede redeclarar y solo se pisa el valor, pero una local en el mismo scope no, eso es un error de compilación. En cambio, con una local de un scope distinto sí se puede, y ahí lo que pasa es el _shadowing_.

### Control de flujo

Con el control de flujo, la VM deja de ejecutar siempre la instrucción siguiente. Las instrucciones de salto (`OP_JUMP_*`) mueven el instruction pointer hacia **adelante**, salteando instrucciones que no se tienen que ejecutar. En los condicionales, como `if-else`:

- `OP_JUMP_IF_FALSE`: si la condición es falsa, saltea el cuerpo del `if` y la VM sigue en el `else` (o en lo que siga, si no hay `else`).
- `OP_JUMP`: al terminar el cuerpo del `if`, saltea el `else` para que no se ejecuten las dos cosas.

La cantidad de bytes a saltear viene en el operando, que ocupa dos bytes, así que la VM los lee con `readShort()` y los vuelve a juntar en un solo número. Para la VM esto no implica mucha más lógica que saber a dónde mover el instruction pointer, lo más interesante ocurre en el [compilador](6-compiler.md).

El caso de los ciclos es parecido, pero al revés: el `OP_LOOP` **resta** un offset en vez de sumarlo, para volver al inicio del ciclo.

### Funciones y call frames

Con el agregado de funciones, la VM deja de tener un único chunk e ip para todo el programa, porque cada función tiene su propio chunk. Cuando se llama a una función, la VM tiene que pausar el chunk que estaba ejecutando, guardar por qué instrucción iba (su `ip`) y empezar a ejecutar el chunk de la función llamada. Cuando esa función termina, tiene que poder volver al chunk anterior y seguir desde la instrucción que había guardado.

Un detalle acá es que no alcanza con guardar el `ip` y la ventana en la propia función, porque no soportaría la recursión. Por ejemplo, en un Fibonacci recursivo, la función se llama a sí misma, y la llamada de adentro usaría el mismo lugar que la de afuera, pisando por dónde iba. Al volver, la de afuera ya no sabría desde dónde seguir. Entonces la solución es guardar uno por **llamada**, no por función. Ahí es que llega el `call frame` que el libro propone, que en sí para nosotros representa una llamada a una función y guarda tres cosas:

- la función que se está ejecutando (al agregar closures, el closure que se está ejecutando)
- su propio ip para saber por dónde va la llamada a la función
- el lugar del stack donde empiezan sus variables

Esto último es importante, y el libro propone guardar las variables locales de todas las funciones en el propio stack de la VM, y esto funciona porque las llamadas siguen el mismo orden que una pila (**LIFO**). Si `f` llama a `g`, las locales de `g` quedan arriba de las de `f`, y como `g` siempre termina antes que `f`, sus locales se sacan primero. Así cada llamada tiene su _ventana en el stack_, arriba de la anterior, sin pisarse.

Esto implica que la posición de una variable local no es siempre la misma: depende en que momento se llama a una función, y el estado del stack en ese momento. Para eso, el compilador emite la posición de las variables locales como _relativas_ a la función, mientras que la VM tiene por cada `CallFrame`, su `slots`, que indica donde arranca la llamada de esa función en el stack de la VM. Al sumar ambos, se obtiene la posición real de la variable local en el stack.

Acá surge una diferencia con el libro: trabaja `slots` como un puntero a la posición del stack, mientras que en glox lo tenemos como un índice, y esto lo hicimos así porque el stack en Go es un slice dinámico y al crecer se podría mover a otro lugar de la memoria y el puntero quedaría viejo, por eso fuimos con el índice que siempre nos sirve.

Otro detalle es que los frames se guardan en su propia pila `frames`, donde el del tope es el de la llamada que se está ejecutando. En el libro se limita a un máximo de 64, es decir que como mucho puede haber ese número de llamadas activas al mismo tiempo. Si bien el libro lo hace con un array y nosotros en Go con un slice que le reservamos capacidad al crear la VM, decidimos pegarnos a ese mismo valor y eso trae un detalle particular: como el script ya ocupa un frame, como mucho puede haber 63 llamadas anidadas y si se pasa de ahí la VM corta con el error de `stack overflow`. Esto es una limitación conocida e importante porque el tree-walk de la primera parte no lo tiene, entonces una recursión profunda de más de 63 llamadas antes funcionaba pero en esta implementación ya corta con el error de `stack overflow` (se rompe la regla de oro donde todas las implementaciones de Lox deberían tener el mismo resultado).

Todo lo que usaba el ip y el chunk de la vm antes, ahora lo lee desde el frame actual. Esto lo hicimos en la VM con el `currentFrame()` que nos devuelve el de arriba. Además, `run` guarda el frame actual en una variable y la actualiza cada vez que entra o sale de una llamada, así no tiene que buscarlo en cada instrucción.

#### Llamar a una funcion

Cuando la VM ejecuta un `OP_CALL` ya sabe cuántos argumentos son, y con eso sabe que la función está esa cantidad de lugares antes del tope. Lo primero que hace es fijarse qué es lo que le piden llamar, y si no es un closure (antes funciones) ni una función nativa, corta con error. Acá hay una diferencia con el libro, que revisa el tipo del objeto, mientras que en Go como los valores de la VM son `any`, nos alcanza con un switch por tipo.

Si es un closure, chequea dos cosas antes de abrir el frame.
   - Primero, que la aridad de la función sea la esperada, y si no devuelve un error que avisa cuántos argumentos esperaba y cuántos recibió. 
   - Después, que no se haya llegado al máximo de frames y si es así falla con `stack overflow`. 
   
Los mensajes del tipo que no se puede llamar y de la aridad son los mismos que ya tenía el tree-walk pero el de `stack overflow` es nuevo porque el tree-walk no tiene ese límite, y es la limitación que ya mencionamos con los call frames.

Si pasa todos los chequeos se abre el frame, y ahí está lo importante. El `slots` es la posición de la función (desde closures, del closure) dentro del stack, y siguiendo están los argumentos. Como los parámetros son las primeras variables locales de la función, cada uno coincide con el argumento que le toca, así que no hace falta copiarlos porque el frame arranca desde ahí directamente.

#### Stack trace

Cuando hay un error en tiempo de ejecución, además del mensaje con la línea, mostramos el **stack trace**, que es el camino de llamadas que llevó hasta el error, desde la función donde falló hasta el script, con la línea de cada una. Por ejemplo `[line 3] in inner(), [line 4] in outer() y [line 6] in script.` Esto lo trae el libro y el tree-walk no lo tiene, ahí solo se ve la línea del error. Como en el libro, mostramos primero la función más interna.

Acá hay una diferencia con el libro, que lo imprime en el momento en que ocurre el error. En glox la VM devuelve los errores y no los imprime, entonces el trace se arma en el `defer` de `Run`, que se ejecuta cuando falla cualquier instrucción, y se lo agrega al error antes de reiniciar el stack porque después de eso ya no quedarían los frames para recorrer.

#### Return

Cuando una función termina, la VM ejecuta un `OP_RETURN`, que lo primero que hace es sacar el valor que se devuelve, así no se pierde, y después descarta todo lo que era de esa llamada: la función (desde closures, el closure), sus argumentos y sus variables locales. Para eso corta el stack en el `slots` del frame y saca el frame de la pila. Antes de cortar, cierra los upvalues que apuntan a esa parte del stack (ver [Closures](#closures)). Si era el último frame, el programa terminó, y si no, vuelve a poner el valor devuelto en el stack, que es donde lo espera el que lo llamó, y sigue con el frame anterior desde donde había quedado.

Para que todas las funciones devuelvan algo, el compilador agrega al final de cada una un `OP_NIL` y un `OP_RETURN` así que si no hay un return explícito se devuelve nil. Esto lo hace el libro también y es por lo que el programa termina cuando vuelve el frame del script.

Acá hay un par de diferencias con el libro:
   - El libro, cuando termina el script, hace un `pop` para sacar la función del script antes de salir. Nosotros no lo necesitamos porque cortar el stack en el `slots` del script ya lo deja todo vacío.
   - Chequeo de que el stack quede vacío cuando termina el script: si no se cumple avisa con un error. No es del libro, pero sí nos basamos en ploxb y lo hacemos solo en el script porque en las funciones el corte ya lo limpia todo.

Como `OP_RETURN` descarta todo lo de la llamada, al terminar el cuerpo de una función no hace falta sacar uno por uno los parámetros y las locales con `OP_POP`.

#### Funciones nativas

Las funciones nativas no están escritas en Lox sino en Go que es el lenguaje con el que hicimos la implementación, y sirven para que un programa pueda hacer cosas que Lox solo no puede, como saber el tiempo. La primera que trae el libro es `clock` y nosotros hicimos la misma, que aprovechamos para usarla para el benchmark mas adelante como ya hicimos en el tree-walk.

Una nativa no tiene bytecode ni chunk solo la cantidad de argumentos que espera y una función de Go que recibe los argumentos y devuelve un valor. Por eso cuando la VM llama a una nativa no abre un frame, como con una función de Lox, sino que directamente chequea la aridad y ejecuta la función de Go con los argumentos que están en el stack y deja el resultado en lugar de la función y sus argumentos.

Una diferencia con el libro que tenemos es que para agregar una nativa, el libro usa `defineNative` que además apila y desapila valores para protegerlos de su garbage collector. En glox eso no hace falta, así que tenemos una tabla con las nativas que la VM instala como globales al crearse y agregar otra es sumar una entrada a esa tabla. Otra diferencia es el `clock`: el de clox mide el tiempo de CPU y el nuestro devuelve el tiempo del reloj igual que el del tree-walk, así los dos caminos miden lo mismo y los benchmarks se pueden comparar.

Además, incorporamos uno de los desafíos de este capítulo: agregarles aridad a las funciones nativas. El libro no la chequea para que llamarlas sea más rápido, así que en clox una llamada como `clock(1)` simplemente ignora el argumento. En glox cada nativa guarda la cantidad de argumentos que espera, y si no coincide, la VM devuelve el mismo error que con una función de Lox. Así se comportan igual que en el tree-walk, donde las nativas también chequean la aridad. Algo positivo de la versión del libro es que permite nativas con una cantidad variable de argumentos, como una `max` que acepte `max(1, 2)` y también `max(1, 2, 3)`. Con nuestro chequeo eso no se puede, pero no le delega a cada función nativa chequear los argumentos que le llegan.

### Closures

Antes las funciones se cargaban con `OP_CONSTANT` como cualquier otra constante. Ahora la VM las carga con `OP_CLOSURE`, que crea un closure con un puntero a la `Function` y sus propios **upvalues**, y pushea el closure a su stack.

Como todas las funciones se emiten con `OP_CLOSURE`, en el stack y en las variables nunca hay una `Function`, siempre un closure que la envuelve. Eso implica que el `callFrame` ahora guarda el closure que se está ejecutando, no la `Function`. Esto es importante, porque mientras se ejecuta una función la VM necesita llegar a sus upvalues, pero la `Function` solo sabe cuántos upvalues hay y de dónde sacarlos, y eso es igual para todos sus closures. En cambio, la variable concreta cambia en cada uno: dos closures de la misma función pueden haberse creado en llamadas distintas a la función de afuera, y cada uno captura la variable de su llamada. Por eso las referencias están en el closure.

El script es el único que no se crea con `OP_CLOSURE`, porque no está declarado dentro de ninguna función, así que la VM lo envuelve en un closure cuando lo recibe en `Run()`, antes de ejecutarlo.

**Upvalues**

Los upvalues son punteros a variables que declara otra función, una más externa. Los necesitamos para que cuando la interna se ejecute pueda usar esas variables, aunque no estén entre sus locales ni entre las globales. Que sea un puntero nos permite que, al modificar la variable, el cambio quede aplicado para todas las llamadas y también se vea desde los niveles de anidamiento más afuera.

En los call frames habíamos mencionado que, a diferencia del libro, usábamos un índice y no un puntero para marcar dónde empieza la ventana del stack de una función, porque si el stack crecía, Go podía moverlo y el puntero quedaba viejo. Acá sí usamos un puntero al stack, y es seguro porque el stack de la VM se crea con una capacidad fija y mientras no se pase de ese tamaño, Go no lo mueve.

Cuando una variable sale del stack (terminó un bloque o una función retornó), el puntero quedaría apuntando a un lugar inválido. Como contamos en el compilador, justo antes de que esto pase movemos la variable al heap, así luego puede seguir usándose. Para eso, se suma al `upvalue` un campo llamado `closed`. Al sacarla del stack, copiamos el valor directamente a este campo y el puntero pasa a apuntar ahí. A esto se le dice **cerrar** el upvalue: mientras apunta al stack está abierto, y cuando ya tiene su propia copia, está cerrado.

**Instrucciones**

- **`OP_CLOSURE`**: crea el closure y lee los pares `<isLocal> <index>` que dejó el compilador. Como en ese momento está corriendo la función de afuera, si la variable es local de ella está en su ventana, en `slots + index`. Ahí la VM busca en `openUpvalues` si ya hay un upvalue para esa posición: si lo hay, lo reutiliza, y si no crea uno y lo inserta en orden. Si la variable no es local, la función de afuera ya la tiene como upvalue, así que reutiliza el que está en su closure, en la posición `index`.
- **`OP_GET_UPVALUE` / `OP_SET_UPVALUE`**: buscan el upvalue en el closure que se está ejecutando y leen o escriben a través del puntero.
- **`OP_CLOSE_UPVALUE`**: el compilador la emite al terminar un bloque para cada variable capturada. Cierra el upvalue de la variable que está en el tope del stack y la saca del stack.
- **`OP_RETURN`**: antes de recortar el stack, cierra todos los upvalues que apuntan a la ventana de la función que termina.

**Upvalues abiertos**

`openUpvalues` es una lista enlazada de upvalues que guarda la VM con todos los upvalues que todavía apuntan al stack. Lo usan dos de las instrucciones de arriba:

- **Al cerrar** (`OP_CLOSE_UPVALUE` y `OP_RETURN`), la VM sabe qué posición del stack se está liberando, pero necesita los upvalues que apuntan ahí. Cada closure tiene sus upvalues, y estos sus punteros al stack, pero desde las posiciones a liberar del stack no tenemos otra forma de saber cuales son los upvalues que lo estaban referenciando. Para eso se crea esta lista: se recorre desde el inicio, cada upvalue que apunta a la posición que se libera o más arriba se cierra y se saca de la lista, y en cuanto aparece uno que apunta más abajo, corta.
- **Al crear un closure** (`OP_CLOSURE`), sirve para que dos closures que capturan la misma variable compartan el mismo upvalue. Si cada uno tuviera uno propio, al moverse al heap cada uno copiaría el valor en su propio `closed`, y desde ahí serían dos variables distintas: si uno la modifica, el otro no se entera (cada puntero apunta a un lugar distinto). Por eso, al crear un upvalue, primero se fija en esta lista si ya existe uno para esa posición.

La lista enlazada está ordenada por posición en el stack, empezando por la más alta (porque el stack siempre se va liberando desde su tope). El libro elige esta estructura porque para mantener el orden a veces hay que insertar en el medio, y al cerrar se sacan los primeros: en un array habría que mover los demás elementos, y en una lista enlazada alcanza con cambiar punteros. Como cada elemento es un upvalue, a estos les agregamos un campo `next` con el siguiente elemento. Por último, el libro compara punteros para mantener la lista ordenada. Como en Go no se pueden comparar punteros con `>`, guardamos también la posición del stack (`stackIndex`) y comparamos eso.

### Clases e instancias

**Clases**

Creamos una clase que, por ahora, solo guarda su nombre. Además le agregamos el método `String()`, para que `Stringify` pueda imprimirla. El libro dice que el nombre no es estrictamente necesario para el programador, pero sirve para mostrarlo en runtime, por ejemplo en el **stack trace**.

Como viene pasando en el resto de la implementación de esta segunda parte, no necesitamos el header de `Obj` ni tampoco el tipo `OBJ_CLASS` ni las macros que propone el libro: al usar Go alcanza con `any` y luego pregunta con `v.(*class)` para saber si un valor es una clase. Tampoco tenemos que liberar la memoria de las clases, porque lo hace el garbage collector de Go.

**Instancias**

Una instancia guarda a qué clase pertenece, para después poder encontrar sus métodos, y sus propios campos.

En Lox los campos no se declaran, sino que se agregan en runtime y se buscan por nombre. Para guardarlos, el libro usa su tabla de hash y nosotros un `map[string]any`, que es la misma idea, igual que con las variables globales. Un detalle es que, como los campos se agregan en runtime, cada acceso implica hashear el nombre y buscarlo. Esto es más lento que en un lenguaje estático, donde el compilador ya conoce los campos de cada clase y su tamaño, y acceder a uno es solo sumar un desplazamiento fijo.

Además, un detalle importante que menciona el libro es que para el que programa, una instancia de algo como un `Auto` y otra de un `Perro` son de tipos distintos, y tiene sentido que así sea. Pero para la VM son el mismo tipo, `instance`: lo único que cambia es a qué clase apuntan.

Para crear la instancia se llama a la clase como si fuera una función, por ejemplo `Auto()`, así que no hace falta una instrucción nueva: se usa `OP_CALL`. Lo que sí se agrega es un caso en `callValue` para detectarlo. Al crear la instancia, la vm la pone en el lugar del stack donde estaba la clase y listo: no hace falta abrir un frame porque no hay código Lox para ejecutar.

#### Get y set para los campos

Para cuando la VM ejecuta `OP_GET_PROPERTY`, la instancia ya está en el tope del stack, porque lo que estaba a la izquierda del punto se compiló antes. En `OP_SET_PROPERTY` cambia: abajo queda la instancia y arriba el valor que se quiere asignar, así que para encontrar la instancia hay que mirar un lugar más abajo del tope.

En el **get**, una vez que ya ve que es una instancia, busca el campo en el map de la misma y deja el valor en el lugar del stack que ocupaba la instancia, básicamente saca la instancia y mete el valor en el stack. En el **set** guarda el valor en el campo, saca el valor y la instancia y vuelve a meter el valor. Con esto se logra que la asignación sea una expresión entonces cuando haces, por ejemplo, `print auto.velocidad = 5;` se termina devolviendo el 5 que se usa después para imprimir.

En Lox los campos no se declaran: si se asigna uno que no existía, se crea, y si ya existía, se pisa. Con eso ya tenemos crear y modificar. En cambio, leer un campo que no existe es un error.

Antes de leer o asignar, la VM se fija que el valor sea una instancia, y si no lo es, da error. Esto incluye a las clases, así que algo como `A.x = 1` también es un error. Los errores de runtime de esta parte son entonces tres: leer un campo que no existe, leer de algo que no es una instancia y asignar a algo que no es una instancia. Un detalle propio de Go es que un campo puede existir y valer `nil`, y el map de Go también devuelve `nil` cuando no encuentra la clave. Para distinguir los dos casos usamos la forma `valor, ok := fields[nombre]`, donde `ok` indica si el campo existe.

Sobre limitaciones conocidas para esta parte, tenemos que no se puede borrar un campo porque si le asignas `nil`, dijimos que era válido y por lo tanto seguiróa existiendo. Tampoco tenemos forma de preguntar si existe el campo sin que nos de error.

#### Métodos

Ahora cada clase puede tener sus métodos y van guardados en un map de closures, van en cada clase y no en cada instancia porque son los mismos para las instancias a diferencia de los campos. Como ha pasado a lo largo de esta segunda parte usamos un map y eso nos ayuda a no encargarnos explícitamente de la liberación de la memoria.

Básicamente para definir un método la vm necesita tres cosas de ella, el nombre que viene entre las constantes del chunk, el closure del método que llega en el tope del stack y la clase que viene un nivel más abajo. Acá no tenemos chequeos en la vm de que lo que está en el stack realmente es una clase y un closure, igual que lo que hace el libro que lo toma así sabiendo el orden. Con eso, el OP_METHOD guarda el closure en el map de la clase con ese nombre y saca solo el closure del stack, así la clase queda arriba para el próximo método. 

Un detalle sobre métodos acá es que si llegan dos con el mismo nombre no es un error, sino que el último pisa al primero igual que en el libro. Y como una curiosidad propia es que ahora sí usamos el klass como en el libro recién ahora, pero porque ya teníamos un tipo class y choca con el mismo nombre de la variable y no podemos, no como en el libro que habla de que class era palabra reservada de C++.

### Diferencias con el libro

1. **Instruction Pointer**: en el libro se utiliza un puntero al siguiente byte del chunk, que avanza con aritmética de punteros. Como Go no la permite, utilizamos un índice, que el libro descarta por ser más lento.
2. **Memoria**: en el libro, va a más bajo nivel y hace manejo de memoria, por estar utilizando C. De nuevo, en Go esto no es necesario, así que aprovechamos las utilidades que nos trae. Un ejemplo de esto es que el libro lleva una lista de objetos creados para luego liberar la memoria al finalizar, mientras que en glox al apoyarnos en Go no hace falta, ya que contamos con un garbage collector que libera cuando ya no se usa (ya no quedan referencias).
3. **Scanner**: a diferencia del libro, no vamos a reimplementar nuestro scanner para que vaya escaneando de a un token a la vez, sino que vamos a reutilizar el que ya teníamos.
4. **Values**: como C no tiene una forma de guardar un valor de cualquier tipo, el libro define un struct con el truco de _Tagged Unions_. En Go, nos alcanza con que los elementos del stack sean de tipo `any`, y chequeamos su tipo con _type assertions_ (por ejemplo, `v.(float64)`). En el caso de los strings, en el libro, se habla de que son de tamaños variables por lo que tienen que guardarlos en el heap y agregar un `VAL_OBJ` que tiene un puntero a ellos. En nuestro glox eso no hace falta, por el mismo motivo de que usamos `any` que puede guardar un string de Go.
5. **Strings**: en Lox dos strings con el mismo contenido son iguales. En el libro tienen la situación que si comparás dos punteros te dirían que son distintos, por eso en el libro se compara la longitud y los bytes. En nuestro glox nos apoyamos en Go, donde el `==` entre dos valores del `any` ya compara el contenido, así que `OP_EQUAL` simplemente usa el `IsEqual` que es la misma función que viene desde el tree walk.

   Además en el libro el `OP_ADD` se fija si ambos operandos son strings para llamar a `concatenate()` que reserva un array nuevo, concatena los strings y crea el objeto, pero en nuestro glox nos apoyamos en que en Go al usar el `+` ya devuelve un string nuevo sin modificar los originales.
6. **Hash Table**: en el libro se implementa a mano una tabla clave-valor rápida que se usa en varios lugares, por ejemplo para guardar variables. Se construye con todos los detalles que necesita clox, cuidando la memoria y la caché. En glox usamos el `map` de Go, como `map[string]any`, que ya resuelve el hash, las colisiones, el crecimiento y el borrado. Además, el libro hace _interning_ de strings para poder compararlos por puntero, pero en Go no es necesario porque el `==` ya compara el contenido. Un detalle a tener en cuenta es que en el libro el hash de cada string se cachea y en Go no, así que es algo para prestar atención en los benchmarks.
7. **OP_DEFINE_GLOBAL**: en el libro primero se guarda el valor en la tabla mirándolo con `peek` y recién después se saca del stack, porque si la tabla tiene que crecer justo en ese momento su garbage collector podría perder el valor. En Go ese problema no existe, así que nosotros lo sacamos del stack directamente y lo guardamos en el `map`.
8. **OP_SET_GLOBAL**: en el libro se guarda el valor en la tabla y si la variable no existía se borra y se levanta el error. Nosotros hacemos la pregunta antes, nos fijamos si la variable existe en el `map`, y si no existe levantamos el error sin haber guardado nada, así nunca hay que borrar. El resultado es el mismo, una variable que no se declaró no se crea por asignarla.
9. **Leer el nombre de una variable**: en el libro el `READ_STRING` no chequea que la constante sea un string, y en glox el `readString()` sí lo chequea y devuelve un error si no lo es.
10. **Orden de los upvalues abiertos**: el libro ordena la lista comparando punteros, y en Go no se puede, así que cada upvalue guarda su posición en el stack (`stackIndex`) y comparamos eso.
11. Aridad de las nativas: el libro no la chequea pero nosotros sí, con el mismo error que las funciones de Lox.

### Limitaciones conocidas

Si bien estas limitaciones las fuimos contando en cada sección donde aparecen, para dejarlo más claro las enumeramos acá. Son límites que el **bytecode compiler** tiene y el **tree-walk interpreter** no y como son comportamientos que dependen de la implementación y no del lenguaje, los documentamos y un programa en Lox no debería depender de ellos.

- **Llamadas anidadas**: puede haber como mucho 63 llamadas anidadas, porque la VM tiene un máximo de 64 frames y el script ocupa uno. Si se pasa corta con un error de `stack overflow` mientras que el tree-walk aguanta una recursión mucho más profunda. Si bien esto es un numero que sigue el libro, decidimos tomar ese mismo.
- **Variables locales**: una función puede tener como mucho 255, contando sus parámetros, porque el operando de una local ocupa un byte y el slot 0 queda reservado para la función. El tree-walk no tiene este límite.
- **Constantes**: cada función puede tener como mucho 256 constantes, porque su posición también ocupa un byte. Como cada función tiene su propio chunk el límite es por función, y en el tree-walk no existe. Con las clases, los nombres de las clases y de los campos también ocupan lugar entre esas constantes.
- **Tamaño de un salto**: el cuerpo de un `if`, de un `while` o `for` no puede pasar de los 65.535 bytes porque el salto se guarda en dos bytes, y esto también limita a `break` y `continue`. En el tree-walk no hay límite.
- **Upvalues**: una función puede capturar como mucho 256 variables de funciones que la envuelven, porque la posición de cada upvalue se emite en un byte. El tree-walk no tiene este límite.

## Tests

Para la VM implementamos una serie de tests unitarios, validando que la ejecución de cada instrucción tenga su efecto esperado. Para correrlos:

```bash
go test ./internal/virtual_machine
```

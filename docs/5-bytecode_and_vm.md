## Bytecode y Virtual Machine

En nuestra versión anterior, teníamos un **tree walking interpreter**, pero vimos que esto hacía que no fuera veloz. Para mejorar nuestro lenguaje, implementaremos un **bytecode compiler**.

Nuestro nuevo diseño se basará en:

1. Convertir código de Lox en bytecode.
2. Ejecutar ese bytecode en una VM.

Como _IR_ (Representación Intermedia), utilizamos _bytecode_, con la intención de desacoplar el almacenamiento del código de su ejecución. Esto no sucedía en nuestro tree walking interpreter ya que el AST no emitía código: se armaba y se ejecutaba en la misma corrida. Además, el AST es un árbol de objetos dispersos en memoria que se recorre de forma recursiva, mientras que el bytecode es una secuencia lineal de bytes, lo cual lo hace más rápido de recorrer. El bytecode nos permite tener esta flexibilidad sin tener que llegar a código máquina, que requeriría adaptarse a cada arquitectura posible.

Estos bytecodes formarán un _instruction set ficticio_, un conjunto secuencial de instrucciones. Nos alcanza con un único conjunto, ya que funcionará donde sea que funcione nuestra VM, sin importar la arquitectura.

### Chunks

Una instrucción tiene:

1. **OpCode** → primer byte, indica qué operación es.
2. **Operandos** → la información que necesita esa operación. Puede ser uno o más bytes, dependiendo de la instrucción.

Los valores que usa el programa se guardan aparte, en un **pool de constantes**, y algunas instrucciones (como `OP_CONSTANT`) usan su operando como índice de ese pool. Por lo que finalmente un **chunk** se compone de este pool y del set de instrucciones.

## Virtual Machine

La VM simula el comportamiento de una CPU, y será la encargada de ejecutar nuestro código. Se compone de:

- **instruction pointer**: indica la siguiente instrucción a ejecutar. Como es una secuencia lineal, simplemente va avanzando, salvo cuando una instrucción de salto lo mueve. Con las funciones cada llamada tiene el suyo, y la VM usa el de la llamada que se está ejecutando.
- **stack**: guarda los valores con los que trabaja la VM. Inicialmente eran los valores temporales que producen las instrucciones en runtime, como el resultado de `1 + 2`. Esos valores son usados por instrucciones que trabajan siempre sobre el tope del stack (por ejemplo `OP_ADD` o `OP_PRINT`). Luego, incorporamos las variables locales, cuyos valores también viven en el stack. A diferencia de los anteriores, no se acceden desde el tope, sino por su índice (con `OP_GET_LOCAL` y `OP_SET_LOCAL`). El compilador conoce ese índice porque lleva la cuenta de en qué posición queda cada variable, y lo emite como operando de esas instrucciones. Con las funciones, en el stack también viven la función que se está ejecutando y sus argumentos.

Al igual que el intérprete, la VM es la que ejecuta el programa. La diferencia está en cómo lo recorre: el intérprete recorría el árbol de forma recursiva, mientras que la VM ejecuta las instrucciones en un único ciclo, ya que son un array lineal.

Si la vm detecta un error de runtime, como puede ser que se recibe un tipo incorrecto, entonces corta y devuelve el error con la linea. Pero como la vm mira con `peek()` antes de sacar del stack, si falla quedaría basura en el stack. Entonces es por este motivo que cuando el `run()` de la vm falla, invocamos a `resetStack()`, que ahora también saca los frames, utilizando el defer propio de Go.

#### Instrucciones de statements y variables globales

Con los statements aparecen dos instrucciones. `OP_PRINT` saca el valor del tope del stack y lo imprime, usando el mismo `Stringify` que el tree-walk así los dos motores imprimen igual, por ejemplo un `nil` se muestra como `nil`. `OP_POP` simplemente saca el valor del tope y lo tira.

Para las globales hay tres:

1. `OP_DEFINE_GLOBAL` lee el nombre entre las constantes del chunk, saca el valor del stack y lo guarda en la tabla de globales, que es un `map[string]any`, y si la variable ya existía la pisa.
2. `OP_GET_GLOBAL` busca el nombre en esa tabla y apila el valor, y si no está levanta un error de runtime. Es importante ver que, para saber si está en la tabla usamos `valor, bool`, porque el `map` de Go devuelve `nil` tanto si la variable _no existe_ como si _existe y vale `nil`_, por ejemplo con `var a;`. Queremos esos casos porque uno es un error de que no existe la variable, y el otro no, porque Lox tolera que una variable sea `nil`.
3. `OP_SET_GLOBAL` primero se fija si la variable existe, y si no existe levanta el error y no la crea, porque en Lox una variable no se puede declarar asignándola, tiene que estar declarada antes con `var`. Si existe le guarda el valor del tope, pero sin sacarlo del stack, porque la asignación es una expresión y su valor se puede seguir usando, como por ejemplo en `a = b = 5`.

#### Variables globales vs locales

La diferencia principal es que las globales son _late bound_: se resuelven en tiempo de ejecución, buscándolas por nombre en una tabla de globales, lo que permite que el código compile aunque se use una variable antes de ser declarada, lo único que importa es que exista cuando se ejecute.

En cambio las locales se resuelven en tiempo de compilación: el compilador ya sabe en qué posición del stack va a estar cada una, y emite esa posición como operando. Esto las hace más rápidas: la VM no tiene que buscar la variable, recibe directamente su posición. Y acceder a una posición del stack es más rápido que buscar en un `map`, que tiene que calcular el hash del nombre.

Otra diferencia es lo que sucede al declarar de nuevo una variable. Una global se puede redeclarar y solo se pisa el valor, pero una local en el mismo scope no, eso es un error de compilación. En cambio, con una local de un scope distinto sí se puede, y ahí lo que pasa es el _shadowing_.

#### Control de flujo

Con el control de flujo, la VM deja de ejecutar siempre la instrucción siguiente. Las instrucciones de salto (`OP_JUMP_*`) mueven el instruction pointer hacia **adelante**, salteando instrucciones que no se tienen que ejecutar. En los condicionales, como `if-else`:

- `OP_JUMP_IF_FALSE`: si la condición es falsa, saltea el cuerpo del `if` y la VM sigue en el `else` (o en lo que siga, si no hay `else`).
- `OP_JUMP`: al terminar el cuerpo del `if`, saltea el `else` para que no se ejecuten las dos cosas.

La cantidad de bytes a saltear viene en el operando, que ocupa dos bytes, así que la VM los lee con `readShort()` y los vuelve a juntar en un solo número. Para la VM esto no implica mucha más lógica que saber a dónde mover el instruction pointer, lo más interesante ocurre en el [compilador](6-compiler.md).

El caso de los ciclos es parecido, pero al revés: el `OP_LOOP` **resta** un offset en vez de sumarlo, para volver al inicio del ciclo.

#### Funciones y call frames

Con el agregado de funciones la VM deja de tener un único chunk e ip para todo porque cuando el programa llama a una función, la VM tiene que dejar lo que estaba haciendo en ese momento y pasarse al código de esa función, y cuando termina la función tiene que volver a donde estaba, y esto es lo que motiva por qué tiene que acordarse de donde estaba para al volver seguir ejecutando desde ahí. Un detalle acá es que no alcanza con uno por función porque eso no soportaría la recursión, en un ejemplo de Fibonacci recursivo se llama a sí mismo y se pisaria entre las propias llamadas. Entonces la solución ahí es guardar uno por llamada, no por función. Ahí es que llega el call frame que el libro propone, que en sí para nosotros representa una llamada a una función y guarda tres cosas, primero la función que se está ejecutando, su propio ip para saber por dónde va la llamada a la función y el lugar del stack donde empiezan sus variables.

El tema de ver donde empiezan sus variables lo necesitamos y el libro propone dejarlo en el propio stack de la VM y la razón de esto es porque las variables locales funcionan como una pila, es decir son LIFO, la última que llega es la primera que sale, y esto tiene sentido porque ponele que tenés las de la función f y luego dentro la de g, primero se consumen las de g, y luego las de f. Esto trae la situación de que la posición de una variable local no es siempre la misma porque depende de qué momento se llamó a la función y cómo estaba el stack ahí. Entonces el compilador emite la posición de cada variable local relativa a la función, mientras que la VM tiene en slots donde arranca la llamada en el stack entonces al sumar esas dos sabe donde está la variable. Acá surge una diferencia con el libro que lo trabaja como un puntero a esa posición, mientras que en glox lo tenemos como un índice, y esto lo hicimos así porque el stack en Go es un slice y al crecer se podría mover a otro lugar de la memoria y queda viejo, por eso fuimos con el índice que siempre nos sirve.

Un detalle es que los frames se guardan en su propia pila y el de arriba de todo es el de la llamada que se está ejecutando. En el libro se limita a un máximo de 64, es decir que como mucho puede haber ese número de llamadas activas al mismo tiempo. Si bien el libro lo hace con un array y nosotros en Go con un slice que le reservamos capacidad al crear la VM, nosotros decidimos pegarnos a ese mismo valor y eso trae un detalle particular, que es que como el script ya ocupa un frame, por lo que como mucho puede haber 63 llmadas anidadas y si se pasa de ahí la VM corta con el error de stack overflow. Esto es una limitación conocida e importante porque el tree-walk de la primera parte no lo tiene, entonces una recursión profunda de más de 63 llamadas antes funcionaba pero en esta implementación ya corta con el error de stack overflow.

Todo lo que usaba el ip y el chunk de la vm antes, ahora lo lee desde el frame actual. Esto lo hicimos en la VM con el currentFrame() que nos devuelve el de arriba. Y un detalle sobre el run es que se guarda el frame actual de una varible y lo va actualizando cuando entra o sale de una llamada así ya lo tiene actualizado al querer usarlo.

#### Llamar a una funcion

Cuando la VM ejecuta un OP_CALL ya sabe cuántos argumentos son, y con eso sabe que la función está esa cantidad de lugares antes del tope. Lo primero que hace es fijarse qué es lo que le piden llamar, y si no es una función corta con error. Acá hay una diferencia con el libro, que revisa el tipo del objeto, mientras que en Go como los valores de la VM son any, nos alcanza con un switch por tipo.

Si es una función, chequea dos cosas antes de abrir el frame. Primero que la aridad sea la esperada, y si no devuelve un error que avisa cuántos argumentos esperaba y cuántos recibió. Y después que no se haya llegado al máximo de frames y si es así falla con stack overflow. Los mensajes del tipo que no se puede llamar y de la aridad son los mismos que ya tenía el tree-walk pero el de stack overflow es nuevo porque el tree-walk no tiene ese límite, y es la limitación que ya contamos con los call frames.

Si pasa todos los chequeos se abre el frame, y ahí está lo importante. El slots apunta a la función dentro del stack, y siguiendo están los argumentos. Como los parámetros son las primeras variables locales de la función, cada uno coincide con el argumento que le toca, así que no hace falta copiarlos porque el frame arranca desde ahí directamente.

#### Stack trace

Cuando hay un error en tiempo de ejecución además del mensaje con la línea mostramos el stack trace, que es el camino de llamadas que llevó hasta el error, desde la función donde falló hasta el script con la línea de cada una. Por ejemplo [line 3] in inner(), [line 4] in outer() y [line 6] in script. Esto lo trae el libro y el tree-walk no lo tiene, ahí solo se ve la línea del error. Como en el libro, mostramos primero la función más interna.

Acá hay una diferencia con el libro, que lo imprime en el momento en que ocurre el error. En glox la VM devuelve los errores y no los imprime, entonces el trace se arma en el defer de Run, que se ejecuta cuando falla cualquier instrucción, y se lo agrega al error antes de reiniciar el stack porque después de eso ya no quedarían los frames para recorrer.

#### Return

Cuando una función termina la VM ejecuta un OP_RETURN, que lo primero que hace es sacar el valor que se devuelve así no se pierde y después descarta todo lo que era de esa llamada, que es la función, sus argumentos y sus variables locales. Para eso corta el stack en el slots del frame y saca el frame de la pila. Si era el último el programa terminó y si no vuelve a poner el valor devuelto en el stack, que es donde lo espera el que lo llamó, y sigue con el frame anterior desde donde había quedado.

Para que todas las funciones devuelvan algo el compilador agrega al final de cada una un OP_NIL y un OP_RETURN así que si no hay un return explícito se devuelve nil. Esto lo hace el libro también y es por lo que el programa termina cuando vuelve el frame del script.

Acá hay un par de diferencias con el libro como cuando termina el script que hace un pop para sacar la función del script antes de salir. Nosotros no lo necesitamos porque cortar el stack en el slots del script ya lo deja todo vacío. Y el chequeo de que el stack quede vacío cuando termina el script, que si no se cumple avisa con un error, no es del libro pero sí nos basamos en ploxb y lo hacemos solo en el script porque en las funciones el corte ya lo limpia todo.

Como OP_RETURN descarta todo lo de la llamada, al terminar el cuerpo de una función no hace falta sacar uno por uno los parámetros y las locales con OP_POP.

#### Funciones nativas

Las funciones nativas no están escritas en Lox sino en Go que es el lenguaje con el que hicimos la implementación, y sirven para que un programa pueda hacer cosas que Lox solo no puede, como saber el tiempo. La primera que trae el libro es clock y nosotros hicimos la misma, que aprovechamos para usarla para el benchmark mas adelante como ya hicimos en el tree-walk.

Una nativa no tiene bytecode ni chunk solo la cantidad de argumentos que espera y una función de Go que recibe los argumentos y devuelve un valor. Por eso cuando la VM llama a una nativa no abre un frame como con una función de Lox si no que directamente chequea la aridad y ejecuta la función de Go con los argumentos que están en el stack y deja el resultado en lugar de la función y sus argumentos.

Una difernecia con el libro que tenemos es qye para agregar una nativa el libro usa defineNative que además apila y desapila valores para protegerlos de su garbage collector. En glox eso no hace falta así que tenemos una tabla con las nativas que la VM instala como globales al crearse y agregar otra es sumar una entrada a esa tabla. Otra diferencia es el clock porque el de clox mide el tiempo de CPU y el nuestro devuelve el tiempo del reloj igual que el del tree-walk así los dos caminos miden lo mismo y los benchmarks se pueden comparar. 

#### Clases e instancias

Creamos una clase que por ahora solo tiene como dato su nombre. Además de eso el método string para que el Stringify pueda llamarlo e imprimirlo. Si bien el libro habla de que no es estrictamente necesario para el programador que lo use, está bueno para mostrar el nombre en runtime como por ejemplo para el stack trace. 

Como viene pasando en el resto de la implementación de esta segunda parte no necesitamos el header de Obj ni tampoco el tipo ni las macros que propone el libro, al usar Go alcanza con any y luego pregunta con v.(*class). Siguiendo un poco con esta diferencia con el libro tampoco nos tenemos que hacer cargo de liberar la memoria para las clases.

Llega el momento de las instancias que acá tienen que guardar a qué clase pertenecen para después poder encontrar sus métodos, y además sus propios campos. Sobre estos últimos es que tenemos una diferencia con el libro, porque también como ha pasado a lo largo de esta segunda parte más que nada se refiere a cómo se guarda esto en memoria. En Lox los campos no es que se declaran sino que se agregan en el runtime y se terminan buscando por nombre, ahí es donde el libro usa su tabla de hash y nosotros volvemos a un map[string]any que en el fondo es la misma idea pero difiere porque nosotros usamos Go. Un detalle acá es que como los campos se agregan en el runtime hay que hashear y buscar, frente a ya tenerlo fijo si el compilador los tiene declarados con cada tamaño, que es más rápido porque eso es solo sumar y desplazarse, eso sería una diferencia frente a un lenguaje estático.

Un detalle que también habla el libro y nos parece importante comentarlo es esto, que para el que programa una instancia de algo como un Auto y otra como un Perro son de tipos distintos y tiene toda la lógica que así sea, pero para nosotros en la vm es el mismo instance, digamos que lo único que cambia es a qué clase apuntan.

Para crear la instancia se llama a la clase como si fuera una función, como por ejemplo Auto(), por lo que no hace falta una instrucción nueva ya que se usa el OP_CALL y lo que sí cambia es el caso en el callValue para detectarlo. Al crear la instancia la vm la pone en el lugar del stack donde estaba la clase y listo, no es necesario meterse con el frame porque no hay código Lox ahí para ejecutar.

#### Get y set para los campos

Para cuando la VM ejecuta el OP_GET_PROPERTY la instancia ya está en el tope del stack porque lo que estaba a la izquierda se compiló antes, que eso cambia para el OP_SET_PROPERTY porque abajo queda la instancia y arriba el valor que se le quiere asignar y eso trae la diferencia entre las dos que para la última necesitamos mirar un lugar más abajo del tope para encontrar la instancia.

En el get, una vez que ya ve que es una instancia busca el campo en el map de la misma y deja el valor en el lugar del stack que ocupaba la instancia, básicamente saca la instancia y mete el valor en el stack. En el set guarda el valor en el campo, saca el valor y la instancia y vuelve a meter el valor. Con esto se logra que la asignación sea una expresión entonces cuando haces, por ejemplo, print auto.velocidad = 5; se termina devolviendo el 5 que se usa después para imprimir.

Un detalle de los campos para Lox es que no se declaran entonces si se asigna uno que no existía simplemente lo crea y guarda el valor, pero si existía lo pisa, con eso ya tenemos el crear y el modificar, pero si a la hora de querer leerlo no existe ahí sí levantamos un error.

Los errores en runtime que tenemos acá son el que recién mencionamos de que no existe el campo en la instancia, el de tratar de leer el valor de algo que no es una instancia y también tratar de asignar a algo que no es una instancia, básicamente antes de leer o asignar la vm se fija que sea una instancia y si no lo es da error, y esto incluye a las clases, así que algo como A.x = 1 también es un error. Pero con un detalle y es que un campo puede ser nil y existir claramente, y como usamos un map que devuelve nil tanto si es el valor como si no lo encontró ahí tenemos que estar atentos para diferenciar los dos casos, esto es particular de Go y de que usamos un map.

Sobre limitaciones conocidas para esta parte tenemos que no se puede borrar un campo porque si le asignas nil dijimos que es válido y sigue existiendo. Tampoco tenemos forma de preguntar si existe el campo sin que nos de error. 

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

### Limitaciones conocidas

Si bien estas limitaciones las fuimos contando en cada sección donde aparecen, para dejarlo más claro las enumeramos acá. Son límites que la VM tiene y el tree-walk no y como son comportamientos que dependen de la implementación y no del lenguaje, los documentamos y un programa en Lox no debería depender de ellos.

- **Llamadas anidadas**: puede haber como mucho 63 llamadas anidadas, porque la VM tiene un máximo de 64 frames y el script ocupa uno. Si se pasa corta con un error de stack overflow mientras que el tree-walk aguanta una recursión mucho más profunda. Si bien esto es un numero que sigue el libro decidimos tomar ese mismo.
- **Variables locales**: una función puede tener como mucho 255, contando sus parámetros, porque el operando de una local ocupa un byte y el slot 0 queda reservado para la función. El tree-walk no tiene este límite.
- **Constantes**: cada función puede tener como mucho 256 constantes, porque su posición también ocupa un byte. Como cada función tiene su propio chunk el límite es por función, y en el tree-walk no existe. Con las clases, los nombres de las clases y de los campos también ocupan lugar entre esas constantes.
- **Tamaño de un salto**: el cuerpo de un if o de un while o for no puede pasar de los 65.535 bytes porque el salto se guarda en dos bytes, y esto también limita a break y continue. En el tree-walk no hay límite.

### Tests

Para la VM implementamos una serie de tests unitarios, validando que la ejecución de cada instrucción tenga su efecto esperado. Para correrlos:

```bash
go test ./internal/virtual_machine
```

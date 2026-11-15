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

- **instruction pointer**: indica la siguiente instrucción a ejecutar. Como es una secuencia lineal, simplemente va avanzando, salvo cuando una instrucción de salto lo mueve.
- **stack**: guarda los valores con los que trabaja la VM. Inicialmente eran los valores temporales que producen las instrucciones en runtime, como el resultado de `1 + 2`. Esos valores son usados por instrucciones que trabajan siempre sobre el tope del stack (por ejemplo `OP_ADD` o `OP_PRINT`). Luego, incorporamos las variables locales, cuyos valores también viven en el stack. A diferencia de los anteriores, no se acceden desde el tope, sino por su índice (con `OP_GET_LOCAL` y `OP_SET_LOCAL`). El compilador conoce ese índice porque lleva la cuenta de en qué posición queda cada variable, y lo emite como operando de esas instrucciones.

Al igual que el intérprete, la VM es la que ejecuta el programa. La diferencia está en cómo lo recorre: el intérprete recorría el árbol de forma recursiva, mientras que la VM ejecuta las instrucciones en un único ciclo, ya que son un array lineal.

Si la vm detecta un error de runtime, como puede ser que se recibe un tipo incorrecto, entonces corta y devuelve el error con la linea. Pero como la vm mira con `peek()` antes de sacar del stack, si falla quedaría basura en el stack. Entonces es por este motivo que cuando el `run()` de la vm falla, invocamos a `resetStack()` utilizando el defer propio de Go.

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

### Tests

Para la VM implementamos una serie de tests unitarios, validando que la ejecución de cada instrucción tenga su efecto esperado. Para correrlos:

```bash
go test ./internal/virtual_machine
```

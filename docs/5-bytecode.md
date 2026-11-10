## Bytecode

En nuestra versión anterior, teníamos un **tree walking interpreter**, pero vimos que esto hacía que no fuera veloz. Para mejorar nuestro lenguaje, implementaremos un **bytecode compiler**.

Nuestro nuevo diseño se basará en:

1. Convertir código de Lox en bytecode.
2. Ejecutar ese bytecode en una VM.

Como IR (Representación Intermedia), utilizamos _bytecode_, con la intención de desacoplar el almacenamiento del código de su ejecución. Esto no sucedía en nuestro tree walking interpreter ya que el AST no emitía código: se armaba y se ejecutaba en la misma corrida. Además, el AST es un árbol de objetos dispersos en memoria que se recorre de forma recursiva, mientras que el bytecode es una secuencia lineal de bytes, lo cual lo hace más rápido de recorrer. El bytecode nos permite tener esta flexibilidad sin tener que llegar a código máquina, que requeriría adaptarse a cada arquitectura posible.

Estos bytecodes formarán un _instruction set ficticio_, un conjunto secuencial de instrucciones. Nos alcanza con un único conjunto, ya que funcionará donde sea que funcione nuestra VM, sin importar la arquitectura.

### Chunks

Una instrucción tiene:

1. **OpCode** → primer byte, indica qué operación es.
2. **Operandos** → la información que necesita esa operación. Puede ser uno o más bytes, dependiendo de la instrucción.

Los valores que usa el programa se guardan aparte, en un **pool de constantes**, y algunas instrucciones (como `OP_CONSTANT`) usan su operando como índice de ese pool. Por lo que finalmente un **chunk** se compone de este pool y del set de instrucciones.

### Virtual Machine

La VM simula el comportamiento de una CPU, y será la encargada de ejecutar nuestro código. Se compone de:

- **instruction pointer**: indica la siguiente instrucción a ejecutar. Como es una secuencia lineal, simplemente va avanzando.
- **stack**: algunas instrucciones generan valores, que pueden ser utilizados por las siguientes instrucciones.

Igual que como hacía el intérprete, la VM recorre en un ciclo las instrucciones y las va ejecutando.

Si la vm detecta un error de runtime como puede ser que se recibe un tipo incorrecto, entonces corta y devuelve el error con la linea. Pero como la vm mira con peek antes de sacar del stack, si falla queda basura en el stack. Entonces es por este motivo que cuando el run de la vm falla, invocamos a resetStack() utilizando el defer propio de Go.

### Diferencias con el libro

1. **Instruction Pointer**: en el libro se utiliza un puntero al siguiente byte del chunk, que avanza con aritmética de punteros. Como Go no la permite, utilizamos un índice, que el libro descarta por ser más lento.
2. **Memoria**: en el libro, va a más bajo nivel y hace manejo de memoria, por estar utilizando C. De nuevo, en Go esto no es necesario, así que aprovechamos las utilidades que nos trae. Un ejemplo de esto es que el libro lleva una lsita de objetos creados para luego liberar la memoria al finalizar, mientras que en glox al apoyarnos en Go no hace falta, ya que contamos con un garbage collector que libera cuando ya no se usa (ya no quedan referencias).
3. **Scanner**: a diferencia del libro, no vamos a reimplementar nuestro scanner para que vaya escaneando de a un token a la vez, sino que vamos a reutilizar el que ya teníamos.
4. **Values**: como C no tiene una forma de guardar un valor de cualquier tipo, el libro define un struct con el truco de _Tagged Unions_. En Go, nos alcanza con que los elementos del stack sean de tipo `any`, y chequeamos su tipo con _type assertions_ (por ejemplo, `v.(float64)`). En el caso de los strings, en el libro, se habla de que son de tamaños variables por lo que tienen que guardarlos en el heap y agregar un VAL_OBJ que tiene un puntero a ellos. En nuestro glox eso no hace falta, por el mismo motivo de que usamos `any` que puede guardar un string de Go.
5. **Strings**: en Lox dos strings con el mismo contenido son iguales. En el libro tienen la situación que si comparás dos punteros te dirían que son distintos, por eso en el libro se compara la longitud y los bytes. En nuestro glox nos apoyamos en Go, donde el == entre dos valores del any ya compara el contenido, así que OP_EQUAL simplemente usa el IsEqual que es la misma función que viene desde el tree walk.
Además en el libro el OP_ADD se fija si ambos operandos son strings para llamar a concatenate() que reserva un array nuevo, concatena los strings y crea el objeto, pero en nuestro glox nos apoyamos en que en Go al usar el + ya devuelve un string nuevo sin modificar los originales. 

### Tests

Para la VM implementamos una serie de tests unitarios, validando que la ejecución de cada instrucción tenga su efecto esperado. Para correrlos:

```bash
go test ./internal/virtual_machine
```

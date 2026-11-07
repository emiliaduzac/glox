## Bytecode

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

### Diferencias con el libro

1. Instruction Pointer: en el libro se utiliza un puntero al siguiente byte del chunk, que avanza con aritmética de punteros. Como Go no la permite, utilizamos un índice, que el libro descarta por ser más lento.
2. Memoria: en el libro, va a más bajo nivel y hace manejo de memoria, por estar utilizando C. De nuevo, en Go esto no es necesario, así que aprovechamos las utilidades que nos trae.

### Tests

Para la VM implementamos una serie de tests unitarios, validando que la ejecución de cada instrucción tenga su efecto esperado. Para correrlos:

```bash
go test ./internal/virtual_machine
```

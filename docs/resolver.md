## Resolver

Antes de meternos dentro del resolver, debemos comprender algunos conceptos, para entender por qué fue necesario implementarlo.

Hasta el momento, veníamos manejando **scopes** de la misma manera que los **entornos** como si fueran lo mismo, pero no son:

- El **entorno** es una herramienta **dinámica y mutable**.
- El **scope** es un concepto **estático**, que **no es necesariamente un bloque entero**, es simplemente **un conjunto de bindings**.

Su principal diferencia es que uno es dinámico y el otro estático, por lo que tratarlos de la misma forma generaba graves errores (dependíamos de como se encontraba el entorno al momento de, por ejemplo, llamar funciones y no de su declaración). Pero entonces, quién tiene la responsabilidad de manejar cada uno?

1. Scanner ⇒ tiene la responsabilidad del **léxico**, el vocabulario

2. Parser ⇒ tiene la responsabilidad de la **sintaxis**, la estructura

3. Interpreter ⇒ ejecuta el programa (runtime, dinámico)

Como vemos, los scopes parecieran no entrar en ningún caso. Acá entra un paso intermedio entre el parser y el interpreter: el **resolver**. Este se encarga de dar pasadas sobre el AST _antes_ de ejecutarlo, que buscan entender el significado de lo que el programador quiso hacer: **análisis semántico**. Es clave entender que estas pasadas no ejecutan nada: no hay efectos secundarios ni control de flujo. Simplemente se encarga del scope estático, resolviendo variables locales.

A diferencia del **environment**, este no busca resolver valores sino distancias. La idea es saber a cuantos saltos está definida una variable, para que luego el intérprete pueda ir a buscarla directamente, sin tener que volver a calcularlo, y evitando errores dinámicos.

Entonces el flujo queda: escanear → parsear → resolver → interpretar.

### Como funciona el resolver

El resolver almacena una **pila de scopes** de la siguiente forma `[]map[string]bool{}`. Recorre las sentencias del árbol y maneja cada caso individualmente:

- **Bloques**: Al iniciar un bloque, almacena un nuevo scope a la pila de scopes. Luego resuelve el bloque (que utilizará el scope recién definido), y al finalizar elimina el scope de la pila.

- **Variables**: al declarar una variable, primero se la marca en el scope actual con false. Esto indica que la variable existe pero aún no fue definida, no puede usarse. Luego de resolver la inicialización, se la marca como `true`, indicando que puede usarse. Esto permite manejar casos como `var a = a;`, donde no sería correcto usar la misma variable para definirla.

- **Expresiones y print statements**: simplemente resuelve la expresión.

- **Condicionales (`if`)**: resuelve la condición (una expresión) y las dos ramas (statements): la que se ejecuta si se cumple la condición, y la del `else`, si la hay. El `if` no abre su propio scope: si alguna de las ramas es un bloque, abrirá su propio scope cuando se resuelva. Si no lo es, quiere decir que no se definirá ninguna variable nueva.

- **Loops**: resuelve la condición y el cuerpo. Tampoco abre su propio scope, por la misma razón que el `if`.

Es importante entender que el resolver mantiene los scopes casi como una copia de los entornos que mantiene el intérprete: se define un scope nuevo exactamente en los mismos lugares donde el intérprete definiría un entorno nuevo. La diferencia es que estos scopes no guardan valores, y por más que guarden booleanos (`declarado`/`definido`), el objetivo principal de todo el resolver es calcular las distancias.

### Como resuelve expresiones

El resolver recorre cada expresión.

La mayoría de los casos, como por ejemplo `Binary`, `Unary`, `Grouping` y `Logical`, se recorre recursivamente ya que están compuestos de más expresiones: resuelven sus operandos.

Los casos más importantes son:

- **Variables**: como mencionamos antes, al intentar usar una variable primero verifica que no se la esté usando dentro de su propia inicialización (el caso `var a = a;`). Si no es ese caso, calcula la distancia recorriendo la pila de scopes y al encontrarla, se la manda al intérprete, que la guarda en su propio `map`. De esta forma, cuando el intérprete tiene que resolver una expresión, primero verifica si sabe cuántos saltos tiene que dar entre los entornos para hallarla e ir directo. Si no sabe, la busca dinámicamente, sabiendo que se trata de una variable global y no local.

- **Asignaciones**: las asignaciones están compuestas de la forma: `<variable> = <expresión>`. Como esa expresión del lado derecho puede contener más variables (por ejemplo `x = y + 1;`), primero resuelve esa expresión. Luego calcula la distancia de la variable a asignar, de la misma forma que al usar variables, y se la manda al intérprete.

### Tests

Implementamos tests unitarios que validan tanto los casos correctos (distancias calculadas para variables en el mismo scope, en scopes más externos, con shadowing, etc.) como los casos de error (usar una variable dentro de su propia inicialización, tipos de statement/expression no soportados).

```bash
go test ./internal/resolver
```

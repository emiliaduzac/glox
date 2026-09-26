## Interpreter

El interpreter tiene la responsabilidad de recorrer el **AST** generado por el parser. La idea es que evalúe las expresiones de los statements de forma recursiva. Según el libro se espera que lo haga en un **recorrido post-order**, es decir primero los hijos para luego operar sobre los resultados de estos.

### Mapeo de tipos en Glox

| Tipo en Lox | Representación en Go |
| --- | --- |
| NUMBER | `float64` |
| STRING | `string` |
| BOOL | `bool` |
| NIL | `nil` |

Particularmente en la parte de nil, nos encontramos que por cómo trabaja Go terminaba saliendo `<nil>`, entonces ahí fue que implementamos la función de `Stringify`.

### Tipos

Para saber el tipo usamos un `switch e := expr.(type)`, a diferencia de, por ejemplo el libro, donde proponen su implementación en Java utilizando el patrón Visitor. 

Acá es que también debemos manejar "a mano" el caso `default` y estar atentos para un error explícito de tipo no soportado.

### Operadores

1. **Truthiness**: Es importante respetar Truthiness para Lox. En este caso se define que solo `false` y `nil` son falsos, y que todo lo demás es verdadero, comportamiento que quedó declarado en la función de `isTruthy`.

2. En glox manejamos la no coerción implícita de tipo. Es decir, `3 == "3"` es `false`, termina comparando tipos distintos de datos y es manejado simplemente con un `false` y no con un error. 

3. **División por cero**: Al dividir por cero, se retorna `+Inf` o `-Inf` según el caso. Y en el caso de `0 / 0`, se retorna `NaN`.

4. Sobre el operador módulo (`%`): Al manejarlo Go solo lo aplica a enteros, por lo que usamos `math.Mod` para preservar los flotantes y no caer en panics. 

### Errores en tiempo de ejecución

Si por ejemplo llega un `true < false`, no se crashea con `panic`, sino que se retorna un error con el número de línea informativo. 

### Statements

Una vez incorporados los statements, el interpreter ya no solo se encarga de evaluar expresiones, sino también ejecutar sentencias.

El interpreter pasa a recibir una lista de sentencias y ejecutarlas una por una. Esto puede o no producir un efecto secundario, pero a diferencia de una expresión, **no produce un valor**, por eso `Execute` devuelve solo un `error`. 
Por ejemplo en los `print statements` se evalúa la expresión y se imprime su resultado, mientras que en las `var statements` se almacenará el resultado en memoria, etc.

#### Entorno y estado

Al agregar estado este tiene que vivir en algún lugar, pero en vez de que sea directamente en el intérprete, bien modular y siguiendo el libro, esto directamente queda en el `Environment`. Esto lo dejamos ahí como un `map[string]any`.

Es clave entender las decisiones semánticas sobre `define`, `get` y `assign`.

Empezando por `define`, es que NUNCA mira para arriba al padre, sino que es local siempre. Otra cosa importante sobre el `define` es que si ya estaba en el scope actual la variable, y se lo llama de nuevo la pisa con el nuevo valor.

Sobre el `get`, primero se devuelve si es que existe en local, pero si no existe se consulta recursivamente al padre. En caso de llegar a la raíz, ahí sí corresponde el error de que la variable no ha sido declarada, básicamente esto es, en ningún momento declaraste esa variable ni en este entorno local, ni más arriba.

Y sobre el `assign` sigue una lógica similar de que si existe en local, actualiza el valor local. Si no existe pero tiene padre lo delega automáticamente, y si no existe en ninguna parte entonces ahí sí tira error. Esto porque en Lox no hay variables implícitas, y entonces si no fue declarada la variable antes del `assign` falla en runtime. Y por el contrario si todo sale bien, la asignación devuelve un valor porque es una expresión en Lox. Además hay un caso más borde que si solo hacés `var a;` se guarda con `nil`. 

Lo bastante interesante de estos métodos y sus combinaciones, es tener en claro lo del shadowing, es decir, al usar `define` en local sobre una variable que ya existe afuera, en ese entorno lo "tapa". Entonces si en local luego consultás por eso mismo, te responde el local, pero nunca modificaste lo de afuera.

Otro caso muy interesante es el de mutación de las variables externas. Es decir, si en local hacés un `assign` a una variable que en local no existe, entonces se va para el padre y se la modifica al padre (si es que él tiene esa). Entonces, terminado el entorno local, pasa que modificaste lo de afuera. 

#### Ejecución de bloques

Para ejecutar un bloque implementamos en el método encargado la idea de guardarse una referencia al entorno anterior como una variable, activás el entorno del hijo y ahí ejecutás cada una de las sentencias del bloque. 

Un detalle propio de implementar Glox, es que en el libro al usar Java se usa un `try` y `finally`, pero en Go no existe, por lo que la forma más idiomática de Go fue usar un `defer`. Esto ayuda a que si una sentencia del bloque falla al ejecutarse, se corta y Go termina ejecutando sí o sí `in.environment = previous` que es el scope del padre, y no se queda atrapado en el del hijo. 

Y un detalle interesante de nuestra implementación hasta el momento es que no tenemos un método para destruir los scopes. Sino que como los punteros van del hijo al padre, cuando se termina el bloque y se cambia el puntero al padre, nadie más referencia al hijo, y eso es lo que le permite al garbage collector de Go detectar que ya nadie lo referencia, por lo tanto es inalcanzable y lo libera de manera automática. Por el momento es algo que está así, analizaremos a futuro si así lo dejamos. 

### Tests

Quisimos dejar plasmado comportamiento en test de las diferentes expresiones y sentencias, y los casos de cada una de ellas. Hay archivos de test separados para que sea más fácil leerlos, y también para seguir la línea de implementación. 
Para correr todos los tests del interpreter:

```bash
go test ./internal/interpreter
```

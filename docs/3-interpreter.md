## Interpreter

El interpreter tiene la responsabilidad de recorrer el **AST** generado por el parser. La idea es que evalúe las expresiones de los statements de forma recursiva. Según el libro se espera que lo haga en un **recorrido post-order**, es decir primero los hijos para luego operar sobre los resultados de estos.

### Mapeo de tipos en Glox

| Tipo en Lox | Representación en Go |
| --- | --- |
| NUMBER | `float64` |
| STRING | `string` |
| BOOL | `bool` |
| NIL | `nil` |
| FUNCTION / CALLABLE | `Callable` (interfaz) |

Particularmente en la parte de nil, nos encontramos que por cómo trabaja Go terminaba saliendo `<nil>`, entonces ahí fue que implementamos la función de `Stringify`.

### Tipos

Para saber el tipo usamos un `switch e := expr.(type)`, a diferencia de, por ejemplo el libro, donde proponen su implementación en Java utilizando el patrón Visitor. 

Acá es que también debemos manejar "a mano" el caso `default` y estar atentos para un error explícito de tipo no soportado.

### Operadores

1. **Truthiness**: Es importante respetar Truthiness para Lox. En este caso se define que solo `false` y `nil` son falsos, y que todo lo demás es verdadero, comportamiento que quedó declarado en la función de `isTruthy`.

2. En glox manejamos la no coerción implícita de tipo. Es decir, `3 == "3"` es `false`, termina comparando tipos distintos de datos y es manejado simplemente con un `false` y no con un error. 

3. **División por cero**: Al dividir por cero, se retorna `+Inf` o `-Inf` según el caso. Y en el caso de `0 / 0`, se retorna `NaN`.

4. Sobre el operador módulo (`%`): Al manejarlo Go solo lo aplica a enteros, por lo que usamos `math.Mod` para preservar los flotantes y no caer en panics. 

5. **Operadores lógicos**: a diferencia de otros lenguajes, evaluar estos operadores no se resuelve a un booleano, sino que devuelven el valor del operando que "decidió" el resultado. Es decir: el de la izquierda si hubo cortocircuito (explicados en [parser.md](parser.md)), el de la derecha si no. 

### Funciones y llamadas

Cuando el intérprete va a evaluar una llamada a una función, primero que nada evalúa la expresión de lo que quiere llamar. Después, evalúa todos los argumentos de izquierda a derecha. Se fija si lo que se quiere llamar realmente es una función, por ejemplo si intentás hacer `"hola"()`, se corta con un error de runtime. Un dato es que el error lo dejamos como en el libro, que habla de clases aunque en este momento no las soportamos, sabemos que vienen y ya quedó el error tal cual como está en el libro en el capítulo 10.
Otra cosa que se fija es la arity, es decir la cantidad de argumentos que le pasaste tiene que ser la misma que la esperada, y si no coinciden tira error avisando cuántas obtuvo y las esperadas.

#### Función clock()

Siguiendo un poco el libro fue que también implementamos la función nativa de `clock()` que tiene aridad 0 y devuelve el tiempo. Podríamos usarla para un propio benchmark.

### Errores en tiempo de ejecución

Si por ejemplo llega un `true < false`, no se crashea con `panic`, sino que se retorna un error con el número de línea informativo. 

### Statements

Una vez incorporados los statements, el interpreter ya no solo se encarga de evaluar expresiones, sino también ejecutar sentencias.

El interpreter pasa a recibir una lista de sentencias y ejecutarlas una por una. Esto puede o no producir un efecto secundario, pero a diferencia de una expresión, **no produce un valor**, por eso `Execute` devuelve solo un `error`. 
Por ejemplo en los `print statements` se evalúa la expresión y se imprime su resultado, mientras que en las `var statements` se almacenará el resultado en memoria, etc.

#### Entorno y estado

Al agregar estado este tiene que vivir en algún lugar, pero en vez de que sea directamente en el intérprete, bien modular y siguiendo el libro, esto directamente queda en el `Environment`. Esto lo dejamos ahí como un `map[string]any`.

Es clave entender las decisiones semánticas sobre `define`, `get` y `assign`.

- **`Define`**: nunca mira para arriba del entorno padre, sino que es local siempre. Además, si la variable vuelve a definirse en el scope actual, se pisa el valor.

- **`Get`**: primero se devuelve si es que existe en local, pero si no existe se consulta recursivamente al padre. En caso de llegar a la raíz, ahí sí corresponde el error de que la variable no ha sido declarada: ni en este entorno local, ni más arriba.

- **`Assign`**: si existe en local, actualiza el valor local. Si no existe pero tiene padre lo delega automáticamente, y si no existe en ninguna parte entonces ahí sí tira error. Esto porque en Lox no hay variables implícitas, y entonces si no fue declarada la variable antes del `assign` falla en runtime. Y por el contrario si todo sale bien, la asignación devuelve un valor porque es una expresión en Lox. Además hay un caso más borde donde si solo se hace `var a;` se guarda con `nil`.

Lo interesante de estos métodos y sus combinaciones, es tener en claro lo del _shadowing_, es decir: al usar `define` en local sobre una variable que ya existe afuera, en ese entorno lo "tapa". Entonces si en local luego consultás por eso mismo, te responde el valor local, pero nunca se modifica lo de afuera.

Otro caso muy interesante es el de _mutación de las variables externas_. Es decir, si en local hacés un `assign` a una variable que en local no existe, entonces se va para el padre y se la modifica al padre (si es que él tiene esa). Entonces, terminado el entorno local, pasa que modificaste lo de afuera.

Más adelante, al incorporar el resolver, esta búsqueda cambia: el intérprete ya no recorre la cadena de entornos para encontrar una variable. Si el resolver calculó una distancia para ella, el intérprete va directo al entorno correcto. Si no hay distancia, es una variable global y se busca en `globals`. Detallado en [Resolver](4-resolver.md).

#### Ejecución de bloques

Para ejecutar un bloque implementamos el siguiente método: la idea es primero guardarse una referencia al entorno actual como una variable (pasa a ser el `previous`), después se activa el entorno del hijo (que pasa a ser el actual) y ahí se ejecuta cada una de las sentencias del bloque. 

Un detalle propio de implementar Glox, es que en el libro al usar Java se usa un `try` y `finally`, pero en Go no existe, por lo que la forma más idiomática de Go fue usar un `defer`. Esto ayuda a que si una sentencia del bloque falla al ejecutarse, se corta y Go termina ejecutando sí o sí `in.environment = previous` que es el scope del padre, y no se queda atrapado en el del hijo. 

Y un detalle interesante de nuestra implementación es que, hasta el momento, no tenemos un método para destruir los scopes explícitamente. Acá hay que distinguir dos punteros distintos: por un lado el `enclosing`: puntero de un entorno hijo hacia su padre, que se define una sola vez al crearlo y nunca cambia después. Por otro lado el `environment`: el puntero de "cuál entorno está activo ahora" en el intérprete: al entrar al bloque apunta al hijo, y al salir (por el `defer`) vuelve a apuntar al padre. Una vez que eso pasa, y la variable local `env` del método sale de scope, nadie en el programa referencia más al entorno del hijo, y eso es lo que le permite al garbage collector de Go detectarlo y liberarlo solo. Al momento de agregar funciones y closures nos acordamos de esto, pero como una función intenta quedarse con ese entorno, si la función sigue "viva", la referencia también y el garbage collector no lo borra, así que las variables de afuera de la función que necesita de contexto siguen, hasta que finalmente se quedan sin referencia. 

Al agregar los ciclos sucede que, cuando el cuerpo es un bloque, se crea un nuevo `environment` en cada iteración. Por lo tanto, las variables definidas dentro del cuerpo se re-declaran en cada vuelta (cada una en su entorno, que vive exactamente una sola iteración), si se quisiera utilizar una variable más a modo acumuladora, debe declararse fuera del ciclo (como sucede en todos los lenguajes). 

#### Declaración de funciones y returns

El intérprete al encontrarse con una función no la ejecuta en ese momento sino que la empaqueta con el entorno actual y el nombre que le pusimos. 

Para los returns acá usamos una estrategia diferente a lo que se habla en el libro y más propia de Java, en nuestro Glox implementamos la idea de un `ReturnValue`. Es decir, tenemos el caso de que dentro de un bucle capaz queremos salir con un return, entonces aprovechamos que si llegamos al return se devuelve el `ReturnValue` para arriba, y como los chequeos son del tipo `err != nil`, este error nuestro sigue su camino cortando hasta llegar al call, que se fija qué es lo que volvió, si un error de `ReturnValue` que en sí no es un error real sino el valor retornado, el cual saca el valor y sale bien, o si realmente era un error del runtime. 

### Tests

Quisimos dejar plasmado comportamiento en test de las diferentes expresiones y sentencias, y los casos de cada una de ellas. Hay archivos de test separados para que sea más fácil leerlos, y también para seguir la línea de implementación. 
Para correr todos los tests del interpreter:

```bash
go test ./internal/interpreter
```

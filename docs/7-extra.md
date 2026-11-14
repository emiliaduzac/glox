## Extra

Como extra, decidimos incorporar **break** y **continue** a los ciclos. Como vimos en clase, cuando queremos agregar funcionalidad nueva, tenemos que ir en el orden correspondiente: _scanner → compiler → vm_.

### Continue

`continue` corta la iteración actual de un ciclo y pasa directamente a la siguiente: lo que queda del cuerpo no se ejecuta. En un `while`, vuelve a evaluar la condición directo, pero en un `for` tiene que ejecutar el incremento y recién después volver a la condición.

En primer lugar, debemos agregarlo como palabra reservada entre los [tokens](../internal/token/token.go). Con eso alcanza para que el scanner lo identifique, ya que, al leer una palabra, la busca en el map de palabras reservadas: si la encuentra, le asigna ese tipo de token, y si no, la considera un identificador.

Luego, seguimos por el compiler. Sabemos que un `continue` debe ser un **statement**, ya que no produce ningún valor sino que genera un efecto: saltear una iteración de un ciclo. Con eso ya sabemos muchas cosas:

1. lo tenemos que incorporar en `statement()`
2. probablemente emita un `OP_LOOP`, ya que debe volver atrás en el código
3. al no ser una expresión no tenemos que modificar la tabla del _Pratt Parser_
4. finaliza con un `;`
5. tiene que dejar el stack tal cual lo encuentra. Con esto vemos que si en el cuerpo de un ciclo se definen variables locales, el `continue` antes de volver al inicio, debe asegurarse de sacar esas variables del stack para que la nueva iteración empiece limpia. Cuando no hay `continue`, esto se limpia solo al llegar al `endScope()`.
6. sí o sí deben estar dentro de un ciclo, en caso contrario sería un error. Pero también debe soportar estar anidado dentro de más ciclos o ifs.

Primero agregamos el `CONTINUE` al `statement()`, pero ahí nos surgió la duda de como seguir: un nuevo statement no tiene conocimiento de los demás statements, cómo sabe si efectivamente está en un ciclo? y cómo sabría a donde volver? y que variables locales eliminar?.

**Solución:**

```Go
type Loop struct {
	loopStart  int
	scopeDepth int
}
```

y luego, al `Compiler` le agregamos un `loops []Loop` (muy similar a los `locals`). Entonces, respondiendo nuestras dudas:

- sabemos que NO estamos en un ciclo si `len(loops) == 0`
- sabemos adónde debemos saltar con `loopStart`. Cada vez que empieza un ciclo, ya sea un `while` o un `for`, apilamos un nuevo `Loop`, y lo sacamos al terminar. Como usamos un stack, con ciclos anidados el `continue` siempre usa el último, que es el más interno.
- para limpiar el stack, sacamos las variables locales con un `Depth` mayor al `scopeDepth` que había al empezar el ciclo, ya que son las que se declararon adentro. Es similar al `endScope()`, pero no lo usamos por dos razones: sus `OP_POP` están al final de cada bloque, y el `continue` no pasa por ahí, va directo al inicio. Además saca las variables de `locals` porque en los bloques ya no se necesitan más, en cambio, en este caso todavía se necesita compilar lo que sigue del cuerpo, porque recién en runtime se define si se ejecuta la rama del `continue` o no. Por eso el `continue` solo emite los `OP_POP`, y las variables se sacan de `locals` recién en el `endScope()` final.

Finalmente, `continueStatement()` se ejecuta luego de encontrar un `continue`. Primero valida que lo siga un `;` y que esté dentro de un ciclo. Si no lo está, reporta el error y sigue compilando: el código está bien escrito, pero rompe una regla del lenguaje, igual que redeclarar una variable local (ver [Manejo de errores](6-compiler.md#manejo-de-errores)). Si está todo bien, obtiene el ciclo del tope del stack de loop, emite un `OP_POP` por cada variable local con profundidad mayor a la del ciclo y, por último, emite un `OP_LOOP` hacia el `loopStart` que guardó ese ciclo al crearse. En el `while` es la condición, y en el `for`, el incremento: por eso en el `for` apilamos el `Loop` recién después de compilar el incremento, cuando ya conocemos su posición.

En la VM no hace falta ningún cambio: el `continue` se arma solo con instrucciones que ya existían (`OP_POP` y `OP_LOOP`).

### Break

Un `break`, a diferencia del `continue`, no corta solo la iteración actual: termina el ciclo por completo. Tanto en un `for` como en un `while`, salta a lo que sigue después del ciclo, sin volver a evaluar la condición ni ejecutar el incremento. Si hay ciclos anidados, sale del más interno.

El primer paso es agregarlo como un token, para que el scanner lo reconozca como palabra reservada.

Al igual que el `continue`, el `break` es un statement, por lo que será manejado dentro de `statement()`, con los mismos requisitos: finalizar con `;`, estar dentro de un ciclo y dejar el stack en el mismo estado, sacando las variables locales del ciclo antes de saltar.

Luego, pasando al compilador, sabemos que emitiremos un `OP_JUMP`, pero como inicialmente lo hacemos con un placeholder y luego lo reemplazamos, surge la duda de cómo resolverlo en este caso. El ciclo se compila dentro de `forStatement()` o `whileStatement()`, por lo que dentro de `breakStatement()` no podemos saber dónde va a terminar.

Entonces, se dividen las responsabilidades, y tenemos que ver cómo combinarlas:

- el `break` emite un `OP_JUMP` con un placeholder en el bytecode, que después hay que completar
- los ciclos saben dónde terminan, y por lo tanto cómo completarlo, pero necesitan saber dónde quedó cada placeholder

Es una especie de comunicación entre ambos statements, y ya tenemos un struct que nos ayuda con esto: `Loop`. Le agregamos un campo donde cada `break` indica dónde quedó su placeholder, para que el ciclo lo complete al terminar. Como puede haber más de un `break` en un mismo ciclo, es un slice:

```Go
type Loop struct {
	loopStart  int
	scopeDepth int
	breakJumps []int
}
```

Entonces, cuando nos encontramos con un `break`, hacemos las mismas validaciones que en el `continue`. Si está todo ok, emitimos un `OP_POP` por cada variable local con profundidad mayor a la del ciclo. Finalmente, emitimos el `OP_JUMP` y guardamos la posición de su placeholder en `breakJumps`.

Luego, al terminar cada ciclo, llamamos a `patchBreakJumps()`, que completa todos los `break` del ciclo actual. Lo hacemos después del `OP_POP` de salida porque los dos caminos de salida llegan con el stack distinto: cuando la condición da falsa, la condición todavía está en el stack y ese `OP_POP` la saca, pero cuando se ejecuta un `break`, la condición ya se sacó al entrar al cuerpo. Si el `break` cayera en ese `OP_POP`, sacaría un valor de más. En un `for` sin condición no hay `OP_POP` de salida, así que los `break` caen justo después del último `OP_LOOP`.

Igual que con el `continue`, en la VM no hace falta ningún cambio: el `break` se arma con `OP_POP` y `OP_JUMP`.

### Limitaciones y posibles mejoras

- **Es una extensión de Lox.** Lox no tiene `break` ni `continue`, así que un programa que los use puede no funcionar en otras implementaciones del lenguaje.

- **Solo en el bytecode compiler, pero el scanner es compartido.** Como el tree-walk interpreter usa el mismo scanner, `break` y `continue` pasaron a ser palabras reservadas en los dos modos, pero si se usan como statements sin `--vm`, el parser del tree-walk da un error de sintaxis.

- **Solo salen del ciclo más interno.** No hay forma de salir de varios ciclos a la vez, como un `break` con un número de niveles (por ej `break 2` si está anidado en 2 loops) o con el tipo de ciclo del que se quiere salir (`break while`). Son extensiones que se proponen en [este hilo de r/ProgrammingLanguages](https://www.reddit.com/r/ProgrammingLanguages/comments/o671sd/what_is_the_correct_approach_to_implementing/).

- **El tamaño del salto.** Al usar `OP_JUMP` y `OP_LOOP`, el salto se guarda en un operando de 2 bytes, así que un `break` o un `continue` puede saltar como máximo 65.535 bytes: si el cuerpo del ciclo es más grande, el compilador da error. No es una limitación nueva (es la misma de los `if` y los ciclos), pero vale la pena destacarla.

- **El código inalcanzable se compila igual.** Lo que viene después de un `break` o un `continue` en el mismo bloque nunca se ejecuta, pero el compilador lo emite de todas formas:

  ```c
  for (var i = 0; i < 10; i = i + 1) {
    break;     // o continue;
    print i;   // nunca se ejecuta, pero se compila
  }
  ```

  Una posible mejora sería no emitir ese código, quizá empezando por el caso más simple: lo que sigue a un `break` o `continue` dentro de su mismo bloque, que nunca se ejecuta. Si el `break` está dentro de un `if`, lo que viene después del `if` sí puede llegar a ejecutarse si la condición es falsa (lo sabríamos recién en runtime), así que eso hay que seguir compilándolo.

### Tests

Para los extras agregados, hicimos distintos tipos de tests:

> unitarios de los **tokens**:

```bash
go test ./internal/scanner -run '^TestContinueAndBreak$'
```

> unitarios del **compiler**:

```bash
go test ./internal/compiler -run '^TestCompile(Continue|Break|LoopsAreClosed)$'
```

> real tests: incorporamos un test con usos reales de `break`  y `continue` ([5-break-continue](../real-tests/5-break-continue.lox)):

```bash
python3 real-tests/script.py "go run ./cmd --vm" 5-break-continue
```

TO DO: ver de agregar funciones + break / continue

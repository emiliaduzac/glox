## Extra

Como extra, decidimos incorporar **break** y **continue** a los ciclos. Como vimos en clase, cuando queremos agregar funcionalidad nueva, tenemos que ir en el orden correspondiente: _scanner → compiler → vm_.

> Estas funcionalidades solo están implementadas en el bytecode compiler, el tree-walk interpreter no las soporta.

### Continue

`continue` corta la iteración actual de un ciclo y pasa directamente a la siguiente: lo que queda del cuerpo no se ejecuta. En un `while`, vuelve a evaluar la condición directo, pero en un `for` tiene que ejecutar el incremento y recién después volver a la condición.

En primer lugar, debemos agregarlo como palabra reservada entre los [tokens](../internal/token/token.go). Con eso alcanza para que el scanner lo identifique, ya que, al leer una palabra, la busca en el map de palabras reservadas: si la encuentra, le asigna ese tipo de token, y si no, la considera un identificador.

Luego, seguimos por el compiler. Sabemos que un `continue` debe ser un **statement**, ya que no produce ningún valor sino que genera un efecto: saltear una iteración de un ciclo. Con eso ya sabemos muchas cosas:

1. lo tenemos que incorporar en `statement()`
2. probablemente emita un `OP_LOOP`, ya que debe volver atrás en el código
3. al no ser una expresión no tenemos que modificar la tabla del _Pratt Parser_
4. finaliza con un `;`
5. tiene que dejar el stack tal cual lo encuentra. Con esto vemos que si en el cuerpo de un ciclo se definen variables locales, el `continue` antes de volver al inicio, debe asegurarse de sacar esas variables del stack para que la nueva iteración empiece limpia. Cuando no hay `continue`, esto se limpia solo al llegar al `endScope()`.
6. si o si deben estar dentro de un ciclo, en caso contrario sería un error. Pero también debe soportar estar anidado dentro de más ciclos o ifs.

Primero agregamos el `CONTINUE` al `statement()`, pero ahí nos surgió la duda de como seguir: un nuevo statement no tiene conocimiento de los demás statements, como sabe si efectivamente está en un ciclo? y como sabría a donde volver? y y que variables locales eliminar?.

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

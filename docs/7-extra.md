## Extra

Como extra, decidimos incorporar **break** y **continue** a los ciclos. Como vimos en clase, cuando queremos agregar funcionalidad nueva, tenemos que ir en el orden correspondiente: _scanner → compiler → vm_.

> Estas funcionalidades solo están implementadas en el bytecode compiler, el tree-walk interpreter no las soporta.

### Continue

`continue` corta la iteración actual de un ciclo y pasa directamente a la siguiente: lo que queda del cuerpo no se ejecuta. En un `while`, vuelve a evaluar la condición directo, pero en un `for`, tiene que ejecutar el incremento y recién después volver a la condición.

En primer lugar, debemos agregarlo como palabra reservada entre los [tokens](../internal/token/token.go). Con eso alcanza para que el scanner lo identifique, ya que, al leer una palabra, la busca en el map de palabras reservadas: si la encuentra, le asigna ese tipo de token, y si no, la considera un identificador.

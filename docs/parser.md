## Parser

Luego del anális léxico que hicimos en el scanner (identificar nuestros tokens válidos), pasamos a un análisis sintáctico (armar estructuras con los tokens identificados). Es decir: armar expresiones válidas, con los tokens válidos.

La gramática de Lox, usamos la definición del [libro](https://craftinginterpreters.com/representing-code.html), incluso para la precedencia y la asociatividad.

### Expresiones

Tenemos 4 expresiones válidas: **literales**, **unarias**, **binarias** y **groupos** (definidas en [`expression.go`](../internal/expression/expression.go)). Como a su vez, estas pueden estar compuestas de más expresiones, debemos poder soportar la recursividad al definirlas.

Como Go no tiene clases y herencias, utilizamos la siguiente estratégia:

```
type Expr interface {
	Expression()
}
```

Para ser una expresión válida en nuestra implementación, alcanza con implementar la función `Expression()`. Por ejemplo:

```
type Unary struct {
	operator token.Token
	right    Expression
}

func (u *Unary) Expression() {}
```

Para incorporar funciones (`Evaluate`, `Print`, etc) no hay que modificar ninguna de las expresiones ya existentes, ya que estas no definen su propio comportamiento para cada posible operación. 
Optamos por implementarlo de una forma más "funcional", donde cada operación hace un `switch` para definir el comportamiento de cada tipo posible. 
En el libro descarta esta opción ya que en Java, es una cadena de `if` que se ejecutan linealmente hasta encontrar el caso que corresponda (los casos definidos más abajo tardarían más). En Go esto no sucede ya que optimiza la búsqueda entre los casos.
Lo que si debemos tener en cuenta, es que no podemos olvidarnos de agregar _todos_ los casos posibles. Si nos falta alguno, el compilador no se dará cuenta.
Implementar un **Visitor Pattern** era innecesario en nuestro caso.
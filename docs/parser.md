## Parser

Luego del anális léxico que hicimos en el scanner (identificar nuestros tokens válidos), pasamos a un análisis sintáctico (armar estructuras con los tokens identificados). Es decir: armar expresiones válidas, con los tokens válidos.

Responsabilidades del parser:

- Dada una lista de tokens, armar el AST (**Árbol de Sintaxis Abstracta**) correspondiente.
- Frente a secuencias inválidas de tokens, detectar y levantar errores. Tiene que detectar cuando la secuencia de tokens no forma ninguna expresión válida según la gramática.

Para la gramática de Lox, usamos la definición del [libro](https://craftinginterpreters.com/representing-code.html), también para la precedencia y la asociatividad.

### Expresiones

Tenemos 4 expresiones válidas: **literales**, **unarias**, **binarias** y **grupos** (definidas en [`expression.go`](../internal/expression/expression.go)). Como a su vez, estas pueden estar compuestas de más expresiones, debemos poder soportar la recursividad al definirlas.

Como Go no tiene clases y herencias, utilizamos la siguiente estratégia:

```go
type Expr interface {
	Expression()
}
```

Para ser una expresión válida en nuestra implementación, alcanza con implementar la función `Expression()`. Por ejemplo:

```go
type Unary struct {
	Operator token.Token
	Right    Expr
}

func (u *Unary) Expression() {}
```

Para incorporar funciones (`Evaluate`, `String`, etc) no hay que modificar ninguna de las expresiones ya existentes, ya que estas no definen su propio comportamiento para cada posible operación. 
Optamos por implementarlo de una forma más "funcional", donde cada operación hace un `switch` para definir el comportamiento de cada tipo posible. 
En el libro descarta esta opción ya que en Java, es una cadena de `if` que se ejecutan linealmente hasta encontrar el caso que corresponda (los casos definidos más abajo tardarían más), y justamente se define que un parser debería ser rápido. En Go esto no sucede ya que optimiza la búsqueda entre los posibles casos del `switch`.
Lo que si debemos tener en cuenta, es que no podemos olvidarnos de agregar _todos_ los casos posibles. Si nos falta alguno, el compilador no se dará cuenta.
Implementar un **Visitor Pattern** era innecesario en nuestro caso.

### Precedencia y asociatividad

Las reglas de precedencia y asociatividad buscan resolver dos preguntas:

1. en una expresión con distintos operadores, cual resuelvo primero?
2. en una expresión con muchos operadores del mismo tipo, cual resuelvo primero?

La gramática con precedencia y asociatividad queda definida:

```
expression → equality ;

equality   → comparison ( ( "!=" | "==" ) comparison )* ;

comparison → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;

term       → factor ( ( "-" | "+" ) factor )* ;

factor     → unary ( ( "/" | "*" ) unary )* ;

unary      → ( "!" | "-" ) unary | primary ;

primary    → NUMBER | STRING | "true" | "false" | "nil" | "(" expression ")" ;
```

1. **primary es la de mayor precedencia**: si se cruza con un número, por ejemplo, se resuelve inmediatamente.

2. **equality es la de menor precedencia**: es lo último que se resuelve, hasta no tener ambos lados de una comparación ya resueltos, no se compara nada.

Para **resolver** una expresión, el proceso va de arriba hacia abajo: **Recursive Descent Parser**. Es decir: de menor a mayor precedencia: `expression → equality → comparison → term → factor → unary → primary`, y el árbol se arma de izquierda a derecha a medida que se desciende y se vuelve.

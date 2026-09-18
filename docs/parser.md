## Parser

Luego del anális léxico que hicimos en el scanner (identificar nuestros tokens válidos), pasamos a un análisis sintáctico (armar estructuras con los tokens identificados). Es decir: armar expresiones válidas, con los tokens válidos.

La gramática de Lox, usamos la definición del [libro](https://craftinginterpreters.com/representing-code.html), incluso para la precedencia y la asociatividad.

### Expresiones

Tenemos 4 expresiones válidas: **literales**, **unarias**, **binarias** y **groupos** (definidas en [`expression.go`](../internal/expression/expression.go)). Como a su vez, estas pueden estar compuestas de más expresiones, debemos poder soportar la recursividad al definirlas.

Como Go no tiene clases y herencias, utilizamos la siguiente estratégia:

```
type Expression interface {
	Expression()
}
```

Para ser una expresión válida en nuestra implementación, alcanza con implementar la función `Expression()`.

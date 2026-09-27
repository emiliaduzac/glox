## Parser

Luego del análisis léxico que hicimos en el scanner (identificar nuestros tokens válidos), pasamos a un análisis sintáctico (armar estructuras con los tokens identificados). Es decir: armar expresiones y statements válidos, con los tokens válidos.

Responsabilidades del parser:

- Dada una lista de tokens, armar el AST (**Árbol de Sintaxis Abstracta**) correspondiente.
- Frente a secuencias inválidas de tokens, detectar y levantar errores. Tiene que detectar cuando la secuencia de tokens no forma ninguna expresión válida según la gramática.

Siguiendo el orden en el que se explican en el libro y en clase, en el parser primero detectamos **expressions** y luego incorporamos **statements**. La documentación sigue ese mismo orden.

- [Expresiones](#expresiones)
- [Statements](#statements)

### Expresiones

Las expresiones son maneras de **producir valores**. Sumar, restar, concatenar, etc. El problema es que solo generan un valor, pero **no hacemos nada con él**.

Tenemos varias expresiones válidas en Lox: **literales**, **unarias**, **binarias**, **grupos**, **variables** y **asignaciones** (definidas en [`expression.go`](../internal/expression/expression.go)), y basadas en la definición que el [libro](https://craftinginterpreters.com/representing-code.html) va dando a medida que avanzan los capitulos (y se incorporan nuevas expresiones). Como a su vez, estas pueden estar compuestas de más expresiones, debemos poder soportar la recursividad al definirlas.

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

#### Precedencia y asociatividad

Las reglas de precedencia y asociatividad buscan resolver dos preguntas:

1. en una expresión con distintos operadores, cual resuelvo primero?
2. en una expresión con muchos operadores del mismo tipo, cual resuelvo primero?

La gramática con precedencia y asociatividad queda definida:

```
expression → assignment ;

assignment → IDENTIFIER "=" assignment | equality ;

equality   → comparison ( ( "!=" | "==" ) comparison )* ;

comparison → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;

term       → factor ( ( "-" | "+" ) factor )* ;

factor     → unary ( ( "/" | "*" | "%" ) unary )* ;

unary      → ( "!" | "-" ) unary | primary ;

primary    → NUMBER | STRING | "true" | "false" | "nil" | "(" expression ")" | IDENTIFIER ;
```

1. **primary es la de mayor precedencia**: si se cruza con un número, por ejemplo, se resuelve inmediatamente.

2. **equality es de precedencia "intermedia"**: hasta no tener ambos lados de una comparación ya resueltos, no se compara nada.

3. **assignment es la de menor precedencia**: Con el soporte de assignments que agregamos ahora estas pasan a ser la de menor precedencia absoluta, incluso menor que la de equality. Esto porque no podemos asignar un valor hasta que la parte derecha esté ya resuelta. Además la asignación es asociativa a derecha, por eso podemos hacer `a = b = 5`, porque se resuelve la parte derecha primero, como `b = 5` y retorna `5`, lo que luego hace que se resuelva `a = 5`, y al final entonces queda `a` y `b` con valor igual a `5`. 

Para **resolver** una expresión, el proceso va de arriba hacia abajo: **Recursive Descent Parser**. Es decir: de menor a mayor precedencia: `expression → assignment → equality → comparison → term → factor → unary → primary`, y el árbol se arma de izquierda a derecha a medida que se desciende y se vuelve.

Como se evalúa dicho árbol será explicado en la implementación del interpreter: [interpreter.md](interpreter.md).

#### El truco de L-Value y R-Value en la asignación

Acá lo que pasaba es que nuestro parser solo mira un token, entonces en el caso de, por ejemplo, leer `x` no puede saber qué viene después, no sabe si es `x = 1` o `x + 1` hasta que no ve el `=`. Ahí es que seguimos el truco que se propone en el libro de parsear el lado izquierdo como si fuera una expresión, luego te fijás si lo que sigue es un `=`, si lo es ahí volvés a mirar para atrás entonces si lo que tenías antes es válido para ir con el `=`. Es decir, si tenés una variable está bien, y entonces convierte el nodo en un `Assign`, si no en ese momento ya te das cuenta del error.

#### Manejo de errores

Al encontrarse con un error, el parser devuelve un error detallado explicando la linea y error hallado, sin seguir parseando el resto de tokens. Todos se generan en `primary()`, que es el último nivel de la tabla de precedencia: si ningún nivel anterior pudo consumir el token, ahí es donde se sabe que no hay una expresión válida.
Con el agregado del truco del libro, se agrega el error "invalid assignment target". Ejemplo `1 + 2 = 3` es inválido, o un `(a) = 3` en Lox es inválido. Y se lanza ese error sintáctico. 

#### Tests

Para testear la correcta detección de las expresiones por parte del parser, definimos tests unitarios buscando cubrir la mayoría de casos posibles: ([`parser_test.go`](../internal/parser/parser_test.go)).
Para ejecutarlos: 

```bash
go test ./internal/parser
```

### Statements

Los statements **usan** expresiones para **modificar el mundo a su alrededor** →  guardar una variable, leer input, controlar el flujo, imprimir por pantalla. En vez de producir valores, **producen efectos**.

Para los statements, hicimos el mismo truco que con las expressions:

```go
type Stmt interface {
	Statement()
}

type PrintStmt struct {
	Expr expression.Expr
}
func (p *PrintStmt) Statement() {}
```

- **Print Statements**: evalúa una expresión y muestra su resultado por pantalla. Solo puede recibir expresiones, no otros statements.
- **Expression Statements**: expresiones seguidas de `;`. Son expresiones que pueden tener efectos secundarios. Cuando se la llama a la expresión suelta, se evalúa, los efectos secundarios (si los hay) se producen, y el valor resultante se descarta.
- **Var statements**: declara una variable nueva en el entorno actual, que puede o no tener un valor inicial. Como Expr es una interfaz, nil ya es un valor válido para ese campo, por lo que no necesitamos ningún `Optional` ni similares para poder declarar el valor inicial (ya que Go no tiene `Optionals`).
- **Block statements**: declara una lista de sentencias y declaraciones entre llaves `{ ... }`. Agregar esto a Glox le permite ahora delimitar en scope local y dentro de él declarar las variables locales y también con esto soporta declarar variables locales dentro.

Se expande nuestra gramática:

```
program     → declaration* EOF ;
declaration → varDecl | statement ;
statement   → exprStmt | printStmt | block ;
block       → "{" declaration* "}" ;

varDecl     → "var" IDENTIFIER ( "=" expression )? ";" ;
exprStmt    → expression ";" ;
printStmt   → "print" expression ";" ;

expression  → assignment ;
assignment  → IDENTIFIER "=" assignment | equality ;
equality    → comparison ( ( "!=" | "==" ) comparison )* ;
comparison  → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
term        → factor ( ( "-" | "+" ) factor )* ;
factor      → unary ( ( "/" | "*" | "%" ) unary )* ;
unary       → ( "!" | "-" ) unary | primary ;
primary     → NUMBER | STRING | "true" | "false" | "nil" | "(" expression ")" | IDENTIFIER ;
```

#### Manejo de errores

Al incorporar statements surge un nuevo error: toda sentencia debe terminar en `;`. Si falta, el parser lanza un error indicando qué encontró en su lugar.
Además agregamos el chequeo del error de no cerrar con `}` porque también podría pasar que siga y se tope con un EOF, así que en ese caso se informa con el error "expected '}' after block". 
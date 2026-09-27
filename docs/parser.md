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

Tenemos varias expresiones válidas en Lox: **literales**, **unarias**, **binarias**, **grupos**, **variables**, **asignaciones** y **operadores lógicos** (definidas en [`expression.go`](../internal/expression/expression.go)), y basadas en la definición que el [libro](https://craftinginterpreters.com/representing-code.html) va dando a medida que avanzan los capitulos (y se incorporan nuevas expresiones). Como a su vez, estas pueden estar compuestas de más expresiones, debemos poder soportar la recursividad al definirlas.

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

assignment → IDENTIFIER "=" assignment | logic_or ;

logic_or   → logic_and ( "or" logic_and )* ;

logic_and  → equality ( "and" equality )* ;

equality   → comparison ( ( "!=" | "==" ) comparison )* ;

comparison → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;

term       → factor ( ( "-" | "+" ) factor )* ;

factor     → unary ( ( "/" | "*" | "%" ) unary )* ;

unary      → ( "!" | "-" ) unary | primary ;

primary    → NUMBER | STRING | "true" | "false" | "nil" | "(" expression ")" | IDENTIFIER ;
```

1. **primary es la de mayor precedencia**: si se cruza con un número, por ejemplo, se resuelve inmediatamente.

2. **equality es de precedencia "intermedia"**: hasta no tener ambos lados de una comparación ya resueltos, no se compara nada.

3. **assignment es la de menor precedencia**: Con el soporte de assignments que agregamos ahora estas pasan a ser la de menor precedencia absoluta, incluso menor que la de equality. Esto es porque no podemos asignar un valor hasta que la parte derecha esté ya resuelta. Además la asignación es asociativa a derecha, por eso podemos hacer `a = b = 5`: se resuelve la parte derecha primero, como `b = 5` y retorna `5`, lo que luego hace que se resuelva `a = 5`, y al final entonces queda `a` y `b` con valor igual a `5`. 

4. **operadores lógicos**: estos nos permiten anidar varias condiciones, y se evalúan de una forma particular: contienen "cortocircuitos", una manera distinta de evaluarse que no permite que sean simples operadores binarios. En el caso del `or`, se evalúa primero la expresión de la izquierda. Si esta es verdadera, ya no se necesita evaluar la expresión de la derecha, ya cumple. En el caso del `and`, si el término de la izquierda es falso, ya no hay más nada que evaluar, ya que sin importar el resultado de la expresión derecha, no va a cumplirse que ambas sean verdaderas. En cuanto a su precedencia, el `or` tiene menor precedencia. Por ejemplo `TRUE or FALSE and TRUE` se resuelve a: `TRUE or (FALSE and TRUE)` → `TRUE or FALSE` → `TRUE`. Por último, destacar que en lox los operadores lógicos son considerados expresiones ya que devuelven el valor del operando (no un booleano).

Para **resolver** una expresión, el proceso va de arriba hacia abajo: **Recursive Descent Parser**. Es decir: de menor a mayor precedencia: `expression → assignment → logic_or → logic_and → equality → comparison → term → factor → unary → primary`, y el árbol se arma de izquierda a derecha a medida que se desciende y se vuelve.

Como se evalúa dicho árbol será explicado en la implementación del interpreter: [interpreter.md](interpreter.md).

#### El truco de L-Value y R-Value en la asignación

Como nuestro parser mira solo de a un token, al leer `x` no puede saber que viene después: si una asignación (`x = 1`) o una expresión (`x + 1`). En el libro se propone un truco: parsear el lado izquierdo como si fuera una expresión y luego fijarse si lo que sigue es un `=`. Si lo es, volvés a mirar hacia atrás: si lo anterior era válido (una variable), es correcto y se convierte en un nodo `Assign`. Si no, es un error.

#### Manejo de errores

Al encontrarse con un error, el parser devuelve un error detallado explicando la linea y error hallado, sin seguir parseando el resto de tokens. Inicialmente, como `primary()` era el último nivel de la tabla de precedencia, todos los errores ocurrían allí: si ningún nivel anterior pudo consumir el token, ahí es donde se sabe que no hay una expresión válida.
Con el agregado del truco del libro, se agrega el error "invalid assignment target". Ejemplo `1 + 2 = 3` es inválido, o un `(a) = 3` en Lox es inválido. Y se lanza ese error sintáctico. 

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

- **Var statements**: declara una variable nueva en el entorno actual, que puede o no tener un valor inicial. Como Expr es una interfaz, `nil` ya es un valor válido para ese campo, por lo que no necesitamos ningún `Optional` ni similares para poder declarar el valor inicial (ya que Go no tiene `Optionals`).

- **Block statements**: declara una lista de sentencias y declaraciones entre llaves `{ ... }`. Agregar esto a Glox le permite ahora delimitar en scope local y dentro de él declarar las variables locales y también con esto soporta declarar variables locales dentro.

- **If statements**: al incorporar control de flujo aparecen los condicionales. Nos permiten guiar el flujo de ejecución: dada una condición, ejecutar una porción de código u otra. Glox no contiene `elif` ni `if else`, por lo que para tener varias condiciones si o si hay que anidar (o si se quiere varias condiciones en un mismo `if`, utilizar operadores lógicos). Para evitar ambiguedades, se resuelve que en caso de haber varios `if`, cada `else` se resuelve al más cercano. En el código esto se da naturalmente por como parseamos y construimos recursivamente el árbol AST. Al encontrar un `if statement`, se consume su `else` más cercano directamente.

- **While statements**: permiten repetir una porción de código una cantidad de veces indefinida: mientras se cumpla una condición.

- **For statements (?)**: técnicamente son lo mismo que un while, pero como repiten una semántica muy utilizada por los usuarios, se resuelve con syntactic sugar para facilitar su uso. Entonces se implementan por detrás con un while (no es un statement nuevo), pero se mantiene la palabra reservada `for` con su semántica, para que el usuario pueda hacer uso del mismo, sin saber como se implementa por detrás. Un `for` entonces contiene: (1) inicializador, (2) condición e (3) incremento: `for ( <initializer> ; <condition> ; <increment>) <body>`.

Se expande nuestra gramática:

```
program     → declaration* EOF ;
declaration → varDecl | statement ;
statement   → exprStmt | forStmt | ifStmt | printStmt | whileStmt | block ;
forStmt     → "for" "(" ( varDecl | exprStmt | ";" ) expression? ";" expression? ")" statement ;
ifStmt      → "if" "(" expression ")" statement ( "else" statement )? ;
whileStmt   → "while" "(" expression ")" statement ;
block       → "{" declaration* "}" ;

varDecl     → "var" IDENTIFIER ( "=" expression )? ";" ;
exprStmt    → expression ";" ;
printStmt   → "print" expression ";" ;

expression  → assignment ;
assignment  → IDENTIFIER "=" assignment | logic_or ;
logic_or    → logic_and ( "or" logic_and )* ;
logic_and   → equality ( "and" equality )* ;
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
Con `if` y `while` aparecen errores análogos: falta del `(` o del `)` alrededor de la condición, o una condición vacía.

#### Tests

Para testear la correcta detección de las expresiones y statements por parte del parser, definimos tests unitarios buscando cubrir la mayoría de casos posibles: ([`parser_expressions_test.go`](../internal/parser/parser_expressions_test.go), [`parser_statements_test.go`](../internal/parser/parser_statements_test.go)).
Para ejecutarlos: 

```bash
go test ./internal/parser
```

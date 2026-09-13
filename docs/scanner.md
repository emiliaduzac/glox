## Scanner

La responsabilidad del scanner es tomar código fuente y traducirlo a tokens, es un análisis léxico. 
Aún no es importante como se relacionan estos tokens, sino solamente saber identificar cuales son válidos.

Para eso, lo primero que debemos definir son los tokens que consideraremos válidos en nuestro lenguaje. Para eso, utilizamos la definición de Lox en el libro [Crafting Interpreters](https://craftinginterpreters.com/scanning.html).
A su vez, los tokens están compuestos por dos partes:
- lexema: secuencia mínima de caracteres crudos que pueden significar algo
- token: el lexema más la información de qué tipo de token es

### Tokens válidos

> De un solo caracter

| caracter | nombre token |
| --- | --- |
| ( | LEFT_PAREN |
| ) | RIGHT_PAREN |
| { | LEFT_BRACE |
| } | RIGHT_BRACE |
| , | COMMA |
| . | DOT |
| - | MINUS |
| + | PLUS |
| ; | SEMICOLON |
| * | STAR |
| % | PERCENT |

> De uno o dos caracteres

| caracter/es | nombre token |
| --- | --- |
| / | SLASH | 
| ! | BANG |
| != | BANG_EQUAL |
| = | EQUAL |
| == | EQUAL_EQUAL |
| > | GREATER |
| >= | GREATER_EQUAL |
| < | LESS |
| <= | LESS_EQUAL |

En caso de encontrar el token `//` es considerado un comentario, y el contenido que lo acompañe será descartado. Los comentarios soportados por Lox son de única linea, no de bloque (`/* ... */`).

> Literales

| nombre token | regla |
| --- | --- |
| IDENTIFIER | secuencia de caracteres (comienza con letras o `_`, puede contener números) |
| STRING | secuencia de caracteres entre comillas |
| NUMBER | uno o más dígitos, opcionalmente seguidos de un `.` y más dígitos. No puede comenzar ni finalizar con `.` |

> Palabras clave

| palabra clave | nombre token |
| --- | --- |
| and | AND |
| class | CLASS |
| else | ELSE |
| false | FALSE |
| fun | FUN |
| for | FOR |
| if | IF |
| nil | NIL |
| or | OR |
| print | PRINT |
| return | RETURN |
| super | SUPER |
| this | THIS |
| true | TRUE |
| var | VAR |
| while | WHILE |

Cabe destacar que las palabras claves deben estar en mínuscula.

> Fin del archivo

| nombre token |
| --- |
| EOF |


// TODO: revisar que hacemos con estos casos:
```
Palabras que estaban en clase pero no en el libro:
- %

Palabras que estan en el libro pero no en clase:
- . (dot)
- CLASS
- SUPER
- THIS
```

### Funcionamiento del scanner

La gramática léxica es un análisis lineal, por lo que recorremos el codigo caracter por caracter. 
El scanner está compuesto por:

```go
type Scanner struct {
	tokens  []token.Token // lista de tokens escaneados
	source  string        // código fuente a escanear
	current int 		  // índice del caracter que el scanner está a punto de leer
	start   int 		  // índice del primer caracter del token actual
}
```

A medida que vamos recorriendo los caracteres en un loop, tenemos que ir identificando su validez. Para hacerlo, el scanner cuenta con dos loops:
- loop inicial: identifica tokens (en `Scan`)
- loop interno: itera caracteres (en `scanToken`)

1. Aquellos **tokens de único caracter** son simples: al detectarlo lo almacenamos en la lista de tokens y continuamos con la siguiente iteración del loop inicial.

2. Cuando nos encontramos con un caracter que podrían ser tanto un **token de único caracter como token de dos caracteres**, debemos analizar también el caracter siguiente para detectar bajo que caso nos encontramos.
Si el siguiente caracter _no_ compone un token de dos caracteres, se maneja como un token simple.
Si el siguiente caracter _si_ compone un token de dos caracteres, se avanza el puntero `current` y se obtiene el token de dos caracteres, almacenandolo en la lista de tokens.

3. Para los **literales que representan strings**, debemos seguir mirando caracteres hasta hallar la comilla de cierre. Aquí es donde nos sirve mantener `start` fijo al inicio del lexema. Se irán leyendo caracteres hasta el final, y gracias a `start` y `current` podemos obtener todo el lexema actual: `s.source[s.start : s.current]`. Si no se encuentra una comilla de cierre, se devuelve un error. 
Detalles a destacar:
	- Los strings comienzan y terminan con `"` (comillas dobles), no simples (`'`).
	- Lox soporta los strings multi-linea.
	- El token contiene el lexema (string fuente, incluyendo las comillas) y un literal (string fuente sin comillas)

4. Para los **literales que representan números** se maneja de la siguiente forma: una vez que detectamos que el primer caracter de un lexema es un dígito, comenzamos a leer los siguientes caracteres en el loop interno. Para ser considerado un número solo pueden seguir números o un punto para marcar decimales. Finalmente se obtiene el lexema válido de la misma manera que con los strings: `s.source[s.start : s.current]`.
A destacar:
	- Solo puede haber _un_ punto (no más), ni tampoco puede ser el último caracter del lexema.
	- El token contiene el lexema (string fuente) y un literal (string ya convertido a número). Del libro _Crafting Interpreters: "All numbers in Lox are floating point at runtime"._

5. **Identificadores y palabras clave**: tanto los identificadores o palabras clave comienzan con letras. Además, los identificadores pueden comenzar con un guión bajo (`_`) y continuar con números. Siguiendo el _principio de maximal munch_, debemos leer la secuencia completa antes de distinguir si se trata de un **identificador** o **una palabra clave**. Esto evita que definamos una secuencia como palabra clave antes de leerla por completa. Por ejemplo, si nos encontramos con `true`, antes de asignarla como un token key word `TRUE`, debemos asegurarnos que no siga ningún otro caracter que la convierta en un identificador, como por ejemplo `truemanshow`.
Entonces, resumiendo, para el caso donde nos encontramos con una letra, recorremos los caracteres hasta llegar al final de la misma (fin de la secuencia `source` completa, o por ejemplo un espacio u otro caracter no soportado). Una vez leída la secuencia completa, chequeamos si el lexema hallado es una palabra clave. Si lo es, le asignamos el token que corresponda. Si no, le asignamos el token `IDENTIFIER`.

6. **Casos particulares**: tanto los saltos de línea, como los espacios en blanco y comentarios son ignorados.

Finalmente, luego de leer todos los tokens del código fuente, se agrega por prolijidad un token `EOF`, indicando que no hay más contenido.

7. **Errores**: como ya mencionamos, el trabajo del scanner es leer secuencialmente, caracter por caracter, y validar que los tokens sean válidos, es decir: que estén dentro de los aceptados por el lenguaje. También nombramos dos iteraciones para lograr esto: la que itera tokens (iteración inicial) y la que itera caracteres (iteración interna, para detectar tokens de varios caracteres). En esta implementación, apenas se encuentra un caracter inválido, se devuelve un error y se corta el escaneo, sin seguir buscando el resto de los tokens.
Ejemplo de errores:
	- caracteres únicos no reconocidos (`@`, `#`, `$`, etc)
	- cadenas que no finalizan con comillas (e.j: `"hello`)
	- números que contienen más de un punto o finalizan con punto (e.j: `3.`, `12.12.12`). Vale la pena aclarar que, dado que el `.` es un token válido por sí mismo (`DOT`), y los números no pueden comenzar con `.`, la secuencia `.123` no es un error, sino que dos tokens consecutivos, `DOT` + `NUMBER`.

### Tests

El scanner cuenta con un conjunto de tests unitarios (`scanner_test.go`) cubriendo tanto los casos válidos como inválidos. 
Para ejecutarlos: 

```bash
go test ./internal/scanner
```

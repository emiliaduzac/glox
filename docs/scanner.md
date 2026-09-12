## Scanner

La responsabilidad del scanner es tomar código fuente y traducirlo a tokens, es un análisis léxico. 
Aún no es importante como se relacionan estos tokens, sino solamente saber identificar cuales son válidos.

Para eso, lo primero que debemos definir son los tokens que consideraremos válidos en nuestro lenguaje. Para eso, utilizamos
la definición de Lox en el libro [Crafting Interpreters](https://craftinginterpreters.com/the-lox-language.html).
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
| / | SLASH | 
| --- | --- |
| ! | BANG |
| != | BANG_EQUAL |
| = | EQUAL |
| == | EQUAL_EQUAL |
| > | GREATER |
| >= | GREATER_EQUAL |
| < | LESS |
| <= | LESS_EQUAL |

En caso de encontrar el token `//` es considerado un comentario, y el contenido que lo acompañe será descartado.

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


```
Palabras que estaban en clase pero no en el libro:
- %
- 

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
	..
}
```

A medida que vamos recorriendo los caracteres, tenemos que ir identificando su validez. 
Aquellos tokens de único caracter son simples: los almacenamos en la lista y avanzamos con el siguiente.
Cuando nos encontramos con un caracter que podrían ser tanto un token de único caracter como token de dos caracteres,
debemos analizar también el caracter siguiente para detectar bajo que caso nos encontramos.


## Scanner

La responsabilidad del scanner es tomar código fuente y traducirlo a tokens, es un análisis léxico. 
Aún no es importante como se relacionan estos tokens, sino solamente saber identificar cuales son válidos.

Para eso, lo primero que debemos definir son los tokens que consideraremos válidos en nuestro lenguaje. Para eso, utilizamos
la definición de Lox en el libro [Crafting Interpreters](https://craftinginterpreters.com/the-lox-language.html).


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
| / | SLASH | 

> De uno o dos caracteres

| caracter/es | nombre token |
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


Palabras que estaban en clase pero no en el libro:
- %
- 

Palabras que estan en el libro pero no en clase:
- . (dot)
- CLASS
- SUPER
- THIS
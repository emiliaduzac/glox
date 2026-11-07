## Compiler

Siguiendo la clase y el libro, nuestro compilador tendrá dos responsabilidades:

- **Parsear** los tokens para darles estructura y entender qué significan.
- **Emitir** las instrucciones que generan, en el orden en que tienen que ejecutarse.

Es decir que reemplazará a nuestro **parser** y a nuestro **resolver**. Esto se define como un **Single Pass Compiler**: de los tokens pasamos directamente a bytecode, emitiendo a medida que leemos (la entrada son los _tokens_ y la salida son los _chunks_).
Otras implementaciones de compiladores aprovechan el AST parseado, pero el libro elige un single pass compiler por su simplicidad. Además, indica que originalmente muchos compiladores se hacían de esta forma por razones de memoria y que este enfoque podría no servir para cualquier lenguaje, pero Lox lo permite por ser de tipado dinámico: el compilador no necesita conocer los tipos de lo que todavía no leyó, porque se resuelven en tiempo de ejecución.

### Comparación con la primera entrega

**Tree-walk**: scanner → parser → resolver → interpreter

**Bytecode**: scanner → compiler → VM

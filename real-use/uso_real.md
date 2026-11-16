# Uso real de Glox

Como un caso de uso real de nuestro lenguaje, decidimos implementar los siguientes algoritmos de ordenamiento: **mergesort**, **quicksort** y **bubblesort**. Esto implicó también crear una **lista enlazada** utilizando Glox. Dado que al momento de implementarlo no contábamos con métodos para clases, usamos funciones que editaban el estado de las clases. A futuro, podría mejorarse haciendo que la lista enlazada y el nodo tengan sus propios métodos en vez de ser modificada por fuera.

La lista enlazada está compuesta por un `Nodo`, que tiene su `dato` y su `próximo`. Al crear una `Lista`, inicializamos con los campos `primero` y `ultimo` en `nil` y `cantidad` 0. Esto nos sirve ya que en Lox los campos no se declaran, sino que se crean al usarlos. Inicializarlos al crear la lista ayuda a que luego no se genere errores en el primer uso de la lista.

Implementamos las funciones típicas que tendría una lista enlazada:

- `estaVacia`: indica si tiene elementos
- `insertarPrimero`: inserta un nodo en `primero`, y hace que `primero.proximo` sea el anterior `primero`. Además, si la lista estaba vacía pasa a ser el `ultimo` también. Suma 1 a la `cantidad`.
- `insertarUltimo`: si la lista estaba vacía, el nuevo nodo será también el `primero`. Luego, hace que el anterior `ultimo` apunte a este, y este pasa a ser el nuevo `ultimo`. También suma 1 a la `cantidad`.
- `borrarPrimero`: hace que el nuevo `primero` sea el `proximo` al que era `primero`. Si esto deja la lista vacía, tanto `primero` como `ultimo` pasan a ser `nil`. Resta uno a la `cantidad` y devuelve el `dato` del eliminado.
- `verPrimero`: devuelve el `dato` del primer nodo.
- `largo`: devuelve la `cantidad` de la lista

Una vez implementada la lista enlazada, pudimos implementar los algoritmos mencionados. Por cada uno realizamos 4 tests:

1. validar que ordene una lista desordenada de 50 elementos
2. la lista vacía queda igual, no hay errores
3. una lista desordenada con elementos repetidos queda correctamente ordenada
4. una lista ya ordenada queda ordenada, también de 50 elementos

Usamos listas de 50 elementos porque quicksort, con el primer elemento como pivote, anida una llamada por elemento en el peor caso (una lista ya ordenada), y la VM admite como máximo 63 llamadas anidadas.

Para ejecutar los tests:

```bash
go run ./cmd --vm ./real-use/sorting_lists.lox
```

Todo esto puede encontrarse en [`sorting_lists.lox`](sorting_lists.lox)

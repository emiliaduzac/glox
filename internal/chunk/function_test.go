package chunk_test

import (
	"testing"
	"tp-lox/internal/chunk"
	"tp-lox/internal/utils"
)

func TestFunctionString(t *testing.T) {
	t.Run("Una función con nombre se imprime como <fn nombre>", func(t *testing.T) {
		function := chunk.NewFunction()
		function.Name = "suma"

		utils.AssertValue(t, utils.Stringify(function), "<fn suma>")
	})

	t.Run("Una función sin nombre es el script", func(t *testing.T) {
		function := chunk.NewFunction()

		utils.AssertValue(t, utils.Stringify(function), "<script>")
	})
}

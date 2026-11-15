package virtualmachine

import (
	"testing"
	"tp-lox/internal/utils"
)

func TestClassString(t *testing.T) {
	t.Run("Una clase se imprimer con su nombre", func(t *testing.T) {
		utils.AssertValue(t, utils.Stringify(newClass("Auto")), "Auto")
	})
}

package virtualmachine

import (
	"testing"
	"tp-lox/internal/utils"
)

func TestClassString(t *testing.T) {
	t.Run("Una clase se imprime con su nombre", func(t *testing.T) {
		utils.AssertValue(t, utils.Stringify(newClass("Auto")), "Auto")
	})
}

func TestInstanceString(t *testing.T) {
	t.Run("Una instancia se imprime con el nombre de su clase + instance", func(t *testing.T) {
		utils.AssertValue(t, utils.Stringify(newInstance(newClass("Auto"))), "Auto instance")
	})
}

func TestNewInstance(t *testing.T) {
	t.Run("Una instancia nueva arranca sin campos", func(t *testing.T) {
		utils.AssertValue(t, len(newInstance(newClass("Auto")).fields), 0)
	})
}

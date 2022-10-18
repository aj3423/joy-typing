package mode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNumber(t *testing.T) {
	{
		n := []string{"haha", "seven", "two"}
		n2 := replaceNumbers(n)
		assert.Equal(t, []string{"haha", "7", "2"}, n2, "fail 1")
	}
	{
		n := []string{"seven", "two", "thousand", "n", "six", "hundred", "n", "five", "haha"}
		n2 := replaceNumbers(n)
		assert.Equal(t, []string{"7", "2605", "haha"}, n2, "fail 2")
	}
	{
		n := []string{"two", "thousand", "hundred", "n", "five"}
		n2 := replaceNumbers(n)
		assert.Equal(t, []string{"200005"}, n2, "fail 3")
	}
	{
		n := []string{"twenty", "five", "thousand"}
		n2 := replaceNumbers(n)
		assert.Equal(t, []string{"25000"}, n2, "fail 4")
	}
}

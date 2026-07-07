package slice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Filter(t *testing.T) {
	type item struct {
		Value int
	}

	items := []item{
		{Value: 1},
		{Value: 2},
		{Value: 3},
	}

	t.Run("filter", func(t *testing.T) {
		assert.Len(t,

			Filter(items,
				func(item item) bool {
					return item.Value == 2
				},
			), 1)
	})
}

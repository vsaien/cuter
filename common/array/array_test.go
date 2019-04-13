package array

import (
	"fmt"
	"testing"
)

func TestInArrayString(t *testing.T) {
	i := InArrayString([]string{"a", "ccc", "kkk", "io"}, "kkk")
	j := InArrayInt([]int{1, 4, 6, 7}, 7)

	l := InArrayFloat([]float64{1.077, 4.90, 6.889, 7.9090}, 7.9090)
	fmt.Println(i, j, l)
}

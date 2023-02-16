package utils

import (
	"fmt"
	"testing"
)

func TestRandStr(t *testing.T) {
	str := RandStr(6)
	fmt.Println(str)
}

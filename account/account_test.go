package account

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	list := Am.List()
	for _, address := range list {
		fmt.Println(address)
	}
}

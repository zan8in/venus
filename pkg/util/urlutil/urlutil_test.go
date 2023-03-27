package urlutil

import (
	"fmt"
	"testing"
)

func TestTopDomain(t *testing.T) {
	t.Parallel()

	a := TopDomain("adfas.example.com")

	fmt.Println(a)

}

package file

//testing
import (
	"fmt"
	"strings"
	"testing"
	//testing
	//go test -bench=.
	//go test --timeout 9999999999999s
)

func TestLineCount(u *testing.T) {
	__(u)
	count, err := LineCount("10.text")
	if err != nil {
		panic(err)
	}
	if count != 9 {
		panic(count)
	}
	select {}
}

func __(u *testing.T) {
	fmt.Printf("\033[1;32m%s\033[0m\n", strings.ReplaceAll(u.Name(), "Test", ""))
}

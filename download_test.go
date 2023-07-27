package file

//testing
import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
	//testing
	//go test -bench=.
	//go test --timeout 9999999999999s
)

func TestMain(u *testing.T) {
	__(u)

	// /select {}
}

func Benchmark1(u *testing.B) {
	u.ReportAllocs()
	u.ResetTimer()
	for n := 0; n < u.N; n++ {

	}
}

func Benchmark2(u *testing.B) {
	u.RunParallel(func(pb *testing.PB) {
		for pb.Next() {

		}
	})
}

func __(u *testing.T) {
	fmt.Printf("\033[1;32m%s\033[0m\n", strings.ReplaceAll(u.Name(), "Test", ""))
}

func cmd(name string, v ...string) {
	c := exec.Command(name, v...)
	r, err := c.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(r))
}

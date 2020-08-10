package prometheus

import (
	"fmt"
	"testing"
)

func TestReplace(t *testing.T) {
	promQL := NewPromQL(`test{name="$server", server="$instance", type="$server"}`)
	s := promQL.Replace("server", "服务器").Replace("instance", "127.0.0.1").GetValue()
	fmt.Println(s)
}

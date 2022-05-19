package ansihtml

import (
	"runtime"
	"testing"
)

func TestColor(t *testing.T) {
	except := func(hsl string, rgb string) {
		c1, ok1 := parseColor(hsl)
		c2, ok2 := parseColor(rgb)
		if !ok1 {
			buf := make([]byte, 4096)
			buf = buf[:runtime.Stack(buf, false)]
			t.Log(string(buf))
			t.Logf("parseColor: %s", hsl)
			t.FailNow()
		}
		if !ok2 {
			buf := make([]byte, 4096)
			buf = buf[:runtime.Stack(buf, false)]
			t.Log(string(buf))
			t.Logf("parseColor: %s", rgb)
			t.FailNow()
		}
		if c1 != c2 {
			buf := make([]byte, 4096)
			buf = buf[:runtime.Stack(buf, false)]
			t.Log(string(buf))
			t.Logf("hsl: %x", c1)
			t.Logf("rgb: %x\n\n", c2)
			t.FailNow()
		}
	}
	except("hsl( 357, 60%, 57%)", "rgb(211 80  86)")
}

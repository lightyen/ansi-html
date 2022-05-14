package ansihtml_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	ansihtml "github.com/lightyen/ansihtml"
)

var options = ansihtml.Options{
	MinimumContrastRatio: 1,
	ClassPrefix:          "ansi-",
}

func newExpect(t *testing.T, c *ansihtml.Converter) func(input string, expected string) {
	return func(input string, expected string) {
		buf := &bytes.Buffer{}
		c.Reset()
		if err := c.Copy(buf, strings.NewReader(input)); err != nil {
			buf := make([]byte, 4096)
			buf = buf[:runtime.Stack(buf, false)]
			t.Log(string(buf))
			t.Fatal(err)
			return
		}
		received := buf.String()
		if received != expected {
			buf := make([]byte, 4096)
			buf = buf[:runtime.Stack(buf, false)]
			t.Log(string(buf))
			t.Logf("expected: %s", expected)
			t.Logf("received: %s\n\n", received)
			t.FailNow()
		}
	}
}

func newExpectError(t *testing.T, c *ansihtml.Converter) func(input string, target error) {
	return func(input string, target error) {
		buf := &bytes.Buffer{}
		c.Reset()
		err := c.Copy(buf, strings.NewReader(input))
		if !errors.Is(err, target) {
			buf := make([]byte, 4096)
			buf = buf[:runtime.Stack(buf, false)]
			t.Log(err)
			t.Log(string(buf))
			t.FailNow()
		}
	}
}

func Test(t *testing.T) {
	c := ansihtml.NewConverter(ansihtml.SetOptions(options))
	expect := newExpect(t, c)

	// simple
	expect("helloworld", "helloworld")
	expect("\x1b[9;31mhelloworld\x1b[0m", `<span style="color:#e05561;text-decoration:line-through">helloworld</span>`)

	// start with ESC
	expect("\x1b[30mhello\x1b[mworld", `<span style="color:#3f4451">hello</span>world`)

	// hyperlink
	expect("he\x1b[31mllo\x1b]8;id=app;http://example.com\x1b\\This is \x1b]8;id=app:rel=noopener noreferrer;http://example.com\x1b\\a \x1b[34mli\x1b[34mnk\x1b]8;;\x1b\\world\x1b[m",
		`he<span style="color:#e05561">llo</span><a href="http://example.com" class="ansi-link" id="app"><span style="color:#e05561">This is </span></a><a href="http://example.com" class="ansi-link" id="app" rel="noopener noreferrer"><span style="color:#e05561">a </span><span style="color:#4aa5f0">link</span></a><span style="color:#4aa5f0">world</span>`)

	// endurance failure
	expect("\x1b[31m\x1b[0;;31;mhelloworld\x1b[m", "helloworld")
	expect("hello\x1b[??2Jhelloworld\x1b[m", "hellohelloworld")
	expect("\x1b[35?35mhello\x1b[m", "hello")
	expect("\x1b[30$?!;;;;;hello\x1b[m", "ello")
	expect("hello\x1b[?,002J\x1b[m", "hello")
	expect("\x1b[31m\x1b[0;;;31mhelloworld\x1b[m", `<span style="color:#e05561">helloworld</span>`)
	expect("\x1b[31m\x1b[0;;31w;mhelloworld\x1b[m", `<span style="color:#e05561">;mhelloworld</span>`)
	expect("\x1b[38;5mhelloworld\x1b[m", `<span style="color:#3f4451">helloworld</span>`)
	expect("\x1b[38;5;mhelloworld\x1b[m", `<span style="color:#3f4451">helloworld</span>`)
	expect("\x1b[38;2;3;mblack\x1b[m", `<span style="color:#000000">black</span>`)
	expect("\x1b[48;5mhelloworld\x1b[m", `<span style="background-color:#3f4451">helloworld</span>`)
	expect("\x1b[48;5;mhelloworld\x1b[m", `<span style="background-color:#3f4451">helloworld</span>`)
	expect("\x1b[48;2;3;mblack\x1b[m", `<span style="background-color:#000000">black</span>`)
	expect("abcde\x1b[", "abcde")
	expect("abcde\x1b]", "abcde")
	expect("\x1b7helloworld", "helloworld")
	expect("\x1b[?25hhelloworld", "helloworld")
	expect("\x1b[?1049hhelloworld", "helloworld")
	expect("\x1b[20;3Hhelloworld", "helloworld")
	expect("abcde\x1b]6;id=app;http://example.com\x1b\\", "abcde")
	expect("\x1b]8;;;;http://example.com\x1b\\helloworld\x1b\\", "helloworld")
}

func TestIverse(t *testing.T) {
	c := ansihtml.NewConverter(ansihtml.SetTheme(ansihtml.Theme{Foreground: "#eee"}))
	expect := newExpect(t, c)
	expect("hello\x1b[7mworld\x1b[m", `<span style="color:#eeeeee">hello</span><span style="background-color:#eeeeee">world</span>`)
}

func TestDim(t *testing.T) {
	c := ansihtml.NewConverter(ansihtml.SetOptions(options))
	expect := newExpect(t, c)
	expect("hello\x1b[2mworld\x1b[m", `hello<span style="opacity:0.5">world</span>`)
	expect("hello\x1b[44;2mworld\x1b[m", `hello<span style="background-color:#4aa5f0">world</span>`)
	expect("hello\x1b[34;2mworld\x1b[m", `hello<span style="color:#4aa5f0;opacity:0.5">world</span>`)
	expect("hello\x1b[34;44;2mworld\x1b[m", `hello<span style="background-color:#4aa5f0;color:#4aa5f080">world</span>`)
}

func TestMinimumContrastRatio(t *testing.T) {
	c := ansihtml.NewConverter(ansihtml.SetOptions(options))
	expect := newExpect(t, c)
	expect("\x1b[31;41mhelloworld\x1b[m", `<span style="background-color:#e05561;color:#e05561">helloworld</span>`)
	c = ansihtml.NewConverter(ansihtml.SetMinimumContrastRatio(4.5))
	expect = newExpect(t, c)
	expect("\x1b[31;41mhelloworld\x1b[m", `<span style="background-color:#e05561;color:#ffffff">helloworld</span>`)
	expect("\x1b[107;92mhelloworld\x1b[m", `<span style="background-color:#d7dae0;color:#6b914b">helloworld</span>`)
}

func TestOtherInline(t *testing.T) {
	expect := newExpect(t, ansihtml.NewConverter())
	expect("\x1b[3;100mhelloworl\x1b[8md\x1b[m", `<span style="background-color:#4f5666;font-style:italic">helloworl</span><span style="background-color:#4f5666;font-style:italic;opacity:0">d</span>`)
	expect("\x1b[3;100;49mhelloworld\x1b[m", `<span style="font-style:italic">helloworld</span>`)
	expect("\x1b[48;2;3;4;5mhelloworld\x1b[m", `<span style="background-color:#030405">helloworld</span>`)
	expect("hello\x0bwo\x1bmrld\x1b[m", `helloworld`)
	expect("\x1b[38;5;2mhelloworld\x1b[m", `<span style="color:#8cc265">helloworld</span>`)
	expect("\x1b[38;5;2;1mhelloworld\x1b[m", `<span style="color:#a5e075;font-weight:700">helloworld</span>`)
	expect("\x1b[38;2;2;4;6mhelloworld\x1b[m", `<span style="color:#020406">helloworld</span>`)
	expect("\x1b]8;;http://example.com\x1b\\This is a link", `<a href="http://example.com" class="ansi-link">This is a link</a>`)
	expect("\x1b[2;31;41mhelloworld\x1b[m", `<span style="background-color:#e05561;color:#fcdfe380">helloworld</span>`)
	expect = newExpect(t, ansihtml.NewConverter(ansihtml.SetTheme(ansihtml.Theme{Foreground: "#eee"})))
	expect("\x1b[2;41mhelloworld\x1b[m", `<span style="background-color:#e05561;color:#eeeeee80">helloworld</span>`)
}

func TestOtherClass(t *testing.T) {
	c := ansihtml.NewConverter(ansihtml.SetMode(ansihtml.Class))
	expect := newExpect(t, c)
	expect("\x1b[7mhelloworld\x1b[m", `<span class="ansi-fg-inverse ansi-bg-inverse">helloworld</span>`)
	expect("\x1b[1;44;38;5;1mhelloworld\x1b[m", `<span class="ansi-fg-9 ansi-bg-4 ansi-bold">helloworld</span>`)
	expect("\x1b[2;3;4;5;6;7;8;9mhelloworld\x1b[m", `<span class="ansi-fg-inverse ansi-bg-inverse ansi-underline ansi-strike ansi-italic ansi-dim ansi-hidden">helloworld</span>`)
	expect("\x1b[2;31;48;2;255;240;103;38;2;2;2;2mhelloworld\x1b[m", `<span class="ansi-dim" style="background-color:#fff067;color:#020202">helloworld</span>`)
}

func TestError(t *testing.T) {
	c := ansihtml.NewConverter()
	expect := newExpectError(t, c)
	expect("\x1b[38;5;288mhelloworld288\x1b[m", ansihtml.ErrColorUndefined)
	expect("\x1b[48;5;728mhelloworld728\x1b[m", ansihtml.ErrColorUndefined)
}

func TestExample(t *testing.T) {
	f, err := os.Create("demo.html")
	if err != nil {
		return
	}
	defer f.Close()
	html, err := ansihtml.ToDemo("", ansihtml.SetTheme(ansihtml.Theme{
		Foreground: "#abb2bf",
		Background: "#23272e",
	}))
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = f.WriteString(html)
	if err != nil {
		return
	}
}

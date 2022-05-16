package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
)

// "[+-]?(\\d*[.])?\\d+"

var rgbaRe = regexp.MustCompile("(?:^rgba?\\(\\s*([+-]?(?:\\d*[.])?\\d+)\\s*,\\s*([+-]?(?:\\d*[.])?\\d+)\\s*,\\s*([+-]?(?:\\d*[.])?\\d+)\\s*(?:,\\s*([+-]?(?:\\d*[.])?\\d+)\\s*)?\\)$)")
var rgbaRe2 = regexp.MustCompile("(?:^rgba?\\(\\s*([+-]?(?:\\d*[.])?\\d+)\\s* \\s*([+-]?(?:\\d*[.])?\\d+)\\s* \\s*([+-]?(?:\\d*[.])?\\d+)\\s*(?:/\\s*([+-]?(?:\\d*[.])?\\d+)\\s*)?\\)$)")

func p(value string) {
	var result = rgbaRe2.FindStringSubmatch(value)
	fmt.Printf("%#v\n", result)
	if result != nil {
		t(result[1])
		t(result[2])
		t(result[3])
	}
}

func t(value string) {
	v, _ := strconv.ParseFloat(value, 64)
	v = math.Min(255.0, math.Max(0, v))
	fmt.Println(math.Round(v))
}

func main() {
	p("rgb(55, 129, 128)")
	p("rgb(  55  ,  129, 128  )")
	p("rgb(  -55  ,  +129.5, 128.5  )")
	p("rgb(55 129 128)")
	p("rgb(  55    129 128  )")
	p("rgb(  -55   +254.4 2128.5/1)")
}

# ansihtml

Convert ANSI escape codes to html

# Installation

```sh
go get -d github.com/lightyen/ansihtml
```

# Usage

```go
package main

import (
	"fmt"
	"strings"

	"github.com/lightyen/ansihtml"
)

func Example() {
	converter := ansihtml.NewConverter()
	var buf strings.Builder
	if err := converter.Copy(&buf, strings.NewReader("\x1b[38;2;66;66;66;44mhelloworld\x1b[m")); err != nil {
		return
	}
	fmt.Println(buf.String())

	// Output:
	// <span style="background-color:#4aa5f0;color:#424242">helloworld</span>
}
```

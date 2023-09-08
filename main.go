package main

import (
	"fmt"
	"gee"
)

func main() {
	r := gee.New()
	r.GET("/", func(ctx *gee.Context) {
		fmt.Fprintf(ctx.Writer, "URL.Path = %q\n", ctx.Req.URL.Path)
	})

	r.GET("/hello", func(ctx *gee.Context) {
		for k, v := range ctx.Req.Header {
			fmt.Fprintf(ctx.Writer, "Header[%q] = %q\n", k, v)
		}
	})

	r.GET("/:name", func(ctx *gee.Context) {
		fmt.Fprintf(ctx.Writer, "Hi "+ctx.Param("name"))
	})

	r.Run(":9999")
}

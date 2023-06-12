package main

import (
	"fmt"
	"gee"
	"net/http"
)

func main() {
	// 类似于gin框架
	// NEW 创建实例
	r := gee.New()
	// GET添加路由
	r.GET("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
	})
	r.GET("/hello", func(w http.ResponseWriter, req *http.Request) {
		for k, v := range req.Header {
			fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
		}
	})
	// RUN 启动
	r.Run(":9999")
}

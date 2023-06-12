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
	r.GET("/", func(c *gee.Context) {
		fmt.Fprintf(c.Writer, "URL.Path = %q\n", c.Req.URL.Path)
	})
	r.GET("/hello", func(c *gee.Context) {
		// expect /hello?name=geektutu
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
	})

	r.POST("/login", func(c *gee.Context) {
		fmt.Println("pause")
		c.JSON(http.StatusOK, gee.H{
			"username": c.PostForm("username"),
			"password": c.PostForm("password"),
		})
	})
	// RUN 启动
	r.Run(":9999")
}

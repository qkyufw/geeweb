package gee

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

// trace 显示错误路径
func trace(message string) string {
	var pcs [32]uintptr
	// 第0个是Callers本身，第一个是trace，再上一个是defer，因此跳过前三个
	n := runtime.Callers(3, pcs[:]) // skip first 3 caller

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)                           // 获取对应函数
		file, line := fn.FileLine(pc)                         // 获取文件名和行号
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line)) // 打印到日志中
	}
	return str.String()
}

// Recovery 回复程序
func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message)) // 打印错误信息
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()

		c.Next()
	}
}

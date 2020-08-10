/*
 @Desc

 @Date 2020-04-24 10:32
 @Author yinjk
*/
package http

import (
	"github.com/gin-gonic/gin"
	"testing"
)

func TestDefault(t *testing.T) {
	engine := Default()
	engine.GET("/test", func(context *gin.Context) {
		context.Writer.Header().Set("Cache-Control", "max-age=3600, must-revalidate")
		_, _ = context.Writer.WriteString("hello world!")
		return
	})
	engine.ListenAndStartUp()
}

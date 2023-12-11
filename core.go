package mygin

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

var engine = gin.New()

func GetApp() *gin.Engine                            { return engine }
func GetAppWithGroup(prefix string) *gin.RouterGroup { return GetApp().Group(prefix) }

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Allow", "POST, OPTIONS,DELETE,GET, PUT")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization,Access-Control-Allow-Origin,Access-Control-Allow-Headers,token, x-requested-with")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS,DELETE, GET, PUT")
		if c.Request.Method == http.MethodOptions || c.Request.Method == http.MethodHead {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	}
}

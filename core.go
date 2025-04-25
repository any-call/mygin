package mygin

import (
	"fmt"
	"github.com/any-call/gobase/util/mycrypto"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
	"sort"
	"strings"
)

var engine = gin.New()

func GetApp() *gin.Engine                            { return engine }
func GetAppWithGroup(prefix string) *gin.RouterGroup { return GetApp().Group(prefix) }

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Allow", "POST, OPTIONS,DELETE,GET, PUT")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin,Content-Type,Content-Length,Accept-Encoding,Authorization,X-CSRF-Token,Access-Control-Allow-Origin,Access-Control-Allow-Headers,token, x-requested-with")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS,DELETE, GET, PUT")
		if c.Request.Method == http.MethodOptions || c.Request.Method == http.MethodHead {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	}
}

func AuthSign(signKey string, secretKey string,
	getParamCb func(ctx *gin.Context) map[string]string,
	postParamCb func(ctx *gin.Context) map[string]string,
	putParamCb func(ctx *gin.Context) map[string]string,
	deleteParamCb func(ctx *gin.Context) map[string]string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取所有请求参数（支持 GET 和 POST）
		var paramMap = make(map[string]string, 0)
		var err error
		switch c.Request.Method {
		case "GET":
			if getParamCb != nil {
				paramMap = getParamCb(c)
			} else {
				err = c.ShouldBindQuery(&paramMap)
			}
			break
		case "POST":
			if postParamCb != nil {
				paramMap = postParamCb(c)
			} else {
				err = c.ShouldBindWith(&paramMap, binding.Form)
			}
			break
		case "PUT":
			if putParamCb != nil {
				paramMap = putParamCb(c)
			} else {
				err = c.ShouldBindWith(&paramMap, binding.Form)
			}
			break
		case "DELETE":
			if deleteParamCb != nil {
				paramMap = deleteParamCb(c)
			} else {
				err = c.ShouldBindBodyWith(&paramMap, FormUrlEncode)
			}
			break
		}

		if paramMap == nil || len(paramMap) == 0 || err != nil {
			c.Next()
			return
		}

		signValue, ok := paramMap[signKey]
		if !ok {
			c.AbortWithStatusJSON(403, gin.H{"error": "missing sign"})
			return
		}

		delete(paramMap, signValue)
		computedSign := GetAuthSign(paramMap, secretKey)
		if computedSign != signValue {
			c.AbortWithStatusJSON(403, gin.H{"error": "invalid sign"})
			return
		}

		// 验签通过，继续处理
		c.Next()
	}
}

// GenerateSign 接收参数 map 和密钥，生成签名字符串（MD5）
func GetAuthSign(params map[string]string, secretKey string) string {
	// 提取 key 并排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构造 key=value&key2=value2 的格式
	var builder strings.Builder
	for i, k := range keys {
		if i < len(keys)-1 {
			builder.WriteString(fmt.Sprintf("%s=%s&", k, params[k]))
		} else {
			builder.WriteString(fmt.Sprintf("%s=%s", k, params[k]))
		}
	}

	// 加上 secretKey
	rawString := builder.String() + secretKey
	return mycrypto.Md5(rawString)
}

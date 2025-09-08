package controller

import (
	"encoding/base64"
	"fmt"
	"frps-panel/pkg/server/model"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func (c *HandleController) BasicAuth() gin.HandlerFunc {
	return func(context *gin.Context) {
		session := sessions.Default(context)
		auth := session.Get(AuthName)
		userRole := session.Get(UserRoleName)

		// 如果未设置管理员账户，则直接放行
		if trimString(c.CommonInfo.AdminUser) == "" || trimString(c.CommonInfo.AdminPwd) == "" {
			if context.Request.RequestURI == LoginUrl {
				context.Redirect(http.StatusTemporaryRedirect, LoginSuccessUrl)
			}
			context.Next()
			return
		}

		if auth != nil && userRole != nil {
			if c.CommonInfo.AdminKeepTime > 0 {
				cookie, _ := context.Request.Cookie(SessionName)
				if cookie != nil {
					cookie.Expires = time.Now().Add(time.Second * time.Duration(c.CommonInfo.AdminKeepTime))
					http.SetCookie(context.Writer, cookie)
				}
			}

			// 验证管理员
			if userRole == UserRoleAdmin {
				username, password, _ := parseBasicAuth(fmt.Sprintf("%v", auth))
				usernameMatch := username == c.CommonInfo.AdminUser
				passwordMatch := password == c.CommonInfo.AdminPwd
				if usernameMatch && passwordMatch {
					context.Next()
					return
				}
			} else if userRole == UserRoleNormal { // 验证普通用户
				// 检查请求路径是否是用户仪表板
				if context.Request.RequestURI == UserDashboardUrl || strings.HasPrefix(context.Request.RequestURI, "/api/user/") {
					context.Next()
					return
				}
			}
		}

		isAjax := context.GetHeader("X-Requested-With") == "XMLHttpRequest"

		if !isAjax && context.Request.RequestURI != LoginUrl {
			context.Redirect(http.StatusTemporaryRedirect, LoginUrl)
		} else {
			context.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

func (c *HandleController) LoginAuth(username, password string, context *gin.Context) bool {
	session := sessions.Default(context)

	// 1. 尝试管理员登录
	adminAuth := encodeBasicAuth(c.CommonInfo.AdminUser, c.CommonInfo.AdminPwd)
	if encodeBasicAuth(username, password) == adminAuth {
		session.Set(AuthName, adminAuth)
		session.Set(UserRoleName, UserRoleAdmin)
		_ = session.Save()
		return true
	}

	// 2. 尝试普通用户登录
	if c.DB != nil {
		var user model.UserToken
		// Trim space from username before querying the database
		trimmedUser := strings.TrimSpace(username)
		// Use Unscoped to bypass the soft delete check for debugging
		result := c.DB.Where("user = ?", trimmedUser).First(&user)
		if result.Error == nil {
			// 找到了用户，现在验证密码和其他条件
			if user.Enable && user.Token == password {
				// 检查用户是否过期
				if user.ExpireDate != "" {
					expireTime, err := time.Parse("2006-01-02 15:04:05", user.ExpireDate)
					if err == nil && time.Now().After(expireTime) {
						return false // 用户已过期
					}
				}
				session.Set(AuthName, encodeBasicAuth(trimmedUser, password)) // 存储用户凭证
				session.Set(UserRoleName, UserRoleNormal)
				session.Set("current_user", trimmedUser) // 存储当前登录的普通用户
				_ = session.Save()
				return true
			}
		}
	}

	// 登录失败，清除会话
	ClearAuth(context)
	return false
}

func ClearAuth(context *gin.Context) {
	session := sessions.Default(context)
	session.Delete(AuthName)
	session.Delete(UserRoleName)
	session.Delete("current_user")
	_ = session.Save()
}

func parseBasicAuth(auth string) (username, password string, ok bool) {
	c, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		return "", "", false
	}
	cs := string(c)
	username, password, ok = strings.Cut(cs, ":")
	if !ok {
		return "", "", false
	}
	return username, password, true
}

func encodeBasicAuth(username, password string) string {
	authString := fmt.Sprintf("%s:%s", username, password)
	return base64.StdEncoding.EncodeToString([]byte(authString))
}

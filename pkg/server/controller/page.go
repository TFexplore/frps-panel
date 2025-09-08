package controller

import (
	"fmt"
	"net/http"

	ginI18n "github.com/gin-contrib/i18n"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// 后台登录
func (c *HandleController) MakeLoginFunc() func(context *gin.Context) {
	return func(context *gin.Context) {
		if context.Request.Method == "GET" {
			session := sessions.Default(context)
			userRole := session.Get(UserRoleName)

			if userRole == UserRoleAdmin {
				context.Redirect(http.StatusTemporaryRedirect, LoginSuccessUrl)
				return
			} else if userRole == UserRoleNormal {
				context.Redirect(http.StatusTemporaryRedirect, UserDashboardUrl)
				return
			}

			context.HTML(http.StatusOK, "login.html", gin.H{
				"version":             c.Version,
				"FrpsPanel":           ginI18n.MustGetMessage(context, "Frps Panel"),
				"Username":            ginI18n.MustGetMessage(context, "Username"),
				"Password":            ginI18n.MustGetMessage(context, "Password"),
				"Login":               ginI18n.MustGetMessage(context, "Login"),
				"PleaseInputUsername": ginI18n.MustGetMessage(context, "Please input username"),
				"PleaseInputPassword": ginI18n.MustGetMessage(context, "Please input password"),
			})
		} else if context.Request.Method == "POST" {
			username := context.PostForm("username")
			password := context.PostForm("password")
			if c.LoginAuth(username, password, context) {
				session := sessions.Default(context)
				userRole := session.Get(UserRoleName)
				redirectUrl := LoginSuccessUrl
				if userRole == UserRoleNormal {
					redirectUrl = UserDashboardUrl
				}
				context.JSON(http.StatusOK, gin.H{
					"success":  true,
					"message":  ginI18n.MustGetMessage(context, "Login success"),
					"redirect": redirectUrl,
				})
			} else {
				context.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": ginI18n.MustGetMessage(context, "Username or password incorrect"),
				})
			}
		}
	}
}

// 后台登出
func (c *HandleController) MakeLogoutFunc() func(context *gin.Context) {
	return func(context *gin.Context) {
		ClearAuth(context)
		context.Redirect(http.StatusTemporaryRedirect, LogoutSuccessUrl)
	}
}

// 获取frps服务器后台面板信息
func (c *HandleController) MakeIndexFunc() func(context *gin.Context) {
	return func(context *gin.Context) {
		session := sessions.Default(context)
		userRole := session.Get(UserRoleName)

		if userRole != UserRoleAdmin {
			context.Redirect(http.StatusTemporaryRedirect, UserDashboardUrl)
			return
		}

		context.HTML(http.StatusOK, "index.html", gin.H{
			"version":                      c.Version,
			"showExit":                     trimString(c.CommonInfo.AdminUser) != "" && trimString(c.CommonInfo.AdminPwd) != "",
			"FrpsPanel":                    ginI18n.MustGetMessage(context, "Frps Panel"),
			"User":                         ginI18n.MustGetMessage(context, "User"),
			"Token":                        ginI18n.MustGetMessage(context, "Token"),
			"Notes":                        ginI18n.MustGetMessage(context, "Notes"),
			"Search":                       ginI18n.MustGetMessage(context, "Search"),
			"Reset":                        ginI18n.MustGetMessage(context, "Reset"),
			"NewUser":                      ginI18n.MustGetMessage(context, "New user"),
			"RemoveUser":                   ginI18n.MustGetMessage(context, "Remove user"),
			"DisableUser":                  ginI18n.MustGetMessage(context, "Disable user"),
			"EnableUser":                   ginI18n.MustGetMessage(context, "Enable user"),
			"Remove":                       ginI18n.MustGetMessage(context, "Remove"),
			"Enable":                       ginI18n.MustGetMessage(context, "Enable"),
			"Disable":                      ginI18n.MustGetMessage(context, "Disable"),
			"PleaseInputUserAccount":       ginI18n.MustGetMessage(context, "Please input user account"),
			"PleaseInputUserToken":         ginI18n.MustGetMessage(context, "Please input user token"),
			"PleaseInputUserNotes":         ginI18n.MustGetMessage(context, "Please input user notes"),
			"AllowedPorts":                 ginI18n.MustGetMessage(context, "Allowed ports"),
			"PleaseInputAllowedPorts":      ginI18n.MustGetMessage(context, "Please input allowed ports"),
			"AllowedDomains":               ginI18n.MustGetMessage(context, "Allowed domains"),
			"PleaseInputAllowedDomains":    ginI18n.MustGetMessage(context, "Please input allowed domains"),
			"AllowedSubdomains":            ginI18n.MustGetMessage(context, "Allowed subdomains"),
			"PleaseInputAllowedSubdomains": ginI18n.MustGetMessage(context, "Please input allowed subdomains"),
			"NotLimit":                     ginI18n.MustGetMessage(context, "Not limit"),
			"None":                         ginI18n.MustGetMessage(context, "None"),
			"ServerInfo":                   ginI18n.MustGetMessage(context, "Server Info"),
			"Users":                        ginI18n.MustGetMessage(context, "Users"),
			"Proxies":                      ginI18n.MustGetMessage(context, "Proxies"),
			"TrafficStatistics":            ginI18n.MustGetMessage(context, "Traffic Statistics"),
			"Name":                         ginI18n.MustGetMessage(context, "Name"),
			"Type":                         ginI18n.MustGetMessage(context, "Type"),
			"Domains":                      ginI18n.MustGetMessage(context, "Domains"),
			"SubDomain":                    ginI18n.MustGetMessage(context, "SubDomain"),
			"Locations":                    ginI18n.MustGetMessage(context, "Locations"),
			"HostRewrite":                  ginI18n.MustGetMessage(context, "HostRewrite"),
			"Encryption":                   ginI18n.MustGetMessage(context, "Encryption"),
			"Compression":                  ginI18n.MustGetMessage(context, "Compression"),
			"Addr":                         ginI18n.MustGetMessage(context, "Addr"),
			"LastStart":                    ginI18n.MustGetMessage(context, "Last Start"),
			"LastClose":                    ginI18n.MustGetMessage(context, "Last Close"),
			"Version":                      ginI18n.MustGetMessage(context, "Version"),
			"BindPort":                     ginI18n.MustGetMessage(context, "Bind Port"),
			"KCPBindPort":                  ginI18n.MustGetMessage(context, "KCP Bind Port"),
			"QUICBindPort":                 ginI18n.MustGetMessage(context, "QUIC Bind Port"),
			"HTTPPort":                     ginI18n.MustGetMessage(context, "HTTP Port"),
			"HTTPSPort":                    ginI18n.MustGetMessage(context, "HTTPS Port"),
			"TCPMUXPort":                   ginI18n.MustGetMessage(context, "TCPMUX Port"),
			"SubdomainHost":                ginI18n.MustGetMessage(context, "Subdomain Host"),
			"MaxPoolCount":                 ginI18n.MustGetMessage(context, "Max Pool Count"),
			"MaxPortsPerClient":            ginI18n.MustGetMessage(context, "Max Ports Per Client"),
			"HeartBeatTimeout":             ginI18n.MustGetMessage(context, "Heart Beat Timeout"),
			"AllowPorts":                   ginI18n.MustGetMessage(context, "Allow Ports"),
			"TLSOnly":                      ginI18n.MustGetMessage(context, "TLS Only"),
			"CurrentConnections":           ginI18n.MustGetMessage(context, "Current Connections"),
			"ClientCounts":                 ginI18n.MustGetMessage(context, "Client Counts"),
			"ProxyCounts":                  ginI18n.MustGetMessage(context, "Proxy Counts"),
			"Server":                       ginI18n.MustGetMessage(context, "Server"),                    // 新增
			"CreateDate":                   ginI18n.MustGetMessage(context, "Create Date"),               // 新增
			"ExpireDate":                   ginI18n.MustGetMessage(context, "Expire Date"),               // 新增
			"PleaseInputServerName":        ginI18n.MustGetMessage(context, "Please input server name"),  // 新增
			"PleaseSelectExpireDate":       ginI18n.MustGetMessage(context, "Please select expire date"), // 新增
			"AllServers":                   ginI18n.MustGetMessage(context, "All Servers"),               // 新增
			"ConfigTemplate":               ginI18n.MustGetMessage(context, "ConfigTemplate"),            // 新增
			"PleaseInputConfigTemplate":    ginI18n.MustGetMessage(context, "PleaseInputConfigTemplate"), // 新增
			"ExportConfig":                 ginI18n.MustGetMessage(context, "ExportConfig"),              // 新增
			"EditConfigTemplate":           ginI18n.MustGetMessage(context, "EditConfigTemplate"),        // 新增
		})
	}
}

func (c *HandleController) MakeUserDashboardFunc() func(context *gin.Context) {
	return func(context *gin.Context) {
		session := sessions.Default(context)
		currentUser := session.Get("current_user")
		if currentUser == nil {
			context.Redirect(http.StatusTemporaryRedirect, LoginUrl)
			return
		}

		context.HTML(http.StatusOK, "user_dashboard.html", gin.H{
			"version":           c.Version,
			"FrpsPanel":         ginI18n.MustGetMessage(context, "Frps Panel"),
			"User":              fmt.Sprintf("%v", currentUser),
			"Logout":            ginI18n.MustGetMessage(context, "Logout"),
			"MyInfo":            ginI18n.MustGetMessage(context, "My Info"),
			"MyProxies":         ginI18n.MustGetMessage(context, "My Proxies"),
			"Name":              ginI18n.MustGetMessage(context, "Name"),
			"Type":              ginI18n.MustGetMessage(context, "Type"),
			"Domains":           ginI18n.MustGetMessage(context, "Domains"),
			"SubDomain":         ginI18n.MustGetMessage(context, "SubDomain"),
			"Locations":         ginI18n.MustGetMessage(context, "Locations"),
			"HostRewrite":       ginI18n.MustGetMessage(context, "HostRewrite"),
			"Encryption":        ginI18n.MustGetMessage(context, "Encryption"),
			"Compression":       ginI18n.MustGetMessage(context, "Compression"),
			"Addr":              ginI18n.MustGetMessage(context, "Addr"),
			"LastStart":         ginI18n.MustGetMessage(context, "Last Start"),
			"LastClose":         ginI18n.MustGetMessage(context, "Last Close"),
			"Connections":       ginI18n.MustGetMessage(context, "Connections"),
			"TrafficIn":         ginI18n.MustGetMessage(context, "Traffic In"),
			"TrafficOut":        ginI18n.MustGetMessage(context, "Traffic Out"),
			"ClientVersion":     ginI18n.MustGetMessage(context, "Client Version"),
			"TrafficStatistics": ginI18n.MustGetMessage(context, "Traffic Statistics"),
			"online":            ginI18n.MustGetMessage(context, "online"),
			"offline":           ginI18n.MustGetMessage(context, "offline"),
			"Total":             ginI18n.MustGetMessage(context, "Total"),
			"Items":             ginI18n.MustGetMessage(context, "Items"),
			"Goto":              ginI18n.MustGetMessage(context, "Go to"),
			"PerPage":           ginI18n.MustGetMessage(context, "Per Page"),
			"NotSet":            ginI18n.MustGetMessage(context, "Not Set"),
			"Proxy":             ginI18n.MustGetMessage(context, "Proxy"),
			"Token":             ginI18n.MustGetMessage(context, "Token"),
			"Notes":             ginI18n.MustGetMessage(context, "Notes"),
			"AllowedPorts":      ginI18n.MustGetMessage(context, "Allowed ports"),
			"AllowedDomains":    ginI18n.MustGetMessage(context, "Allowed domains"),
			"AllowedSubdomains": ginI18n.MustGetMessage(context, "Allowed subdomains"),
			"NotLimit":          ginI18n.MustGetMessage(context, "Not limit"),
			"None":              ginI18n.MustGetMessage(context, "None"),
			"CreateDate":        ginI18n.MustGetMessage(context, "Create Date"),
			"ExpireDate":        ginI18n.MustGetMessage(context, "Expire Date"),
		})
	}
}

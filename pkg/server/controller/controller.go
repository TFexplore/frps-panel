package controller

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"frps-panel/pkg/server/model"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	plugin "github.com/fatedier/frp/pkg/plugin/server"
	ginI18n "github.com/gin-contrib/i18n"
	"github.com/gin-contrib/sessions" // 导入 sessions 包
	"github.com/gin-gonic/gin"
)

// frps连接处理
func (c *HandleController) MakeHandlerFunc() gin.HandlerFunc {
	return func(context *gin.Context) {
		var response plugin.Response
		var err error

		request := plugin.Request{}
		if err := context.BindJSON(&request); err != nil {
			_ = context.AbortWithError(http.StatusBadRequest, err)
			return
		}

		jsonStr, err := json.Marshal(request.Content)
		if err != nil {
			_ = context.AbortWithError(http.StatusBadRequest, err)
			return
		}

		if request.Op == "Login" {
			content := plugin.LoginContent{}
			err = json.Unmarshal(jsonStr, &content)
			response = c.HandleLogin(&content, context.ClientIP())
		} else if request.Op == "NewProxy" {
			content := plugin.NewProxyContent{}
			err = json.Unmarshal(jsonStr, &content)
			response = c.HandleNewProxy(&content)
		} else if request.Op == "Ping" {
			content := plugin.PingContent{}
			err = json.Unmarshal(jsonStr, &content)
			response = c.HandlePing(&content)
		} else if request.Op == "NewWorkConn" {
			content := plugin.NewWorkConnContent{}
			err = json.Unmarshal(jsonStr, &content)
			response = c.HandleNewWorkConn(&content)
		} else if request.Op == "NewUserConn" {
			content := plugin.NewUserConnContent{}
			err = json.Unmarshal(jsonStr, &content)
			response = c.HandleNewUserConn(&content)
		}

		if err != nil {
			log.Printf("handle %s error: %v", context.Request.URL.Path, err)
			var e *HTTPError
			switch {
			case errors.As(err, &e):
				context.JSON(e.Code, &Response{Msg: e.Err.Error()})
			default:
				context.JSON(http.StatusInternalServerError, &Response{Msg: err.Error()})
			}
			return
		} else {
			resStr, _ := json.Marshal(response)
			log.Printf("handle:%v , result: %v", request.Op, string(resStr))
		}

		context.JSON(http.StatusOK, response)
	}
}

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

// 语言相关
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

// 语言相关
func (c *HandleController) MakeLangFunc() func(context *gin.Context) {
	return func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"User":                  ginI18n.MustGetMessage(context, "User"),
			"Token":                 ginI18n.MustGetMessage(context, "Token"),
			"Server":                ginI18n.MustGetMessage(context, "Server"),
			"CreateDate":            ginI18n.MustGetMessage(context, "Create Date"),
			"ExpireDate":            ginI18n.MustGetMessage(context, "Expire Date"),
			"Notes":                 ginI18n.MustGetMessage(context, "Notes"),
			"Status":                ginI18n.MustGetMessage(context, "Status"),
			"Operation":             ginI18n.MustGetMessage(context, "Operation"),
			"Enable":                ginI18n.MustGetMessage(context, "Enable"),
			"Disable":               ginI18n.MustGetMessage(context, "Disable"),
			"NewUser":               ginI18n.MustGetMessage(context, "New user"),
			"Confirm":               ginI18n.MustGetMessage(context, "Confirm"),
			"Cancel":                ginI18n.MustGetMessage(context, "Cancel"),
			"RemoveUser":            ginI18n.MustGetMessage(context, "Remove user"),
			"DisableUser":           ginI18n.MustGetMessage(context, "Disable user"),
			"ConfirmRemoveUser":     ginI18n.MustGetMessage(context, "Confirm to remove user"),
			"ConfirmDisableUser":    ginI18n.MustGetMessage(context, "Confirm to disable user"),
			"TakeTimeMakeEffective": ginI18n.MustGetMessage(context, "will take sometime to make effective"),
			"ConfirmEnableUser":     ginI18n.MustGetMessage(context, "Confirm to enable user"),
			"OperateSuccess":        ginI18n.MustGetMessage(context, "Operate success"),
			"OperateError":          ginI18n.MustGetMessage(context, "Operate error"),
			"OperateFailed":         ginI18n.MustGetMessage(context, "Operate failed"),
			"UserExist":             ginI18n.MustGetMessage(context, "User exist"),
			"UserNotExist":          ginI18n.MustGetMessage(context, "User not exist"),
			"UserFormatError":       ginI18n.MustGetMessage(context, "User format error"),
			"TokenFormatError":      ginI18n.MustGetMessage(context, "Token format error"),
			"ShouldCheckUser":       ginI18n.MustGetMessage(context, "Please check at least one user"),
			"OperationConfirm":      ginI18n.MustGetMessage(context, "Operation confirm"),
			"EmptyData":             ginI18n.MustGetMessage(context, "Empty data"),
			"NotLimit":              ginI18n.MustGetMessage(context, "Not limit"),
			"AllowedPorts":          ginI18n.MustGetMessage(context, "Allowed ports"),
			"AllowedDomains":        ginI18n.MustGetMessage(context, "Allowed domains"),
			"AllowedSubdomains":     ginI18n.MustGetMessage(context, "Allowed subdomains"),
			"PortsInvalid":          ginI18n.MustGetMessage(context, "Ports is invalid"),
			"DomainsInvalid":        ginI18n.MustGetMessage(context, "Domains is invalid"),
			"SubdomainsInvalid":     ginI18n.MustGetMessage(context, "Subdomains is invalid"),
			"CommentInvalid":        ginI18n.MustGetMessage(context, "Comment is invalid"),
			"ParamError":            ginI18n.MustGetMessage(context, "Param error"),
			"OtherError":            ginI18n.MustGetMessage(context, "Other error"),
			"Name":                  ginI18n.MustGetMessage(context, "Name"),
			"Port":                  ginI18n.MustGetMessage(context, "Port"),
			"Connections":           ginI18n.MustGetMessage(context, "Connections"),
			"TrafficIn":             ginI18n.MustGetMessage(context, "Traffic In"),
			"TrafficOut":            ginI18n.MustGetMessage(context, "Traffic Out"),
			"ClientVersion":         ginI18n.MustGetMessage(context, "Client Version"),
			"TrafficStatistics":     ginI18n.MustGetMessage(context, "Traffic Statistics"),
			"online":                ginI18n.MustGetMessage(context, "online"),
			"offline":               ginI18n.MustGetMessage(context, "offline"),
			"true":                  ginI18n.MustGetMessage(context, "true"),
			"false":                 ginI18n.MustGetMessage(context, "false"),
			"NetworkTraffic":        ginI18n.MustGetMessage(context, "Network Traffic"),
			"today":                 ginI18n.MustGetMessage(context, "today"),
			"now":                   ginI18n.MustGetMessage(context, "now"),
			"Proxies":               ginI18n.MustGetMessage(context, "Proxies"),
			"NotSet":                ginI18n.MustGetMessage(context, "Not Set"),
			"Proxy":                 ginI18n.MustGetMessage(context, "Proxy"),
			"TokenInvalid":          ginI18n.MustGetMessage(context, "Token invalid"),
			"Total":                 ginI18n.MustGetMessage(context, "Total"),
			"Items":                 ginI18n.MustGetMessage(context, "Items"),
			"Goto":                  ginI18n.MustGetMessage(context, "Go to"),
			"PerPage":               ginI18n.MustGetMessage(context, "Per Page"),
			"ConfigTemplate":        ginI18n.MustGetMessage(context, "ConfigTemplate"),       // 新增
			"PortCount":             ginI18n.MustGetMessage(context, "PortCount"),            // 新增
			"PleaseInputPortCount":  ginI18n.MustGetMessage(context, "PleaseInputPortCount"), // 新增
		})
	}
}

// 后台获取最大端口
func (c *HandleController) MakeGetMaxPortFunc() func(context *gin.Context) {
	return func(context *gin.Context) {
		serverName := context.Query("server")
		if serverName == "" {
			context.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Server name is required",
			})
			return
		}

		maxPort := 0
		for _, tokenInfo := range c.Tokens {
			if tokenInfo.Server == serverName {
				for _, p := range tokenInfo.Ports {
					switch v := p.(type) {
					case int:
						if v > maxPort {
							maxPort = v
						}
					case string:
						// 处理 "10000-10200" 格式的端口范围
						parts := strings.Split(v, "-")
						if len(parts) == 2 {
							endPort, err := strconv.Atoi(strings.TrimSpace(parts[1]))
							if err == nil && endPort > maxPort {
								maxPort = endPort
							}
						} else {
							// 如果是单个端口号的字符串形式，例如 "8080"
							singlePort, err := strconv.Atoi(strings.TrimSpace(v))
							if err == nil && singlePort > maxPort {
								maxPort = singlePort
							}
						}
					}
				}
			}
		}

		context.JSON(http.StatusOK, gin.H{
			"success": true,
			"maxPort": maxPort,
			"message": "Get max port success",
		})
	}
}

// 后台获取最大端口列表
func (c *HandleController) MakeGetAllMaxPortsFunc() func(context *gin.Context) {
	return func(context *gin.Context) {
		maxPortsMap := make(map[string]int)

		for _, tokenInfo := range c.Tokens {
			serverName := tokenInfo.Server
			if _, ok := maxPortsMap[serverName]; !ok {
				maxPortsMap[serverName] = 0
			}

			for _, p := range tokenInfo.Ports {
				switch v := p.(type) {
				case int:
					if v > maxPortsMap[serverName] {
						maxPortsMap[serverName] = v
					}
				case string:
					parts := strings.Split(v, "-")
					if len(parts) == 2 {
						endPort, err := strconv.Atoi(strings.TrimSpace(parts[1]))
						if err == nil && endPort > maxPortsMap[serverName] {
							maxPortsMap[serverName] = endPort
						}
					} else {
						singlePort, err := strconv.Atoi(strings.TrimSpace(v))
						if err == nil && singlePort > maxPortsMap[serverName] {
							maxPortsMap[serverName] = singlePort
						}
					}
				}
			}
		}

		context.JSON(http.StatusOK, gin.H{
			"success":     true,
			"maxPortsMap": maxPortsMap,
			"message":     "Get all max ports success",
		})
	}
}

// 查询用户列表
func (c *HandleController) MakeQueryUserInfoFunc() func(context *gin.Context) {
	return func(context *gin.Context) {
		session := sessions.Default(context)
		currentUser := fmt.Sprintf("%v", session.Get("current_user"))

		if currentUser == "" {
			context.JSON(http.StatusUnauthorized, &TokenResponse{
				Code:  ParamError,
				Msg:   "User not logged in",
				Count: 0,
				Data:  []TokenInfo{},
			})
			return
		}

		tokenInfo, ok := c.Tokens[currentUser]
		if !ok {
			context.JSON(http.StatusNotFound, &TokenResponse{
				Code:  UserNotExist,
				Msg:   "User info not found",
				Count: 0,
				Data:  []TokenInfo{},
			})
			return
		}

		context.JSON(http.StatusOK, &TokenResponse{
			Code:  0,
			Msg:   "query user info success",
			Count: 1,
			Data:  []TokenInfo{tokenInfo},
		})
	}
}

// 用户查询自己的连接代理信息
func (c *HandleController) MakeQueryUserProxiesFunc() func(context *gin.Context) {
	return func(context *gin.Context) {
		session := sessions.Default(context)
		currentUser := fmt.Sprintf("%v", session.Get("current_user"))

		if currentUser == "" {
			context.JSON(http.StatusUnauthorized, gin.H{
				"code":  ParamError,
				"msg":   "User not logged in",
				"count": 0,
				"data":  []interface{}{},
			})
			return
		}

		// 调用 MakeProxyFunc 的核心逻辑来获取所有代理信息
		// 这里需要模拟 MakeProxyFunc 的行为，或者重构 MakeProxyFunc 使其可复用
		// 为了简化，我们直接在这里实现代理查询和过滤逻辑
		var client *http.Client
		var protocol string

		if c.CurrentDashboardIndex >= len(c.Dashboards) {
			context.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "No dashboard configured or invalid current index",
			})
			return
		}

		currentDashboard := c.Dashboards[c.CurrentDashboardIndex]

		if currentDashboard.DashboardTls {
			client = &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}
			protocol = "https://"
		} else {
			client = http.DefaultClient
			protocol = "http://"
		}

		res := ProxyResponse{}
		host := currentDashboard.DashboardAddr
		port := currentDashboard.DashboardPort

		host, _ = strings.CutPrefix(host, protocol)

		// 从查询参数中获取 proxyType，如果未提供则默认为 "http"
		proxyType := context.DefaultQuery("proxyType", "http")
		// 构建请求 frps 的代理列表 API
		requestUrl := protocol + host + ":" + strconv.Itoa(port) + "/api/proxy/" + proxyType
		request, _ := http.NewRequest("GET", requestUrl, nil)
		username := currentDashboard.DashboardUser
		password := currentDashboard.DashboardPwd
		if trimString(username) != "" && trimString(password) != "" {
			request.SetBasicAuth(username, password)
			log.Printf("Proxy to %s", requestUrl)
		}

		response, err := client.Do(request)

		if err != nil {
			res.Code = FrpServerError
			res.Success = false
			res.Message = err.Error()
			log.Print(err)
			context.JSON(http.StatusOK, &res)
			return
		}

		res.Code = response.StatusCode
		body, err := io.ReadAll(response.Body)

		if err != nil {
			res.Success = false
			res.Message = err.Error()
		} else {
			if res.Code == http.StatusOK {
				var frpsProxies struct {
					Proxies []struct {
						Name          string `json:"name"`
						Type          string `json:"type"`
						Status        string `json:"status"`
						Connections   int    `json:"curConns"`
						TrafficIn     int64  `json:"todayTrafficIn"`
						TrafficOut    int64  `json:"todayTrafficOut"`
						ClientVersion string `json:"clientVersion"`
						LastStart     string `json:"lastStartTime"`
						LastClose     string `json:"lastCloseTime"`
						Conf          struct {
							RemotePort int `json:"remotePort"`
							Transport  struct {
								UseEncryption  bool `json:"useEncryption"`
								UseCompression bool `json:"useCompression"`
							} `json:"transport"`
						} `json:"conf"`
					} `json:"proxies"`
				}
				if err := json.Unmarshal(body, &frpsProxies); err != nil {
					res.Success = false
					res.Message = fmt.Sprintf("Failed to parse frps proxy response: %v", err)
					context.JSON(http.StatusOK, &res)
					return
				}

				var userProxies []gin.H
				for _, proxy := range frpsProxies.Proxies {
					if strings.HasPrefix(proxy.Name, currentUser) {
						userProxies = append(userProxies, gin.H{
							"Name":           proxy.Name,
							"Type":           proxy.Type,
							"Status":         proxy.Status,
							"Connections":    proxy.Connections,
							"TrafficIn":      proxy.TrafficIn,
							"TrafficOut":     proxy.TrafficOut,
							"ClientVersion":  proxy.ClientVersion,
							"LastStart":      proxy.LastStart,
							"LastClose":      proxy.LastClose,
							"RemotePort":     proxy.Conf.RemotePort,
							"UseEncryption":  proxy.Conf.Transport.UseEncryption,
							"UseCompression": proxy.Conf.Transport.UseCompression,
						})
					}
				}

				context.JSON(http.StatusOK, gin.H{
					"code":  0,
					"msg":   "query user proxies success",
					"count": len(userProxies),
					"data":  userProxies,
				})
				return
			} else {
				res.Success = false
				if res.Code == http.StatusNotFound {
					res.Message = fmt.Sprintf("Proxy to %s error: url not found", requestUrl)
				} else {
					res.Message = fmt.Sprintf("Proxy to %s error: %s", requestUrl, string(body))
				}
			}
		}
		log.Printf(res.Message)
		context.JSON(http.StatusOK, &res)
	}
}

// 管理员查询用户信息
func (c *HandleController) MakeQueryTokensFunc() func(context *gin.Context) {
	return func(context *gin.Context) {
		session := sessions.Default(context)
		userRole := session.Get(UserRoleName)
		currentUser := fmt.Sprintf("%v", session.Get("current_user"))

		search := TokenSearch{}
		search.Limit = 0

		err := context.BindQuery(&search)
		if err != nil {
			return
		}

		var tokenList []TokenInfo
		for _, tokenInfo := range c.Tokens {
			// 如果是普通用户，只显示自己的信息
			if userRole == UserRoleNormal && tokenInfo.User != currentUser {
				continue
			}
			tokenList = append(tokenList, tokenInfo)
		}
		sort.Slice(tokenList, func(i, j int) bool {
			return strings.Compare(tokenList[i].User, tokenList[j].User) < 0
		})

		var filtered []TokenInfo
		for _, tokenInfo := range tokenList {
			// 添加服务器名称过滤
			if search.Server != "" && tokenInfo.Server != search.Server {
				continue
			}
			// 如果是普通用户，强制过滤为当前用户
			if userRole == UserRoleNormal {
				if tokenInfo.User == currentUser && filter(tokenInfo, search.TokenInfo) {
					filtered = append(filtered, tokenInfo)
				}
			} else { // 管理员用户
				if filter(tokenInfo, search.TokenInfo) {
					filtered = append(filtered, tokenInfo)
				}
			}
		}
		if filtered == nil {
			filtered = []TokenInfo{}
		}

		count := len(filtered)
		if search.Limit > 0 {
			start := max((search.Page-1)*search.Limit, 0)
			end := min(search.Page*search.Limit, len(filtered))
			filtered = filtered[start:end]
		}
		context.JSON(http.StatusOK, &TokenResponse{
			Code:  0,
			Msg:   "query Tokens success",
			Count: count,
			Data:  filtered,
		})
	}
}

func (c *HandleController) MakeAddTokenFunc() func(context *gin.Context) {
	return func(context *gin.Context) {
		info := TokenInfo{
			Enable: true,
		}
		response := OperationResponse{
			Success: true,
			Code:    Success,
			Message: "user add success",
		}
		err := context.BindJSON(&info)
		if err != nil {
			response.Success = false
			response.Code = ParamError
			response.Message = fmt.Sprintf("user add failed, param error : %v", err)
			log.Printf(response.Message)
			context.JSON(http.StatusOK, &response)
			return
		}

		// 自动设置创建日期
		loc, _ := time.LoadLocation("Asia/Shanghai")
		info.CreateDate = time.Now().In(loc).Format("2006-01-02 15:04:05")

		result := c.verifyToken(info, TOKEN_ADD)

		if !result.Success {
			context.JSON(http.StatusOK, &result)
			return
		}

		info.Comment = cleanString(info.Comment)
		info.Ports = cleanPorts(info.Ports)
		info.Domains = cleanStrings(info.Domains)
		info.Subdomains = cleanStrings(info.Subdomains)
		info.Server = cleanString(info.Server)         // 清理服务器名称
		info.ExpireDate = cleanString(info.ExpireDate) // 清理到期时间

		// Save to database or file
		if c.DB != nil {
			userToken, err := FromTokenInfo(info)
			if err != nil {
				response.Success = false
				response.Code = SaveError
				response.Message = fmt.Sprintf("user add failed, data conversion error : %v", err)
				log.Printf(response.Message)
				context.JSON(http.StatusOK, &response)
				return
			}
			result := c.DB.Create(&userToken)
			if result.Error != nil {
				response.Success = false
				response.Code = SaveError
				response.Message = fmt.Sprintf("user add failed, db error : %v", result.Error)
				log.Printf(response.Message)
				context.JSON(http.StatusOK, &response)
				return
			}
		}

		c.Tokens[info.User] = info

		context.JSON(0, &response)
	}
}

func (c *HandleController) MakeUpdateTokensFunc() func(context *gin.Context) {
	return func(context *gin.Context) {
		response := OperationResponse{
			Success: true,
			Code:    Success,
			Message: "user update success",
		}
		update := TokenUpdate{}
		err := context.BindJSON(&update)
		if err != nil {
			response.Success = false
			response.Code = ParamError
			response.Message = fmt.Sprintf("update failed, param error : %v", err)
			log.Printf(response.Message)
			context.JSON(http.StatusOK, &response)
			return
			// 导入 time 包
		}

		before := update.Before
		after := update.After

		if before.User != after.User {
			response.Success = false
			response.Code = ParamError
			response.Message = fmt.Sprintf("update failed, user should be same : before -> %v, after -> %v", before.User, after.User)
			log.Printf(response.Message)
			context.JSON(http.StatusOK, &response)
			return
		}

		result := c.verifyToken(after, TOKEN_UPDATE)

		if !result.Success {
			context.JSON(http.StatusOK, &result)
			return
		}

		after.Comment = cleanString(after.Comment)
		after.Ports = cleanPorts(after.Ports)
		after.Domains = cleanStrings(after.Domains)
		after.Subdomains = cleanStrings(after.Subdomains)
		after.Server = cleanString(after.Server)         // 清理服务器名称
		after.ExpireDate = cleanString(after.ExpireDate) // 清理到期时间
		after.CreateDate = before.CreateDate             // 创建日期不应改变

		// Save to database or file
		if c.DB != nil {
			userToken, err := FromTokenInfo(after)
			if err != nil {
				response.Success = false
				response.Code = SaveError
				response.Message = fmt.Sprintf("user update failed, data conversion error : %v", err)
				log.Printf(response.Message)
				context.JSON(http.StatusOK, &response)
				return
			}
			result := c.DB.Save(&userToken)
			if result.Error != nil {
				response.Success = false
				response.Code = SaveError
				response.Message = fmt.Sprintf("user update failed, db error : %v", result.Error)
				log.Printf(response.Message)
				context.JSON(http.StatusOK, &response)
				return
			}
		}
		c.Tokens[after.User] = after

		context.JSON(http.StatusOK, &response)
	}
}

func (c *HandleController) MakeRemoveTokensFunc() func(context *gin.Context) {
	return func(context *gin.Context) {
		response := OperationResponse{
			Success: true,
			Code:    Success,
			Message: "user remove success",
		}
		remove := TokenRemove{}
		err := context.BindJSON(&remove)
		if err != nil {
			response.Success = false
			response.Code = ParamError
			response.Message = fmt.Sprintf("user remove failed, param error : %v", err)
			log.Printf(response.Message)
			context.JSON(http.StatusOK, &response)
			return
		}

		for _, user := range remove.Users {
			result := c.verifyToken(user, TOKEN_REMOVE)

			if !result.Success {
				context.JSON(http.StatusOK, &result)
				return
			}
		}

		for _, user := range remove.Users {
			if c.DB != nil {
				result := c.DB.Delete(&model.UserToken{}, "user = ?", user.User)
				if result.Error != nil {
					response.Success = false
					response.Code = SaveError
					response.Message = fmt.Sprintf("user remove failed for %s, db error : %v", user.User, result.Error)
					log.Printf(response.Message)
					context.JSON(http.StatusOK, &response)
					return
				}
			}
			delete(c.Tokens, user.User)
		}

		if c.DB == nil {
		}

		context.JSON(http.StatusOK, &response)
	}
}

func (c *HandleController) MakeDisableTokensFunc() func(context *gin.Context) {
	return func(context *gin.Context) {
		response := OperationResponse{
			Success: true,
			Code:    Success,
			Message: "remove success",
		}
		disable := TokenDisable{}
		err := context.BindJSON(&disable)
		if err != nil {
			response.Success = false
			response.Code = ParamError
			response.Message = fmt.Sprintf("disable failed, param error : %v", err)
			log.Printf(response.Message)
			context.JSON(http.StatusOK, &response)
			return
		}

		for _, user := range disable.Users {
			result := c.verifyToken(user, TOKEN_DISABLE)

			if !result.Success {
				context.JSON(http.StatusOK, &result)
				return
			}
		}

		for _, user := range disable.Users {
			token := c.Tokens[user.User]
			token.Enable = false
			if c.DB != nil {
				result := c.DB.Model(&model.UserToken{}).Where("user = ?", user.User).Update("enable", false)
				if result.Error != nil {
					response.Success = false
					response.Code = SaveError
					response.Message = fmt.Sprintf("user disable failed for %s, db error : %v", user.User, result.Error)
					log.Printf(response.Message)
					context.JSON(http.StatusOK, &response)
					return
				}
			}
			c.Tokens[user.User] = token
		}

		if c.DB == nil {
		}

		context.JSON(http.StatusOK, &response)
	}
}

func (c *HandleController) MakeEnableTokensFunc() func(context *gin.Context) {
	return func(context *gin.Context) {
		response := OperationResponse{
			Success: true,
			Code:    Success,
			Message: "remove success",
		}
		enable := TokenEnable{}
		err := context.BindJSON(&enable)
		if err != nil {
			response.Success = false
			response.Code = ParamError
			response.Message = fmt.Sprintf("enable failed, param error : %v", err)
			log.Printf(response.Message)
			context.JSON(http.StatusOK, &response)
			return
		}

		for _, user := range enable.Users {
			result := c.verifyToken(user, TOKEN_ENABLE)

			if !result.Success {
				context.JSON(http.StatusOK, &result)
				return
			}
		}

		for _, user := range enable.Users {
			token := c.Tokens[user.User]
			token.Enable = true
			if c.DB != nil {
				result := c.DB.Model(&model.UserToken{}).Where("user = ?", user.User).Update("enable", true)
				if result.Error != nil {
					response.Success = false
					response.Code = SaveError
					response.Message = fmt.Sprintf("user enable failed for %s, db error : %v", user.User, result.Error)
					log.Printf(response.Message)
					context.JSON(http.StatusOK, &response)
					return
				}
			}
			c.Tokens[user.User] = token
		}

		if c.DB == nil {
		}

		context.JSON(http.StatusOK, &response)
	}
}

// 获取所有服务器信息
func (c *HandleController) MakeQueryDashboardsFunc() func(context *gin.Context) {
	return func(context *gin.Context) {
		session := sessions.Default(context)
		userRole := session.Get(UserRoleName)

		// For normal users, only return non-sensitive dashboard information
		if userRole == UserRoleNormal {
			type UserDashboardInfo struct {
				Name          string `json:"name"`
				DashboardAddr string `json:"dashboard_addr"`
			}
			var userDashboards []UserDashboardInfo
			for _, dashboard := range c.Dashboards {
				userDashboards = append(userDashboards, UserDashboardInfo{
					Name:          dashboard.Name,
					DashboardAddr: dashboard.DashboardAddr,
				})
			}
			context.JSON(http.StatusOK, gin.H{
				"code":          0,
				"msg":           "success",
				"data":          userDashboards,
				"current_index": c.CurrentDashboardIndex,
			})
			return
		}

		// For admin users, return all dashboard information
		context.JSON(http.StatusOK, gin.H{
			"code":          0,
			"msg":           "success",
			"data":          c.Dashboards,
			"current_index": c.CurrentDashboardIndex,
		})
	}
}

// 切换当前显示的服务器信息
func (c *HandleController) MakeSwitchDashboardFunc() func(context *gin.Context) {
	return func(context *gin.Context) {
		var req struct {
			Index int `json:"index"`
		}
		if err := context.BindJSON(&req); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request body",
			})
			return
		}

		if req.Index < 0 || req.Index >= len(c.Dashboards) {
			context.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid dashboard index",
			})
			return
		}

		c.CurrentDashboardIndex = req.Index
		context.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Dashboard switched successfully",
		})
	}
}

// 保存模板信息
func (c *HandleController) MakeSaveConfigTemplateFunc() func(context *gin.Context) {
	return func(context *gin.Context) {
		var req struct {
			Template string `json:"template"`
		}
		if err := context.BindJSON(&req); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request body",
			})
			return
		}

		// 构建配置模板文件路径
		assets := filepath.Join("assets", "static", "config_template.json")
		_, err := os.Stat(assets)
		if err != nil && !os.IsExist(err) {
			assets = "./assets/static/config_template.json"
		}

		// 创建JSON对象
		configTemplate := struct {
			Template string `json:"template"`
		}{
			Template: req.Template,
		}

		// 将JSON对象转换为字节
		jsonData, err := json.MarshalIndent(configTemplate, "", "  ")
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to marshal JSON: " + err.Error(),
			})
			return
		}

		// 写入文件
		err = os.WriteFile(assets, jsonData, 0644)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to write file: " + err.Error(),
			})
			return
		}

		context.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Config template saved successfully",
		})
	}
}

// 获取当前服务器的面板信息
func (c *HandleController) MakeProxyFunc() func(context *gin.Context) {
	return func(context *gin.Context) {
		var client *http.Client
		var protocol string

		if c.CurrentDashboardIndex >= len(c.Dashboards) {
			context.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "No dashboard configured or invalid current index",
			})
			return
		}

		currentDashboard := c.Dashboards[c.CurrentDashboardIndex]

		if currentDashboard.DashboardTls {
			client = &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}
			protocol = "https://"
		} else {
			client = http.DefaultClient
			protocol = "http://"
		}

		res := ProxyResponse{}
		host := currentDashboard.DashboardAddr
		port := currentDashboard.DashboardPort

		host, _ = strings.CutPrefix(host, protocol)

		requestUrl := protocol + host + ":" + strconv.Itoa(port) + context.Param("serverApi")
		request, _ := http.NewRequest("GET", requestUrl, nil)
		username := currentDashboard.DashboardUser
		password := currentDashboard.DashboardPwd
		if trimString(username) != "" && trimString(password) != "" {
			request.SetBasicAuth(username, password)
			log.Printf("Proxy to %s", requestUrl)
		}

		response, err := client.Do(request)

		if err != nil {
			res.Code = FrpServerError
			res.Success = false
			res.Message = err.Error()
			log.Print(err)
			context.JSON(http.StatusOK, &res)
			return
		}

		res.Code = response.StatusCode
		body, err := io.ReadAll(response.Body)

		if err != nil {
			res.Success = false
			res.Message = err.Error()
		} else {
			if res.Code == http.StatusOK {
				res.Success = true
				res.Data = string(body)
				res.Message = fmt.Sprintf("Proxy to %s success", requestUrl)
			} else {
				res.Success = false
				if res.Code == http.StatusNotFound {
					res.Message = fmt.Sprintf("Proxy to %s error: url not found", requestUrl)
				} else {
					res.Message = fmt.Sprintf("Proxy to %s error: %s", requestUrl, string(body))
				}
			}
		}
		log.Printf(res.Message)
		context.JSON(http.StatusOK, &res)
	}
}

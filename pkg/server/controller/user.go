package controller

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"frps-panel/pkg/server/model"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

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

package controller

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"frps-panel/pkg/server/model"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

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

		var userTokens []model.UserToken
		if result := c.DB.Where("server = ?", serverName).Find(&userTokens); result.Error != nil {
			context.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to query tokens",
			})
			return
		}

		maxPort := 0
		for _, userToken := range userTokens {
			tokenInfo, err := ToUserTokenInfo(userToken)
			if err != nil {
				log.Printf("Failed to convert user token: %v", err)
				continue
			}
			for _, p := range tokenInfo.Ports {
				switch v := p.(type) {
				case int:
					if v > maxPort {
						maxPort = v
					}
				case float64:
					if int(v) > maxPort {
						maxPort = int(v)
					}
				case string:
					parts := strings.Split(v, "-")
					if len(parts) == 2 {
						endPort, err := strconv.Atoi(strings.TrimSpace(parts[1]))
						if err == nil && endPort > maxPort {
							maxPort = endPort
						}
					} else {
						singlePort, err := strconv.Atoi(strings.TrimSpace(v))
						if err == nil && singlePort > maxPort {
							maxPort = singlePort
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
		var userTokens []model.UserToken
		if result := c.DB.Find(&userTokens); result.Error != nil {
			context.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to query tokens",
			})
			return
		}

		maxPortsMap := make(map[string]int)
		for _, userToken := range userTokens {
			tokenInfo, err := ToUserTokenInfo(userToken)
			if err != nil {
				log.Printf("Failed to convert user token: %v", err)
				continue
			}
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
				case float64:
					if int(v) > maxPortsMap[serverName] {
						maxPortsMap[serverName] = int(v)
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

// 获取所有服务器信息
func (c *HandleController) MakeQueryDashboardsFunc() func(context *gin.Context) {
	return func(context *gin.Context) {
		session := sessions.Default(context)
		userRole := session.Get(UserRoleName)

		var servers []ServerInfo
		if result := c.DB.Find(&servers); result.Error != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to query servers"})
			return
		}

		// For normal users, only return non-sensitive dashboard information
		if userRole == UserRoleNormal {
			type UserDashboardInfo struct {
				Name          string `json:"name"`
				DashboardAddr string `json:"dashboard_addr"`
			}
			var userDashboards []UserDashboardInfo
			for _, server := range servers {
				userDashboards = append(userDashboards, UserDashboardInfo{
					Name:          server.Name,
					DashboardAddr: server.DashboardAddr,
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
			"data":          servers,
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

		var count int64
		c.DB.Model(&ServerInfo{}).Count(&count)

		if req.Index < 0 || req.Index >= int(count) {
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

		var servers []ServerInfo
		if result := c.DB.Find(&servers); result.Error != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to query servers"})
			return
		}

		if c.CurrentDashboardIndex >= len(servers) {
			context.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "No dashboard configured or invalid current index",
			})
			return
		}

		currentDashboard := servers[c.CurrentDashboardIndex]

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

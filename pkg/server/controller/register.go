package controller

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HandleController struct {
	CommonInfo            CommonInfo
	Tokens                map[string]TokenInfo
	Version               string
	ConfigFile            string
	TokensFile            string
	Dashboards            []DashboardConfig
	CurrentDashboardIndex int
	DB                    *gorm.DB
	Database              DatabaseConfig
}

func NewHandleController(config *HandleController) *HandleController {
	return config
}

func (c *HandleController) Register(rootDir string, engine *gin.Engine) {
	assets := filepath.Join(rootDir, "assets")
	_, err := os.Stat(assets)
	if err != nil && !os.IsExist(err) {
		assets = "./assets"
	}

	engine.Delims("${", "}")
	engine.LoadHTMLGlob(filepath.Join(assets, "templates/*"))
	engine.POST("/handler", c.MakeHandlerFunc())
	engine.Static("/static", filepath.Join(assets, "static"))
	engine.GET("/lang.json", c.MakeLangFunc())
	engine.GET(LoginUrl, c.MakeLoginFunc())
	engine.POST(LoginUrl, c.MakeLoginFunc())
	engine.GET(LogoutUrl, c.MakeLogoutFunc())
	engine.GET(UserDashboardUrl, c.MakeUserDashboardFunc()) // 新增普通用户仪表板路由

	var adminGroup *gin.RouterGroup
	if len(c.CommonInfo.AdminUser) != 0 {
		adminGroup = engine.Group("/", c.BasicAuth())
	} else {
		adminGroup = engine.Group("/")
	}
	adminGroup.GET("/", c.MakeIndexFunc())
	adminGroup.GET("/tokens", c.MakeQueryTokensFunc())
	adminGroup.POST("/add", c.MakeAddTokenFunc())
	adminGroup.POST("/update", c.MakeUpdateTokensFunc())
	adminGroup.POST("/remove", c.MakeRemoveTokensFunc())
	adminGroup.POST("/disable", c.MakeDisableTokensFunc())
	adminGroup.POST("/enable", c.MakeEnableTokensFunc())
	adminGroup.GET("/proxy/*serverApi", c.MakeProxyFunc())
	adminGroup.GET("/dashboards", c.MakeQueryDashboardsFunc())
	adminGroup.POST("/switch_dashboard", c.MakeSwitchDashboardFunc())
	adminGroup.GET("/get_max_port", c.MakeGetMaxPortFunc())
	adminGroup.GET("/get_all_max_ports", c.MakeGetAllMaxPortsFunc())
	adminGroup.POST("/save_config_template", c.MakeSaveConfigTemplateFunc())

	// 普通用户API路由
	userApiGroup := engine.Group("/api/user", c.BasicAuth())
	userApiGroup.GET("/info", c.MakeQueryUserInfoFunc())       // 新增获取用户信息的API
	userApiGroup.GET("/proxies", c.MakeQueryUserProxiesFunc()) // 新增获取用户代理列表的API
	userApiGroup.GET("/dashboards", c.MakeQueryDashboardsFunc())
}

package controller

import (
	"encoding/json"
	"frps-panel/pkg/server/model"
	"regexp"
)

const (
	Success int = iota
	ParamError
	UserExist
	UserNotExist
	SaveError
	UserFormatError
	TokenFormatError
	CommentFormatError
	PortsFormatError
	DomainsFormatError
	SubdomainsFormatError
	ExpireDateFormatError // 新增到期时间格式错误
	FrpServerError
)

const (
	TOKEN_ADD int = iota
	TOKEN_UPDATE
	TOKEN_REMOVE
	TOKEN_ENABLE
	TOKEN_DISABLE
)

const (
	SessionName      = "GOSESSION"
	AuthName         = "_PANEL_AUTH"
	LoginUrl         = "/login"
	LoginSuccessUrl  = "/"
	LogoutUrl        = "/logout"
	LogoutSuccessUrl = "/login"
	UserDashboardUrl = "/user/dashboard" // 新增普通用户仪表板URL
)

const (
	UserRoleAdmin  = "admin"
	UserRoleNormal = "normal"
	UserRoleName   = "_PANEL_USER_ROLE" // 新增用户角色会话键
)

var (
	userFormat        = regexp.MustCompile("^\\w+$")
	tokenFormat       = regexp.MustCompile("^[\\w!@#$%^&*()]+$")
	commentFormat     = regexp.MustCompile("[\\n\\t\\r]")
	portsFormatSingle = regexp.MustCompile("^\\s*\\d{1,5}\\s*$")
	portsFormatRange  = regexp.MustCompile("^\\s*\\d{1,5}\\s*-\\s*\\d{1,5}\\s*$")
	domainFormat      = regexp.MustCompile("^([a-zA-Z0-9]+(-[a-zA-Z0-9]+)*\\.)+[a-zA-Z]{2,}$")
	subdomainFormat   = regexp.MustCompile("^[a-zA-z0-9-]{1,20}$")
	expireDateFormat  = regexp.MustCompile("^\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2}$") // 新增到期时间格式
	trimAllSpace      = regexp.MustCompile("[\\n\\t\\r\\s]")
)

type Response struct {
	Msg string `json:"msg"`
}

type HTTPError struct {
	Code int
	Err  error
}

type DatabaseConfig struct {
	Enable bool   `toml:"enable"`
	Type   string `toml:"type"`
	Dsn    string `toml:"dsn"`
}

type Common struct {
	Common   CommonInfo     `toml:"common"`
	Database DatabaseConfig `toml:"database"`
}

type CommonInfo struct {
	PluginAddr    string `toml:"plugin_addr"`
	PluginPort    int    `toml:"plugin_port"`
	AdminUser     string `toml:"admin_user"`
	AdminPwd      string `toml:"admin_pwd"`
	AdminKeepTime int    `toml:"admin_keep_time"`
	TlsMode       bool   `toml:"tls_mode"`
	TlsCertFile   string `toml:"tls_cert_file"`
	TlsKeyFile    string `toml:"tls_key_file"`
}

type ServerInfo struct {
	Name          string `toml:"name" json:"name"`
	DashboardAddr string `toml:"dashboard_addr" json:"dashboard_addr"`
	DashboardPort int    `toml:"dashboard_port" json:"dashboard_port"`
	DashboardUser string `toml:"dashboard_user" json:"dashboard_user"`
	DashboardPwd  string `toml:"dashboard_pwd" json:"dashboard_pwd"`
	DashboardTls  bool   `json:"dashboard_tls"`
}

type UserTokenInfo struct {
	User       string   `json:"user" form:"user"`
	Token      string   `json:"token" form:"token"`
	Comment    string   `json:"comment" form:"comment"`
	Ports      []any    `json:"ports" form:"ports"`
	Domains    []string `json:"domains" form:"domains"`
	Subdomains []string `json:"subdomains" form:"subdomains"`
	Enable     bool     `json:"enable" form:"enable"`
	Server     string   `json:"server" form:"server"`
	CreateDate string   `json:"create_date" form:"create_date"`
	ExpireDate string   `json:"expire_date" form:"expire_date"`
}

type TokenResponse struct {
	Code  int             `json:"code"`
	Msg   string          `json:"msg"`
	Count int             `json:"count"`
	Data  []UserTokenInfo `json:"data"`
}

type OperationResponse struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ProxyResponse struct {
	OperationResponse
	Data string `json:"data"`
}

type TokenSearch struct {
	UserTokenInfo
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

type TokenUpdate struct {
	Before UserTokenInfo `json:"before"`
	After  UserTokenInfo `json:"after"`
}

type TokenRemove struct {
	Users []UserTokenInfo `json:"users"`
}

type TokenDisable struct {
	TokenRemove
}

type TokenEnable struct {
	TokenDisable
}

func (e *HTTPError) Error() string {
	return e.Err.Error()
}

func ToUserTokenInfo(userToken model.UserToken) (UserTokenInfo, error) {
	info := UserTokenInfo{
		User:       userToken.User,
		Token:      userToken.Token,
		Comment:    userToken.Comment,
		Enable:     userToken.Enable,
		Server:     userToken.Server,
		CreateDate: userToken.CreateDate,
		ExpireDate: userToken.ExpireDate,
	}
	if userToken.Ports != "" {
		if err := json.Unmarshal([]byte(userToken.Ports), &info.Ports); err != nil {
			return info, err
		}
	}
	if userToken.Domains != "" {
		if err := json.Unmarshal([]byte(userToken.Domains), &info.Domains); err != nil {
			return info, err
		}
	}
	if userToken.Subdomains != "" {
		if err := json.Unmarshal([]byte(userToken.Subdomains), &info.Subdomains); err != nil {
			return info, err
		}
	}
	return info, nil
}

func FromUserTokenInfo(info UserTokenInfo) (model.UserToken, error) {
	userToken := model.UserToken{
		User:       info.User,
		Token:      info.Token,
		Comment:    info.Comment,
		Enable:     info.Enable,
		Server:     info.Server,
		CreateDate: info.CreateDate,
		ExpireDate: info.ExpireDate,
	}
	ports, err := json.Marshal(info.Ports)
	if err != nil {
		return userToken, err
	}
	userToken.Ports = string(ports)

	domains, err := json.Marshal(info.Domains)
	if err != nil {
		return userToken, err
	}
	userToken.Domains = string(domains)

	subdomains, err := json.Marshal(info.Subdomains)
	if err != nil {
		return userToken, err
	}
	userToken.Subdomains = string(subdomains)

	return userToken, nil
}

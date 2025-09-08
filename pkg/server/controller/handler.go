package controller

import (
	"fmt"
	"frps-panel/pkg/server/model"
	"log"
	"strconv"
	"strings"

	plugin "github.com/fatedier/frp/pkg/plugin/server"
)

func (c *HandleController) HandleLogin(content *plugin.LoginContent, remoteIP string) plugin.Response {
	token := content.Metas["token"]
	user := content.User
	res := c.JudgeToken(user, token)
	if res.Reject {
		return res
	}

	var userToken model.UserToken
	if result := c.DB.Where("user = ?", user).First(&userToken); result.Error == nil {
		if userToken.Server != "" {
			var serverConf model.ServerInfo
			if result := c.DB.Where("name = ?", userToken.Server).First(&serverConf); result.Error != nil {
				res.Reject = true
				res.RejectReason = fmt.Sprintf("user [%s] is configured for server [%s], but this server is not defined", user, userToken.Server)
				return res
			}

			if serverConf.Name == "" {
				res.Reject = true
				res.RejectReason = fmt.Sprintf("user [%s] is configured for server [%s], but this server is not defined", user, userToken.Server)
			} else if serverConf.DashboardAddr != remoteIP {
				res.Reject = true
				res.RejectReason = fmt.Sprintf("user [%s] is not allowed to login from this server [%s]", user, remoteIP)
			}
		}
	}
	return res
}

func (c *HandleController) HandleNewProxy(content *plugin.NewProxyContent) plugin.Response {
	token := content.User.Metas["token"]
	user := content.User.User
	judgeToken := c.JudgeToken(user, token)
	if judgeToken.Reject {
		return judgeToken
	}
	return c.JudgePort(content)
}

func (c *HandleController) HandlePing(content *plugin.PingContent) plugin.Response {
	token := content.User.Metas["token"]
	user := content.User.User
	return c.JudgeToken(user, token)
}

func (c *HandleController) HandleNewWorkConn(content *plugin.NewWorkConnContent) plugin.Response {
	token := content.User.Metas["token"]
	user := content.User.User
	return c.JudgeToken(user, token)
}

func (c *HandleController) HandleNewUserConn(content *plugin.NewUserConnContent) plugin.Response {
	token := content.User.Metas["token"]
	user := content.User.User
	return c.JudgeToken(user, token)
}

func (c *HandleController) JudgeToken(user string, token string) plugin.Response {
	var res plugin.Response
	if user == "" || token == "" {
		res.Reject = true
		res.RejectReason = "user or meta token can not be empty"
		return res
	}

	var userToken model.UserToken
	if result := c.DB.Where("user = ?", user).First(&userToken); result.Error != nil {
		res.Reject = true
		res.RejectReason = fmt.Sprintf("user [%s] not exist", user)
		return res
	}

	if !userToken.Enable {
		res.Reject = true
		res.RejectReason = fmt.Sprintf("user [%s] is disabled", user)
	} else if userToken.Token != token {
		res.Reject = true
		res.RejectReason = fmt.Sprintf("invalid meta token for user [%s]", user)
	} else {
		res.Unchange = true
	}

	return res
}

func (c *HandleController) JudgePort(content *plugin.NewProxyContent) plugin.Response {
	var res plugin.Response
	var portErr error
	var reject = false
	supportProxyTypes := []string{
		"tcp", "tcpmux", "udp", "http", "https",
	}
	proxyType := content.ProxyType
	if !stringContains(proxyType, supportProxyTypes) {
		log.Printf("proxy type [%v] not support, plugin do nothing", proxyType)
		res.Unchange = true
		return res
	}

	user := content.User.User
	userPort := content.RemotePort
	userDomains := content.CustomDomains
	userSubdomain := content.SubDomain

	portAllowed := true
	if proxyType == "tcp" || proxyType == "udp" {
		portAllowed = false
		var userToken model.UserToken
		if result := c.DB.Where("user = ?", user).First(&userToken); result.Error == nil {
			info, err := ToUserTokenInfo(userToken)
			if err != nil {
				portErr = fmt.Errorf("failed to convert user token: %v", err)
			} else {
				for _, port := range info.Ports {
					if str, ok := port.(string); ok {
						if strings.Contains(str, "-") {
							allowedRanges := strings.Split(str, "-")
							if len(allowedRanges) != 2 {
								portErr = fmt.Errorf("user [%v] port range [%v] format error", user, port)
								break
							}
							start, err := strconv.Atoi(trimString(allowedRanges[0]))
							if err != nil {
								portErr = fmt.Errorf("user [%v] port rang [%v] start port [%v] is not a number", user, port, allowedRanges[0])
								break
							}
							end, err := strconv.Atoi(trimString(allowedRanges[1]))
							if err != nil {
								portErr = fmt.Errorf("user [%v] port rang [%v] end port [%v] is not a number", user, port, allowedRanges[0])
								break
							}
							if max(userPort, start) == userPort && min(userPort, end) == userPort {
								portAllowed = true
								break
							}
						} else {
							if str == "" {
								portAllowed = true
								break
							}
							allowed, err := strconv.Atoi(str)
							if err != nil {
								portErr = fmt.Errorf("user [%v] allowed port [%v] is not a number", user, port)
							}
							if allowed == userPort {
								portAllowed = true
								break
							}
						}
					} else {
						allowed := port
						if allowed == userPort {
							portAllowed = true
							break
						}
					}
				}
			}
		} else {
			portAllowed = true
		}
	}
	if !portAllowed {
		if portErr == nil {
			portErr = fmt.Errorf("user [%v] port [%v] is not allowed", user, userPort)
		}
		reject = true
	}

	domainAllowed := true
	if proxyType == "http" || proxyType == "https" || proxyType == "tcpmux" {
		if portAllowed {
			var userToken model.UserToken
			if result := c.DB.Where("user = ?", user).First(&userToken); result.Error == nil {
				info, err := ToUserTokenInfo(userToken)
				if err != nil {
					portErr = fmt.Errorf("failed to convert user token: %v", err)
				} else {
					if stringContains("", info.Domains) {
						domainAllowed = true
					} else {
						for _, userDomain := range userDomains {
							if !stringContains(userDomain, info.Domains) {
								domainAllowed = false
								break
							}
						}
					}
				}
			}
			if !domainAllowed {
				portErr = fmt.Errorf("user [%v] domain [%v] is not allowed", user, strings.Join(userDomains, ","))
				reject = true
			}
		}
	}

	subdomainAllowed := true
	if proxyType == "http" || proxyType == "https" {
		subdomainAllowed = false
		if portAllowed && domainAllowed {
			var userToken model.UserToken
			if result := c.DB.Where("user = ?", user).First(&userToken); result.Error == nil {
				info, err := ToUserTokenInfo(userToken)
				if err != nil {
					portErr = fmt.Errorf("failed to convert user token: %v", err)
				} else {
					if stringContains("", info.Subdomains) {
						subdomainAllowed = true
					} else {
						for _, subdomain := range info.Subdomains {
							if subdomain == userSubdomain {
								subdomainAllowed = true
								break
							}
						}
					}
				}
			} else {
				subdomainAllowed = true
			}
			if !subdomainAllowed {
				portErr = fmt.Errorf("user [%v] subdomain [%v] is not allowed", user, userSubdomain)
				reject = true
			}
		}
	}

	if reject {
		res.Reject = true
		res.RejectReason = portErr.Error()
	} else {
		res.Unchange = true
	}
	return res
}

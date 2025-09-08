package controller

import (
	"net/http"

	ginI18n "github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
)

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

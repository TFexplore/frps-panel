package controller

import (
	"fmt"
	"frps-panel/pkg/server/model"
	"log"
	"strings"
)

func filter(main UserTokenInfo, sub UserTokenInfo) bool {
	replaceSpaceUser := trimAllSpace.ReplaceAllString(sub.User, "")
	if len(replaceSpaceUser) != 0 {
		if !strings.Contains(main.User, replaceSpaceUser) {
			return false
		}
	}

	replaceSpaceToken := trimAllSpace.ReplaceAllString(sub.Token, "")
	if len(replaceSpaceToken) != 0 {
		if !strings.Contains(main.Token, replaceSpaceToken) {
			return false
		}
	}

	replaceSpaceComment := trimAllSpace.ReplaceAllString(sub.Comment, "")
	if len(replaceSpaceComment) != 0 {
		if !strings.Contains(main.Comment, replaceSpaceComment) {
			return false
		}
	}
	return true
}

func trimString(str string) string {
	return strings.TrimSpace(str)
}

func (c *HandleController) verifyToken(token UserTokenInfo, operate int) OperationResponse {
	response := OperationResponse{
		Success: true,
		Code:    Success,
		Message: "operate success",
	}

	var (
		validateExist      = false
		validateNotExist   = false
		validateUser       = false
		validateToken      = false
		validateComment    = false
		validatePorts      = false
		validateDomains    = false
		validateSubdomains = false
		validateExpireDate = false // 新增验证到期时间
	)

	if operate == TOKEN_ADD {
		validateExist = true
		validateUser = true
		validateToken = true
		validateComment = true
		validatePorts = true
		validateDomains = true
		validateSubdomains = true
		validateExpireDate = true // 新增验证到期时间
	} else if operate == TOKEN_UPDATE {
		validateNotExist = true
		validateUser = true
		validateToken = true
		validateComment = true
		validatePorts = true
		validateDomains = true
		validateSubdomains = true
		validateExpireDate = true // 新增验证到期时间
	} else if operate == TOKEN_ENABLE || operate == TOKEN_DISABLE || operate == TOKEN_REMOVE {
		validateNotExist = true
	}

	if validateUser && !userFormat.MatchString(token.User) {
		response.Success = false
		response.Code = UserFormatError
		response.Message = fmt.Sprintf("operate failed, user [%s] format error", token.User)
		log.Printf(response.Message)
		return response
	}

	if validateExist {
		var count int64
		c.DB.Model(&model.UserToken{}).Where("user = ?", token.User).Count(&count)
		if count > 0 {
			response.Success = false
			response.Code = UserExist
			response.Message = fmt.Sprintf("operate failed, user [%s] exist ", token.User)
			log.Printf(response.Message)
			return response
		}
	}

	if validateNotExist {
		var count int64
		c.DB.Model(&model.UserToken{}).Where("user = ?", token.User).Count(&count)
		if count == 0 {
			response.Success = false
			response.Code = UserNotExist
			response.Message = fmt.Sprintf("operate failed, user [%s] not exist ", token.User)
			log.Printf(response.Message)
			return response
		}
	}

	if validateToken && !tokenFormat.MatchString(token.Token) {
		response.Success = false
		response.Code = TokenFormatError
		response.Message = fmt.Sprintf("operate failed, token [%s] format error", token.Token)
		log.Printf(response.Message)
		return response
	}

	trimmedComment := trimString(token.Comment)
	if validateComment && trimmedComment != "" && commentFormat.MatchString(trimmedComment) {
		response.Success = false
		response.Code = CommentFormatError
		response.Message = fmt.Sprintf("operate failed, comment [%s] format error", token.Comment)
		log.Printf(response.Message)
		return response
	}

	if validatePorts {
		for _, port := range token.Ports {
			if str, ok := port.(string); ok {
				trimmedPort := trimString(str)
				if trimmedPort != "" && !portsFormatSingle.MatchString(trimmedPort) && !portsFormatRange.MatchString(trimmedPort) {
					response.Success = false
					response.Code = PortsFormatError
					response.Message = fmt.Sprintf("operate failed, ports [%v] format error", token.Ports)
					log.Printf(response.Message)
					return response
				}
			}
		}
	}

	if validateDomains {
		for _, domain := range token.Domains {
			trimmedDomain := trimString(domain)
			if trimmedDomain != "" && !domainFormat.MatchString(trimmedDomain) {
				response.Success = false
				response.Code = DomainsFormatError
				response.Message = fmt.Sprintf("operate failed, domains [%v] format error", token.Domains)
				log.Printf(response.Message)
				return response
			}
		}
	}

	if validateSubdomains {
		for _, subdomain := range token.Subdomains {
			trimmedSubdomain := trimString(subdomain)
			if trimmedSubdomain != "" && !subdomainFormat.MatchString(trimmedSubdomain) {
				response.Success = false
				response.Code = SubdomainsFormatError
				response.Message = fmt.Sprintf("operate failed, subdomains [%v] format error", token.Subdomains)
				log.Printf(response.Message)
				return response
			}
		}
	}

	// 新增到期时间验证
	trimmedExpireDate := trimString(token.ExpireDate)
	if validateExpireDate && trimmedExpireDate != "" && !expireDateFormat.MatchString(trimmedExpireDate) {
		response.Success = false
		response.Code = ExpireDateFormatError
		response.Message = fmt.Sprintf("operate failed, expire date [%s] format error", token.ExpireDate)
		log.Printf(response.Message)
		return response
	}

	return response
}

func cleanPorts(ports []any) []any {
	cleanedPorts := make([]any, len(ports))
	for i, port := range ports {
		if str, ok := port.(string); ok {
			cleanedPorts[i] = cleanString(str)
		} else {
			//float64, for JSON numbers
			cleanedPorts[i] = int(port.(float64))
		}
	}
	return cleanedPorts
}

func cleanStrings(originalStrings []string) []string {
	cleanedStrings := make([]string, len(originalStrings))
	for i, str := range originalStrings {
		cleanedStrings[i] = cleanString(str)
	}
	return cleanedStrings
}

func cleanString(originalString string) string {
	return trimString(originalString)
}

func stringContains(element string, data []string) bool {
	for _, v := range data {
		if element == v {
			return true
		}
	}
	return false
}

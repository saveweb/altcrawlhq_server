package clientauth

import "github.com/gin-gonic/gin"

func IsAuthorized(c *gin.Context) bool {
	authKey := c.GetHeader("X-Auth-Key")
	authSecret := c.GetHeader("X-Auth-Secret")
	// identifier := c.GetHeader("X-Identifier")

	if authKey == "" || authSecret == "" {
		return false
	}

	if authKey == "saveweb_key" && authSecret == "saveweb_sec" {
		return true
	}

	return false
}

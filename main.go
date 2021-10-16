package main

import (
	"github.com/gin-gonic/gin"
	"github.com/michibiki-io/ldap-jwt-go/controller"
)

func main() {
	engine := gin.Default()
	v1 := engine.Group("/v1")
	{
		v1.Any("/authorize", controller.Authorize)
		v1.Any("/verify", controller.Verify)
		v1.Any("/refresh", controller.Refresh)
		v1.Any("/deauthorize", controller.Deauthorize)
	}
	engine.Run(":80")
}

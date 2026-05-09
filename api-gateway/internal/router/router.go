package router

import (
	"github.com/gin-gonic/gin"

	"github.com/sunxi/92-Account-Center/account-service/internal/handler"
	"github.com/sunxi/92-Account-Center/account-service/internal/service"
)

// SetupRouter sets up the Gin router with all routes.
func SetupRouter(userService service.UserService) *gin.Engine {
	r := gin.Default()

	// Register handler
	registerHandler := handler.NewRegisterHandler(userService)

	// User registration endpoint
	r.POST("/register", registerHandler.Register)

	return r
}
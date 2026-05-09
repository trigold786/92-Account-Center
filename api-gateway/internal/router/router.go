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

	// Password handler
	passwordHandler := handler.NewPasswordHandler(userService)

	// Password change endpoints
	r.POST("/password/send-verification-code", passwordHandler.SendVerificationCode)
	r.POST("/password/change", passwordHandler.ChangePassword)

	// Deletion handler
	deletionHandler := handler.NewDeletionHandler(userService)

	// Account deletion endpoints
	r.POST("/account/deletion/request", deletionHandler.RequestDeletion)
	r.POST("/account/deletion/cancel", deletionHandler.CancelDeletion)
	r.GET("/account/deletion/status", deletionHandler.GetDeletionStatus)

	// TODO: Add login routes when auth service is implemented
	// For now, we'll add placeholder comments

	return r
}
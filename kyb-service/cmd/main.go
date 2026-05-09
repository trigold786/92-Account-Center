package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"account-center/kyb-service/internal/handler"
	"account-center/kyb-service/internal/service"
)

func main() {
	r := gin.Default()

	kybService := service.NewKYBService()
	kybHandler := handler.NewKYBHandler(kybService)

	kybGroup := r.Group("/kyb")
	{
		kybGroup.POST("/enterprise/submit", kybHandler.SubmitEnterprise)
		kybGroup.POST("/enterprise/micro-payment/init", kybHandler.InitiateMicroPayment)
		kybGroup.POST("/enterprise/micro-payment/verify", kybHandler.VerifyMicroPayment)
		kybGroup.POST("/enterprise/face/verify", kybHandler.SubmitFaceVerification)
		kybGroup.GET("/enterprise/status/:enterprise_id", kybHandler.GetEnterpriseStatus)
	}

	log.Println("KYB service starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
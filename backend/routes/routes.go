package routes

import (
	servicesU "github.com/dath-241/coin-price-be-go/services/trigger-service/services"
	servicesA "github.com/dath-241/coin-price-be-go/services/trigger-service/services/alert"
	servicesI "github.com/dath-241/coin-price-be-go/services/trigger-service/services/indicator"
	services "github.com/dath-241/coin-price-be-go/services/trigger-service/services/snooze"
	"github.com/gin-gonic/gin"
)

func SetupRoute() *gin.Engine {
	route := gin.Default()

	alerts := route.Group("/api/v1/vip2")
	{
		alerts.POST("/alerts", servicesA.CreateAlert)
		alerts.GET("/alerts", servicesA.GetAlerts)
		alerts.GET("/alerts/:id", servicesA.GetAlert)
		alerts.DELETE("/alerts/:id", servicesA.DeleteAlert)
		alerts.GET("/symbol-alerts", servicesA.GetSymbolAlerts)
		alerts.POST("/alerts/symbol", servicesA.SetSymbolAlert)

		alerts.POST("/start-alert-checker", func(c *gin.Context) {
			services.StartRunning()
			c.JSON(200, gin.H{"status": "Alert checker started"})
		})

		alerts.POST("/stop-alert-checker", func(c *gin.Context) {
			services.StopRunning()
			c.JSON(200, gin.H{"status": "Alert checker stopped"})
		})

	}

	indicators := route.Group("/api/v1/vip3/indicators")
	{
		indicators.POST("/", servicesI.SetAdvancedIndicatorAlert)
	}

	users := route.Group("/api/v1/users")
	{
		users.GET("/:id/alerts", servicesU.GetUserAlerts)
		users.POST("/:id/alerts/notify", servicesU.NotifyUser)
		users.POST("/", servicesU.CreateUser)
	}

	return route
}
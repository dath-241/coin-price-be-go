package services

import (
	"context"
	"log"
	"net/http"
	"time"

	models "github.com/dath-241/coin-price-be-go/services/trigger-service/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	config "github.com/dath-241/coin-price-be-go/services/admin_service/config"
)

// Handler to create an alert
// @Summary Create an alert
// @Description Create a new alert with the given details
// @Tags Alerts
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param body body models.Alert true "Alert details"
// @Success 201 {object} models.ResponseAlertCreated "Successfully created alert"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 500 {object} models.ErrorResponse "Failed to create alert"
// @Security ApiKeyAuth
// @Router /api/v1/vip2/alerts [post]
func CreateAlert(c *gin.Context) {

	var newAlert models.Alert
	if err := c.ShouldBindJSON(&newAlert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	// if newAlert.Symbol == "" || newAlert.Price == 0 || (newAlert.Condition != ">=" && newAlert.Condition != "<=" && newAlert.Condition != "==") {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid fields"})
	// 	return
	// }

	// Add default or new values for additional fields
	newAlert.ID = primitive.NewObjectID()
	newAlert.IsActive = true
	currentTime := primitive.NewDateTimeFromTime(time.Now())
	newAlert.CreatedAt = currentTime
	newAlert.UpdatedAt = currentTime

	// Add default values for new fields
	if newAlert.Frequency == "" {
		newAlert.Frequency = "immediate" // Set default frequency if not provided
	}
	if newAlert.MaxRepeatCount == 0 {
		newAlert.MaxRepeatCount = 5 // Set default max repeat count if not provided
	}
	if newAlert.SnoozeCondition == "" {
		newAlert.SnoozeCondition = "none" // Set default snooze condition if not provided
	}
	if newAlert.Range == nil {
		newAlert.Range = []float64{} // Set default range if not provided
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := config.AlertCollection.InsertOne(ctx, newAlert)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create alert"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Alert created successfully",
		"alert_id": newAlert.ID.Hex(),
	})
}

// Handler to retrieve all alerts
// @Summary Get all alerts
// @Description Retrieve all alerts, optionally filter by type
// @Tags Alerts
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param type query string false "Filter by alert type (e.g., new_listing, delisting)"
// @Success 200 {array} models.ResponseAlertList "List of alerts"
// @Failure 500 {object} models.ErrorResponse "Failed to retrieve alerts"
// @Router /api/v1/vip2/alerts [get]
func GetAlerts(c *gin.Context) {

	var results []models.Alert
	alertType := c.Query("type")

	filter := bson.M{}
	if alertType != "" {
		filter["type"] = alertType
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := config.AlertCollection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve alerts"})
		return
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse alerts"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// Handler to get an alert by ID
// @Summary Get an alert by ID
// @Description Retrieve an alert by its unique identifier
// @Tags Alerts
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param id path string true "Alert ID"
// @Success 200 {object} models.ResponseAlertDetail "Alert details"
// @Failure 400 {object} models.ErrorResponse "Invalid alert ID"
// @Failure 404 {object} models.ErrorResponse "Alert not found"
// @Router /api/v1/vip2/alerts/{id} [get]
func GetAlert(c *gin.Context) {

	id := c.Param("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	var alert models.Alert
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = config.AlertCollection.FindOne(ctx, bson.M{"_id": objectId}).Decode(&alert)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
		return
	}

	c.JSON(http.StatusOK, alert)
}

// Handler to delete an alert by ID
// @Summary Delete an alert
// @Description Delete an alert by its unique identifier
// @Tags Alerts
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param id path string true "Alert ID"
// @Success 200 {object} models.ResponseAlertDeleted "Alert deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid alert ID"
// @Failure 404 {object} models.ErrorResponse "Alert not found"
// @Router /api/v1/vip2/alerts/{id} [delete]
func DeleteAlert(c *gin.Context) {

	id := c.Param("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := config.AlertCollection.DeleteOne(ctx, bson.M{"_id": objectId})
	if err != nil || result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert deleted successfully"})
}

// Handler to retrieve new and delisted symbols
// @Summary Get new and delisted symbols
// @Description Retrieve new and delisted symbols from Binance
// @Tags Alerts
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} models.ResponseNewDelistedSymbols "List of new and delisted symbols"
// @Failure 500 {object} models.ErrorResponse "Failed to retrieve symbols"
// @Router /api/v1/vip2/symbols-alerts [get]
func GetSymbolAlerts(c *gin.Context) {

	newSymbols, delistedSymbols, err := FetchSymbolsFromBinance()
	if err != nil {
		log.Printf("Error fetching symbol data: %v", err)
		return
	}

	if newSymbols == nil {
		newSymbols = []string{}
	}
	if delistedSymbols == nil {
		delistedSymbols = []string{}
	}

	response := gin.H{
		"new_symbols":      newSymbols,
		"delisted_symbols": delistedSymbols,
	}

	c.JSON(http.StatusOK, response)
}

// Handler to set a symbol alert for new or delisted symbols
// Handler to set a symbol alert for new or delisted symbols
// @Summary Set an alert for new or delisted symbols
// @Description Set a new alert for symbols that are newly listed or delisted
// @Tags Alerts
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param body body models.Alert true "Alert details"
// @Success 201 {object} models.ResponseSetSymbolAlert "Successfully created alert for symbol"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 500 {object} models.ErrorResponse "Failed to create alert for symbol"
// @Router /api/v1/vip2/alerts/symbol [post]
func SetSymbolAlert(c *gin.Context) {

	var newAlert models.Alert
	if err := c.ShouldBindJSON(&newAlert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// if (newAlert.Type != "new_listing" && newAlert.Type != "delisting") || newAlert.NotificationMethod == "" || len(newAlert.Symbols) == 0 || newAlert.Frequency == "" {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid fields"})
	// 	return
	// }

	newAlert.ID = primitive.NewObjectID()
	newAlert.IsActive = true

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := config.AlertCollection.InsertOne(ctx, newAlert)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create alert"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Alert created successfully",
		"alert_id": newAlert.ID.Hex(),
	})

}

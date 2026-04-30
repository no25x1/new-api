package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"new-api/common"
	"new-api/middleware"
	"new-api/model"
	"new-api/router"
)

func main() {
	// Load environment variables from .env file if it exists
	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	common.SetupLogger()
	common.SysLog("New API starting...")

	// Initialize database
	err = model.InitDB()
	if err != nil {
		common.FatalLog("Failed to initialize database: " + err.Error())
	}
	defer model.CloseDB()

	// Initialize Redis if configured
	err = common.InitRedisClient()
	if err != nil {
		// Redis is optional; log the error but continue startup
		common.SysError("Failed to initialize Redis (continuing without it): " + err.Error())
	}

	// Initialize options from database
	err = model.InitOptionMap()
	if err != nil {
		common.FatalLog("Failed to initialize options: " + err.Error())
	}

	// Set Gin mode based on environment
	// Default to release mode for safety; set GIN_MODE=debug to enable debug logging
	if os.Getenv("GIN_MODE") == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	server := gin.New()
	server.Use(gin.Recovery())
	server.Use(middleware.RequestId())
	middleware.SetUpLogger(server)

	// Register all routes
	router.SetRouter(server)

	var port = os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(common.ServerPort)
	}

	// Log the full address for clarity when running locally
	common.SysLog("Server is running on http://localhost:" + port)
	common.SysLog("Tip: set GIN_MODE=debug in your .env file to enable verbose request logging")
	// Note: use SERVER_ADDR env var to bind to a specific interface (e.g. 0.0.0.0 or 127.0.0.1)
	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = ":"
	} else {
		addr = addr + ":"
	}
	err = server.Run(addr + port)
	if err != nil {
		common.FatalLog("Failed to start server: " + err.Error())
	}
}

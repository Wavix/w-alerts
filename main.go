package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/wavix/w-alerts/api"
	"github.com/wavix/w-alerts/requests"
	"github.com/wavix/w-alerts/rule"
	"github.com/wavix/w-alerts/types"
	"github.com/wavix/w-alerts/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	loadEnvironmentVariables()
}

func main() {
	done := make(chan bool)

	registry := rule.Registry{
		Rules: make(map[string]*rule.Rule),
		Mutex: sync.RWMutex{},
	}

	loadRules(&registry)

	go process(&registry)
	go ticker(&registry, done)

	startApplicationServer(&registry)

	<-done
}

func ticker(registry *rule.Registry, done chan bool) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			utils.Logger.Info().Msg("Stopping the ticker")
			return
		case <-ticker.C:
			process(registry)
		}
	}
}

func process(registry *rule.Registry) {
	for _, rule := range registry.Rules {
		if rule.IsStaticAlert {
			continue
		}

		nextRunTimer := rule.GetNextRunAt()
		if nextRunTimer.After(time.Now()) {
			continue
		}

		result, err := execRule(rule)
		if err != nil {
			utils.Logger.Error().Msgf("Error executing rule: %v", err)
			continue
		}

		rule.ProcessResponse(*result)
	}
}

func execRule(rule *rule.Rule) (*types.RuleResponse, error) {
	if rule.Request.Elastic != nil {
		result, err := requests.ExecElasticRule(rule)
		if err != nil {
			return nil, err
		}

		return &result, nil
	}

	if rule.Request.Http != nil {
		result, err := requests.ExecHttpRule(rule)
		if err != nil {
			return nil, err
		}

		return &result, nil
	}

	return nil, fmt.Errorf("unsupported request type")
}

func startApplicationServer(registry *rule.Registry) {
	gin.SetMode(gin.ReleaseMode)

	router := setupRouter(registry)
	startServer(router)
}

func setupRouter(registry *rule.Registry) *gin.Engine {
	router := gin.New()
	router.RedirectTrailingSlash = true

	allowedIPs := getAllowIps()

	utils.Logger.Info().Msg("Allowed IPs: " + strings.Join(allowedIPs, ", "))

	controllers := api.NewControllers(registry)
	router.Use(utils.IpCheckMiddleware(allowedIPs))
	router.Use(utils.GinLogger())
	controllers.Routes(router)

	// Serve static files
	router.Static("/static", "./public")

	// Serve status page
	router.GET("/", func(c *gin.Context) {
		c.File("./public/status.html")
	})

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Route not found"})
	})

	return router
}

func startServer(router *gin.Engine) {
	port := os.Getenv("PORT")
	utils.Logger.Info().Msgf("Starting REST API server on port: %s", port)

	err := router.Run(fmt.Sprintf("0.0.0.0:%s", port))
	if err != nil {
		log.Panic(err)
		return
	}
}

func loadEnvironmentVariables() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func getAllowIps() []string {
	allowIps := make([]string, 1, 10)
	allowIps = append(allowIps, "127.0.0.1")
	allowIps = append(allowIps, "::1")

	whitelist := os.Getenv("WHITELIST")
	if whitelist == "" {
		return allowIps
	}

	allowIps = append(allowIps, strings.Split(whitelist, ",")...)

	return allowIps
}

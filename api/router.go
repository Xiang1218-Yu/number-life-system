package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"number-life-system/internal/handler"
	"number-life-system/internal/middleware"
	"number-life-system/internal/service"
)

func NewRouter(db *gorm.DB, secret, origin string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), cors(origin))
	h := &handler.Handler{Auth: &service.AuthService{DB: db, Secret: secret}, Accounts: &service.AccountService{DB: db}, Subscriptions: &service.SubscriptionService{DB: db}, Footprints: &service.FootprintService{DB: db}, Backups: &service.BackupService{DB: db}, DataLocations: &service.DataLocationService{DB: db}, Categories: &service.CategoryService{DB: db}, Notifications: &service.NotificationService{DB: db}, CSV: &service.CSVService{DB: db}, Security: &service.SecurityService{DB: db}, Stats: &service.StatsService{DB: db}, Export: &service.ExportService{DB: db}}
	r.StaticFS("/static", http.Dir("./web"))
	r.GET("/", func(c *gin.Context) { c.File("./web/index.html") })
	auth := r.Group("/api/v1/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	protected := r.Group("/api/v1")
	protected.Use(middleware.Auth(secret))
	protected.GET("/accounts", h.AccountsList)
	protected.POST("/accounts", h.AccountCreate)
	protected.GET("/accounts/:id", h.AccountGet)
	protected.PUT("/accounts/:id", h.AccountUpdate)
	protected.DELETE("/accounts/:id", h.AccountDelete)
	protected.GET("/subscriptions", h.SubscriptionsList)
	protected.POST("/subscriptions", h.SubscriptionCreate)
	protected.PUT("/subscriptions/:id", h.SubscriptionUpdate)
	protected.DELETE("/subscriptions/:id", h.SubscriptionCancel)
	protected.GET("/subscriptions/upcoming", h.SubscriptionUpcoming)
	protected.GET("/security/score", h.SecurityScore)
	protected.GET("/security/audit", h.SecurityScore)
	protected.GET("/stats/overview", h.Overview)
	protected.GET("/stats/subscription", h.SubscriptionTrend)
	protected.GET("/footprints", h.FootprintsList)
	protected.POST("/footprints", h.FootprintCreate)
	protected.GET("/backups", h.BackupList)
	protected.POST("/backups", h.BackupCreate)
	protected.GET("/data-locations", h.DataLocationsList)
	protected.POST("/data-locations", h.DataLocationCreate)
	protected.DELETE("/data-locations/:id", h.DataLocationDelete)
	protected.GET("/categories", h.CategoriesList)
	protected.POST("/categories", h.CategoryCreate)
	protected.PUT("/categories/:id", h.CategoryUpdate)
	protected.DELETE("/categories/:id", h.CategoryDelete)
	protected.GET("/notifications", h.NotificationsList)
	protected.GET("/notifications/summary", h.NotificationsSummary)
	protected.POST("/notifications/refresh", h.NotificationsRefresh)
	protected.POST("/notifications/read-all", h.NotificationsReadAll)
	protected.POST("/notifications/:id/read", h.NotificationRead)
	protected.GET("/accounts/export.csv", h.AccountsCSVExport)
	protected.POST("/accounts/import.csv", h.AccountsCSVImport)
	protected.POST("/export", h.ExportData)
	protected.POST("/import", h.ImportData)
	return r
}
func cors(origin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

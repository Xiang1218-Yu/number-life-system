package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"number-life-system/internal/middleware"
	"number-life-system/internal/service"
	"strconv"
)

type Handler struct {
	Auth          *service.AuthService
	Accounts      *service.AccountService
	Subscriptions *service.SubscriptionService
	Footprints    *service.FootprintService
	Backups       *service.BackupService
	DataLocations *service.DataLocationService
	Categories    *service.CategoryService
	Notifications *service.NotificationService
	CSV           *service.CSVService
	Security      *service.SecurityService
	Stats         *service.StatsService
	Export        *service.ExportService
}

func (h *Handler) Register(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Name     string `json:"name" binding:"required"`
		Password string `json:"password" binding:"required,min=8"`
	}
	if !bind(c, &input) {
		return
	}
	result, err := h.Auth.Register(input.Email, input.Name, input.Password)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}
func (h *Handler) Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if !bind(c, &input) {
		return
	}
	result, err := h.Auth.Login(input.Email, input.Password)
	if err != nil {
		fail(c, http.StatusUnauthorized, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
func (h *Handler) AccountsList(c *gin.Context) {
	var twoFactor *bool
	if raw := c.Query("two_factor"); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err == nil {
			twoFactor = &value
		}
	}
	filter := service.AccountFilter{Page: service.NewPageRequest(c.Request.URL.Query()), Search: c.Query("search"), Category: c.Query("category"), Security: c.Query("security"), TwoFactor: twoFactor}
	result, err := h.Accounts.ListPage(middleware.UserID(c), c.Query("archived") == "true", filter)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
func (h *Handler) AccountGet(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	row, err := h.Accounts.Get(middleware.UserID(c), id)
	if err != nil {
		fail(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, row)
}
func (h *Handler) AccountCreate(c *gin.Context) {
	var input service.AccountInput
	if !bind(c, &input) {
		return
	}
	row, err := h.Accounts.Create(middleware.UserID(c), input)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusCreated, row)
}
func (h *Handler) AccountUpdate(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var input service.AccountInput
	if !bind(c, &input) {
		return
	}
	row, err := h.Accounts.Update(middleware.UserID(c), id, input)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, row)
}
func (h *Handler) AccountDelete(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := h.Accounts.Delete(middleware.UserID(c), id); err != nil {
		fail(c, http.StatusNotFound, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *Handler) SubscriptionsList(c *gin.Context) {
	filter := service.SubscriptionFilter{Page: service.NewPageRequest(c.Request.URL.Query()), Search: c.Query("search"), Status: c.Query("status"), Cycle: c.Query("cycle")}
	result, err := h.Subscriptions.ListPage(middleware.UserID(c), filter)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
func (h *Handler) SubscriptionCreate(c *gin.Context) {
	var input service.SubscriptionInput
	if !bind(c, &input) {
		return
	}
	row, err := h.Subscriptions.Create(middleware.UserID(c), input)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusCreated, row)
}
func (h *Handler) SubscriptionUpdate(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var input service.SubscriptionInput
	if !bind(c, &input) {
		return
	}
	row, err := h.Subscriptions.Update(middleware.UserID(c), id, input)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, row)
}
func (h *Handler) SubscriptionCancel(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := h.Subscriptions.Cancel(middleware.UserID(c), id); err != nil {
		fail(c, http.StatusNotFound, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *Handler) SubscriptionUpcoming(c *gin.Context) {
	days := 7
	if raw := c.Query("days"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err == nil && value > 0 {
			days = value
		}
	}
	rows, err := h.Subscriptions.Upcoming(middleware.UserID(c), days)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}
func (h *Handler) SecurityScore(c *gin.Context) {
	report, err := h.Security.Report(middleware.UserID(c))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, report)
}
func (h *Handler) Overview(c *gin.Context) {
	report, err := h.Stats.Overview(middleware.UserID(c))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, report)
}
func (h *Handler) SubscriptionTrend(c *gin.Context) {
	months := 6
	if raw := c.Query("months"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			months = parsed
		}
	}
	rows, err := h.Stats.SubscriptionTrend(middleware.UserID(c), months)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}
func (h *Handler) FootprintsList(c *gin.Context) {
	rows, err := h.Footprints.List(middleware.UserID(c))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}
func (h *Handler) FootprintCreate(c *gin.Context) {
	var input service.FootprintInput
	if !bind(c, &input) {
		return
	}
	row, err := h.Footprints.Create(middleware.UserID(c), input)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusCreated, row)
}
func (h *Handler) BackupList(c *gin.Context) {
	rows, err := h.Backups.List(middleware.UserID(c))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}
func (h *Handler) BackupCreate(c *gin.Context) {
	var input service.BackupInput
	if !bind(c, &input) {
		return
	}
	row, err := h.Backups.Create(middleware.UserID(c), input)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusCreated, row)
}
func (h *Handler) DataLocationsList(c *gin.Context) {
	rows, err := h.DataLocations.List(middleware.UserID(c))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}
func (h *Handler) DataLocationCreate(c *gin.Context) {
	var input service.DataLocationInput
	if !bind(c, &input) {
		return
	}
	row, err := h.DataLocations.Create(middleware.UserID(c), input)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusCreated, row)
}
func (h *Handler) DataLocationDelete(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := h.DataLocations.Delete(middleware.UserID(c), id); err != nil {
		fail(c, http.StatusNotFound, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *Handler) ExportData(c *gin.Context) {
	data, err := h.Export.Export(middleware.UserID(c))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	c.Header("Content-Disposition", "attachment; filename=digital-life.json")
	c.Data(http.StatusOK, "application/json; charset=utf-8", data)
}
func (h *Handler) ImportData(c *gin.Context) {
	data, err := c.GetRawData()
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	if err := h.Export.Import(middleware.UserID(c), data); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "导入成功"})
}
func bind(c *gin.Context, value any) bool {
	if err := c.ShouldBindJSON(value); err != nil {
		fail(c, http.StatusBadRequest, err)
		return false
	}
	return true
}
func pathID(c *gin.Context) (uint, bool) {
	value, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || value == 0 {
		if err == nil {
			err = errors.New("无效的资源编号")
		}
		fail(c, http.StatusBadRequest, err)
		return 0, false
	}
	return uint(value), true
}
func fail(c *gin.Context, status int, err error) { c.JSON(status, gin.H{"error": err.Error()}) }

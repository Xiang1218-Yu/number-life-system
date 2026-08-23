package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"number-life-system/internal/middleware"
	"number-life-system/internal/service"
)

func (h *Handler) CategoriesList(c *gin.Context) {
	rows, err := h.Categories.List()
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}

func (h *Handler) CategoryCreate(c *gin.Context) {
	var input service.CategoryInput
	if !bind(c, &input) {
		return
	}
	row, err := h.Categories.Create(input)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusCreated, row)
}

func (h *Handler) CategoryUpdate(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var input service.CategoryInput
	if !bind(c, &input) {
		return
	}
	row, err := h.Categories.Update(id, input)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, row)
}

func (h *Handler) CategoryDelete(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := h.Categories.Delete(id); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) NotificationsList(c *gin.Context) {
	limit := 50
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}
	filter := service.NotificationFilter{Status: c.Query("status"), Type: c.Query("type"), Limit: limit}
	rows, err := h.Notifications.List(middleware.UserID(c), filter)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}

func (h *Handler) NotificationsSummary(c *gin.Context) {
	summary, err := h.Notifications.Summary(middleware.UserID(c))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *Handler) NotificationsRefresh(c *gin.Context) {
	created, err := h.Notifications.Refresh(middleware.UserID(c))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"created": created})
}

func (h *Handler) NotificationRead(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := h.Notifications.MarkRead(middleware.UserID(c), id); err != nil {
		fail(c, http.StatusNotFound, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) NotificationsReadAll(c *gin.Context) {
	if err := h.Notifications.MarkAllRead(middleware.UserID(c)); err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) AccountsCSVExport(c *gin.Context) {
	data, err := h.CSV.ExportAccounts(middleware.UserID(c), c.Query("archived") == "true")
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	c.Header("Content-Disposition", "attachment; filename=accounts.csv")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

func (h *Handler) AccountsCSVImport(c *gin.Context) {
	data, err := c.GetRawData()
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	result, err := h.CSV.ImportAccounts(middleware.UserID(c), data)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

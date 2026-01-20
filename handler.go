package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	nft    *NFTManager
	cfg    *Config
	logger *log.Logger
}

// NewHandler creates a new Handler
func NewHandler(nft *NFTManager, cfg *Config, logger *log.Logger) *Handler {
	return &Handler{
		nft:    nft,
		cfg:    cfg,
		logger: logger,
	}
}

// ListQuotas handles GET /api/v1/quotas
func (h *Handler) ListQuotas(c echo.Context) error {
	quotas, err := h.nft.ListQuotas()
	if err != nil {
		h.logger.Printf("Error listing quotas: %v", err)
		return c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	allowedPorts, err := h.nft.ListAllowedPorts()
	if err != nil {
		h.logger.Printf("Error listing allowed ports: %v", err)
		// Non-fatal: continue without allowed ports
		allowedPorts = []AllowedPort{}
	}

	return c.JSON(http.StatusOK, QuotasResponse{
		Quotas:          quotas,
		AllowedPorts:    allowedPorts,
		ReadOnly:        h.cfg.ReadOnly,
		RefreshInterval: h.cfg.RefreshInterval,
	})
}

// ResetQuota handles POST /api/v1/quotas/:id/reset
func (h *Handler) ResetQuota(c echo.Context) error {
	id := c.Param("id")

	if err := h.nft.ResetQuota(id); err != nil {
		h.logger.Printf("Error resetting quota %s: %v", id, err)
		return c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	h.logger.Printf("Quota reset: %s", id)
	return c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Quota reset successfully",
	})
}

// BatchResetQuotas handles POST /api/v1/quotas/batch-reset
func (h *Handler) BatchResetQuotas(c echo.Context) error {
	var req BatchResetRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if len(req.IDs) == 0 {
		return c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "No IDs provided",
		})
	}

	if err := h.nft.BatchResetQuotas(req.IDs); err != nil {
		h.logger.Printf("Error batch resetting quotas: %v", err)
		return c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	h.logger.Printf("Batch reset %d quotas", len(req.IDs))
	return c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Quotas reset successfully",
	})
}

// ModifyQuota handles PUT /api/v1/quotas/:id
func (h *Handler) ModifyQuota(c echo.Context) error {
	id := c.Param("id")

	var req ModifyQuotaRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if req.Bytes <= 0 {
		return c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Quota limit must be positive",
		})
	}

	if err := h.nft.ModifyQuota(id, req.Bytes); err != nil {
		h.logger.Printf("Error modifying quota %s: %v", id, err)
		return c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	h.logger.Printf("Quota modified: %s to %d bytes", id, req.Bytes)
	return c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Quota modified successfully",
	})
}

// AddQuota handles POST /api/v1/quotas
func (h *Handler) AddQuota(c echo.Context) error {
	var req AddQuotaRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if req.Port < 1 || req.Port > 65535 {
		return c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Port must be between 1 and 65535",
		})
	}

	if req.Bytes <= 0 {
		return c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Quota limit must be positive",
		})
	}

	if err := h.nft.AddQuota(req.Port, req.Bytes, req.Comment); err != nil {
		h.logger.Printf("Error adding quota for port %d: %v", req.Port, err)
		return c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	h.logger.Printf("Quota added: port %d, limit %d bytes", req.Port, req.Bytes)
	return c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: "Quota added successfully",
	})
}

// DeleteQuota handles DELETE /api/v1/quotas/:id
func (h *Handler) DeleteQuota(c echo.Context) error {
	id := c.Param("id")

	if err := h.nft.DeleteQuota(id); err != nil {
		h.logger.Printf("Error deleting quota %s: %v", id, err)
		return c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	h.logger.Printf("Quota deleted: %s", id)
	return c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Quota deleted successfully",
	})
}

// AddPort handles POST /api/v1/ports
func (h *Handler) AddPort(c echo.Context) error {
	var req AddPortRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if req.Port < 1 || req.Port > 65535 {
		return c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Port must be between 1 and 65535",
		})
	}

	if err := h.nft.AddAllowedPort(req.Port); err != nil {
		h.logger.Printf("Error adding allowed port %d: %v", req.Port, err)
		return c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	h.logger.Printf("Allowed port added: %d", req.Port)
	return c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: "Port added successfully",
	})
}

// DeletePort handles DELETE /api/v1/ports/:handle
func (h *Handler) DeletePort(c echo.Context) error {
	handleStr := c.Param("handle")
	handle, err := strconv.ParseInt(handleStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid handle",
		})
	}

	if err := h.nft.DeleteAllowedPort(handle); err != nil {
		h.logger.Printf("Error deleting allowed port handle %d: %v", handle, err)
		return c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	h.logger.Printf("Allowed port deleted: handle %d", handle)
	return c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Port deleted successfully",
	})
}

package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	nft      *NFTManager
	fwd      *ForwardingManager
	cfg      *Config
	logger   *log.Logger
	tokenGen *TokenGenerator
}

// NewHandler creates a new Handler
func NewHandler(nft *NFTManager, fwd *ForwardingManager, cfg *Config, logger *log.Logger, tokenGen *TokenGenerator) *Handler {
	return &Handler{
		nft:      nft,
		fwd:      fwd,
		cfg:      cfg,
		logger:   logger,
		tokenGen: tokenGen,
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

// ListQuotasWithTokens handles GET /api/v1/quotas when tokens are enabled
// Returns quotas with their query tokens for the admin panel
func (h *Handler) ListQuotasWithTokens(c echo.Context) error {
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
		allowedPorts = []AllowedPort{}
	}

	// Add tokens to quotas
	quotasWithTokens := make([]QuotaWithToken, len(quotas))
	for i, q := range quotas {
		quotasWithTokens[i] = QuotaWithToken{
			QuotaRule: q,
			Token:     h.tokenGen.Generate(q.Port),
		}
	}

	return c.JSON(http.StatusOK, QuotasResponseWithTokens{
		Quotas:          quotasWithTokens,
		AllowedPorts:    allowedPorts,
		ReadOnly:        h.cfg.ReadOnly,
		RefreshInterval: h.cfg.RefreshInterval,
	})
}

// QueryByToken handles GET /api/v1/public/query/:token (NO AUTH REQUIRED)
// Allows users to query quota usage with a token
func (h *Handler) QueryByToken(c echo.Context) error {
	token := c.Param("token")

	// Validate token format
	if !IsValidTokenFormat(token) {
		return c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid token format",
		})
	}

	// Get all quotas
	quotas, err := h.nft.ListQuotas()
	if err != nil {
		h.logger.Printf("Error listing quotas for token query: %v", err)
		return c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Internal server error",
		})
	}

	// Find matching quota
	quota := h.tokenGen.FindQuotaByToken(token, quotas)
	if quota == nil {
		// Return generic error to prevent enumeration
		return c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Error:   "Token not found",
		})
	}

	return c.JSON(http.StatusOK, PublicQueryResponse{
		Port:         quota.Port,
		UsedBytes:    quota.UsedBytes,
		QuotaBytes:   quota.QuotaBytes,
		UsagePercent: quota.UsagePercent,
		Status:       quota.Status,
		Comment:      quota.Comment,
	})
}

// ListForwarding handles GET /api/v1/forwarding
func (h *Handler) ListForwarding(c echo.Context) error {
	rules, err := h.fwd.ListForwardingRules()
	if err != nil {
		h.logger.Printf("Error listing forwarding rules: %v", err)
		return c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, ForwardingResponse{
		Rules:    rules,
		ReadOnly: h.cfg.ReadOnly,
	})
}

// AddForwarding handles POST /api/v1/forwarding
func (h *Handler) AddForwarding(c echo.Context) error {
	var req AddForwardingRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if req.SrcPort < 1 || req.SrcPort > 65535 {
		return c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Source port must be between 1 and 65535",
		})
	}

	if req.DstPort < 1 || req.DstPort > 65535 {
		return c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Destination port must be between 1 and 65535",
		})
	}

	if req.Protocol != "tcp" && req.Protocol != "udp" && req.Protocol != "both" {
		return c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Protocol must be 'tcp', 'udp', or 'both'",
		})
	}

	if req.LimitMbps < 0 {
		return c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Limit must be >= 0 (0 = no limit)",
		})
	}

	if err := h.fwd.AddForwardingRule(req.SrcPort, req.DstIP, req.DstPort, req.Protocol, req.Comment, req.LimitMbps); err != nil {
		h.logger.Printf("Error adding forwarding rule: %v", err)
		return c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	h.logger.Printf("Forwarding rule added: %d -> %s:%d (%s) limit=%d Mbps", req.SrcPort, req.DstIP, req.DstPort, req.Protocol, req.LimitMbps)
	return c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: "Forwarding rule added successfully",
	})
}

// EditForwarding handles PUT /api/v1/forwarding/:id
func (h *Handler) EditForwarding(c echo.Context) error {
	id := c.Param("id")

	var req EditForwardingRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if req.DstPort < 1 || req.DstPort > 65535 {
		return c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Destination port must be between 1 and 65535",
		})
	}

	if req.Protocol != "tcp" && req.Protocol != "udp" && req.Protocol != "both" {
		return c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Protocol must be 'tcp', 'udp', or 'both'",
		})
	}

	if req.LimitMbps < 0 {
		return c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Limit must be >= 0 (0 = no limit)",
		})
	}

	if err := h.fwd.EditForwardingRule(id, req.DstIP, req.DstPort, req.Protocol, req.Comment, req.LimitMbps); err != nil {
		h.logger.Printf("Error editing forwarding rule %s: %v", id, err)
		return c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	h.logger.Printf("Forwarding rule edited: %s -> %s:%d (%s) limit=%d Mbps", id, req.DstIP, req.DstPort, req.Protocol, req.LimitMbps)
	return c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Forwarding rule updated successfully",
	})
}

// DeleteForwarding handles DELETE /api/v1/forwarding/:id
func (h *Handler) DeleteForwarding(c echo.Context) error {
	id := c.Param("id")

	if err := h.fwd.DeleteForwardingRule(id); err != nil {
		h.logger.Printf("Error deleting forwarding rule %s: %v", id, err)
		return c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	h.logger.Printf("Forwarding rule deleted: %s", id)
	return c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Forwarding rule deleted successfully",
	})
}

// EnableForwarding handles POST /api/v1/forwarding/:id/enable
func (h *Handler) EnableForwarding(c echo.Context) error {
	id := c.Param("id")

	if err := h.fwd.EnableForwardingRule(id); err != nil {
		h.logger.Printf("Error enabling forwarding rule %s: %v", id, err)
		return c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	h.logger.Printf("Forwarding rule enabled: %s", id)
	return c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Forwarding rule enabled successfully",
	})
}

// DisableForwarding handles POST /api/v1/forwarding/:id/disable
func (h *Handler) DisableForwarding(c echo.Context) error {
	id := c.Param("id")

	if err := h.fwd.DisableForwardingRule(id); err != nil {
		h.logger.Printf("Error disabling forwarding rule %s: %v", id, err)
		return c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	h.logger.Printf("Forwarding rule disabled: %s", id)
	return c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Forwarding rule disabled successfully",
	})
}

// GetRawRuleset handles GET /api/v1/raw-ruleset
func (h *Handler) GetRawRuleset(c echo.Context) error {
	rawData, err := h.nft.GetRawRuleset()
	if err != nil {
		h.logger.Printf("Error getting raw ruleset: %v", err)
		return c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    rawData,
	})
}

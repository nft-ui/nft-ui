package main

// QuotaRule represents a parsed nftables quota rule
type QuotaRule struct {
	ID           string  `json:"id"`            // inet_filter_output_<handle>
	Handle       int64   `json:"handle"`        // nft handle for deletion
	Port         int     `json:"port"`          // source port
	QuotaBytes   int64   `json:"quota_bytes"`   // quota limit in bytes
	UsedBytes    int64   `json:"used_bytes"`    // current usage in bytes
	UsagePercent float64 `json:"usage_percent"` // calculated: used/quota * 100
	Status       string  `json:"status"`        // "ok" | "warning" | "exceeded"
	Comment      string  `json:"comment"`       // rule comment
}

// AllowedPort represents an allowed inbound port from the input chain
type AllowedPort struct {
	Port    int    `json:"port"`
	Handle  int64  `json:"handle"`
	Managed bool   `json:"managed"` // true if comment == "nft-ui managed"
	Comment string `json:"comment,omitempty"`
}

// AddPortRequest is the request body for adding a new allowed port
type AddPortRequest struct {
	Port int `json:"port"`
}

// QuotasResponse is the API response for listing quotas
type QuotasResponse struct {
	Quotas          []QuotaRule   `json:"quotas"`
	AllowedPorts    []AllowedPort `json:"allowed_ports"`
	ReadOnly        bool          `json:"read_only"`
	RefreshInterval int           `json:"refresh_interval"`
}

// AddQuotaRequest is the request body for adding a new quota
type AddQuotaRequest struct {
	Port    int    `json:"port"`
	Bytes   int64  `json:"bytes"`
	Comment string `json:"comment"`
}

// ModifyQuotaRequest is the request body for modifying a quota
type ModifyQuotaRequest struct {
	Bytes int64 `json:"bytes"`
}

// BatchResetRequest is the request body for batch resetting quotas
type BatchResetRequest struct {
	IDs []string `json:"ids"`
}

// APIResponse is a generic API response
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// NFT JSON structures for parsing nft -j output

// NFTRuleset is the top-level structure from nft -j list chain
type NFTRuleset struct {
	NFTables []NFTObject `json:"nftables"`
}

// NFTObject wraps different nftables objects
type NFTObject struct {
	Metainfo *NFTMetainfo `json:"metainfo,omitempty"`
	Chain    *NFTChain    `json:"chain,omitempty"`
	Rule     *NFTRule     `json:"rule,omitempty"`
}

// NFTMetainfo contains nftables version info
type NFTMetainfo struct {
	Version string `json:"version"`
}

// NFTChain represents an nftables chain
type NFTChain struct {
	Family string `json:"family"`
	Table  string `json:"table"`
	Name   string `json:"name"`
	Handle int64  `json:"handle"`
	Type   string `json:"type"`
	Hook   string `json:"hook"`
	Prio   int    `json:"prio"`
	Policy string `json:"policy"`
}

// NFTRule represents an nftables rule
type NFTRule struct {
	Family  string                   `json:"family"`
	Table   string                   `json:"table"`
	Chain   string                   `json:"chain"`
	Handle  int64                    `json:"handle"`
	Expr    []map[string]interface{} `json:"expr"`
	Comment string                   `json:"comment,omitempty"`
}

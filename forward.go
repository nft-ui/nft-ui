package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ForwardingComment is the prefix used to identify forwarding rules managed by nft-ui
const ForwardingComment = "nft-ui fwd"

// ForwardingManager handles port forwarding (DNAT + MASQUERADE) operations
type ForwardingManager struct {
	mu                  sync.Mutex
	binary              string
	disabledForwardsPath string
}

// NewForwardingManager creates a new ForwardingManager
func NewForwardingManager(cfg *Config) *ForwardingManager {
	path := cfg.DisabledForwardsPath
	if path == "" {
		path = "/var/lib/nft-ui/disabled-forwards.json"
	}
	return &ForwardingManager{
		binary:              cfg.NFTBinary,
		disabledForwardsPath: path,
	}
}

// execNFT executes an nft command and returns the output
func (m *ForwardingManager) execNFT(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, m.binary, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("nft %s: %w (output: %s)", strings.Join(args, " "), err, string(output))
	}
	return output, nil
}

// EnsureNatSetup ensures the nat table and required chains exist
func (m *ForwardingManager) EnsureNatSetup() error {
	// Check if nat table exists by trying to list it
	_, err := m.execNFT("list", "table", "ip", "nat")
	if err != nil {
		// Table doesn't exist, create it
		if _, err := m.execNFT("add", "table", "ip", "nat"); err != nil {
			return fmt.Errorf("failed to create nat table: %w", err)
		}
	}

	// Check if prerouting chain exists
	_, err = m.execNFT("list", "chain", "ip", "nat", "prerouting")
	if err != nil {
		// Create prerouting chain
		if _, err := m.execNFT("add", "chain", "ip", "nat", "prerouting",
			"{ type nat hook prerouting priority dstnat ; policy accept ; }"); err != nil {
			return fmt.Errorf("failed to create prerouting chain: %w", err)
		}
	}

	// Check if postrouting chain exists
	_, err = m.execNFT("list", "chain", "ip", "nat", "postrouting")
	if err != nil {
		// Create postrouting chain
		if _, err := m.execNFT("add", "chain", "ip", "nat", "postrouting",
			"{ type nat hook postrouting priority srcnat ; policy accept ; }"); err != nil {
			return fmt.Errorf("failed to create postrouting chain: %w", err)
		}
	}

	// Check if output chain exists
	_, err = m.execNFT("list", "chain", "ip", "nat", "output")
	if err != nil {
		// Create output chain
		if _, err := m.execNFT("add", "chain", "ip", "nat", "output",
			"{ type nat hook output priority dstnat ; policy accept ; }"); err != nil {
			return fmt.Errorf("failed to create output chain: %w", err)
		}
	}

	return nil
}

// ListForwardingRules returns all forwarding rules (enabled from nftables + disabled from file)
func (m *ForwardingManager) ListForwardingRules() ([]ForwardingRule, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure nat table exists
	if err := m.EnsureNatSetup(); err != nil {
		return nil, err
	}

	// Get prerouting rules (DNAT)
	preOutput, err := m.execNFT("-j", "-a", "list", "chain", "ip", "nat", "prerouting")
	if err != nil {
		return nil, fmt.Errorf("failed to list prerouting chain: %w", err)
	}

	// Get postrouting rules (MASQUERADE)
	postOutput, err := m.execNFT("-j", "-a", "list", "chain", "ip", "nat", "postrouting")
	if err != nil {
		return nil, fmt.Errorf("failed to list postrouting chain: %w", err)
	}

	// Parse enabled rules from nftables
	enabledRules, err := m.parseForwardingRules(preOutput, postOutput)
	if err != nil {
		return nil, err
	}

	// Load disabled rules from file
	disabledRules, err := m.loadDisabledRules()
	if err != nil {
		// If file doesn't exist, that's fine
		disabledRules = []ForwardingRule{}
	}

	// Merge enabled and disabled rules
	allRules := append(enabledRules, disabledRules...)
	return allRules, nil
}

// parseForwardingRules parses JSON output from prerouting and postrouting chains
func (m *ForwardingManager) parseForwardingRules(preData, postData []byte) ([]ForwardingRule, error) {
	var preRuleset, postRuleset NFTRuleset
	if err := json.Unmarshal(preData, &preRuleset); err != nil {
		return nil, fmt.Errorf("failed to parse prerouting JSON: %w", err)
	}
	if err := json.Unmarshal(postData, &postRuleset); err != nil {
		return nil, fmt.Errorf("failed to parse postrouting JSON: %w", err)
	}

	// Build a map of postrouting handles by srcPort for managed rules
	postHandles := make(map[int]int64)
	for _, obj := range postRuleset.NFTables {
		if obj.Rule == nil || obj.Rule.Chain != "postrouting" {
			continue
		}
		if strings.HasPrefix(obj.Rule.Comment, ForwardingComment) {
			srcPort := m.extractSrcPortFromComment(obj.Rule.Comment)
			if srcPort > 0 {
				postHandles[srcPort] = obj.Rule.Handle
			}
		}
	}

	var rules []ForwardingRule

	// Parse ALL prerouting rules that have DNAT
	for _, obj := range preRuleset.NFTables {
		if obj.Rule == nil || obj.Rule.Chain != "prerouting" {
			continue
		}

		rule := m.extractForwardingRule(obj.Rule)
		if rule != nil {
			rule.Enabled = true
			rule.PreHandle = obj.Rule.Handle

			// Check if this is a managed rule (has our comment)
			if strings.HasPrefix(obj.Rule.Comment, ForwardingComment) {
				rule.Managed = true
				// Try to find matching postrouting handle
				if handle, ok := postHandles[rule.SrcPort]; ok {
					rule.PostHandle = handle
				}
			} else {
				rule.Managed = false
			}

			rules = append(rules, *rule)
		}
	}

	return rules, nil
}

// extractSrcPortFromComment extracts source port from comment like "nft-ui fwd 12103 some comment"
func (m *ForwardingManager) extractSrcPortFromComment(comment string) int {
	parts := strings.Fields(comment)
	if len(parts) >= 3 {
		port, err := strconv.Atoi(parts[2])
		if err == nil {
			return port
		}
	}
	return 0
}

// extractForwardingRule extracts forwarding rule info from a prerouting DNAT rule
func (m *ForwardingManager) extractForwardingRule(rule *NFTRule) *ForwardingRule {
	var srcPort, dstPort int
	var dstIP string
	var protocol string

	for _, expr := range rule.Expr {
		// Look for protocol meta match
		if metaData, ok := expr["match"]; ok {
			if mm, ok := metaData.(map[string]interface{}); ok {
				if left, ok := mm["left"].(map[string]interface{}); ok {
					if meta, ok := left["meta"].(map[string]interface{}); ok {
						if key, ok := meta["key"].(string); ok && key == "l4proto" {
							// Check for protocol set
							if right, ok := mm["right"].(map[string]interface{}); ok {
								if set, ok := right["set"].([]interface{}); ok {
									hasTCP := false
									hasUDP := false
									for _, p := range set {
										if ps, ok := p.(string); ok {
											if ps == "tcp" {
												hasTCP = true
											} else if ps == "udp" {
												hasUDP = true
											}
										}
									}
									if hasTCP && hasUDP {
										protocol = "both"
									} else if hasTCP {
										protocol = "tcp"
									} else if hasUDP {
										protocol = "udp"
									}
								}
							}
						}
					}
					// Check for payload match (dport)
					if payload, ok := left["payload"].(map[string]interface{}); ok {
						if field, ok := payload["field"].(string); ok && field == "dport" {
							if right, ok := mm["right"].(float64); ok {
								srcPort = int(right)
							}
						}
					}
				}
			}
		}

		// Look for DNAT expression
		if dnatData, ok := expr["dnat"]; ok {
			if dm, ok := dnatData.(map[string]interface{}); ok {
				if addr, ok := dm["addr"].(string); ok {
					dstIP = addr
				}
				if port, ok := dm["port"].(float64); ok {
					dstPort = int(port)
				}
			}
		}
	}

	// If we couldn't determine protocol, try from the rule structure
	if protocol == "" {
		for _, expr := range rule.Expr {
			if matchData, ok := expr["match"]; ok {
				if mm, ok := matchData.(map[string]interface{}); ok {
					if left, ok := mm["left"].(map[string]interface{}); ok {
						if payload, ok := left["payload"].(map[string]interface{}); ok {
							if proto, ok := payload["protocol"].(string); ok {
								protocol = proto
							}
						}
					}
				}
			}
		}
	}

	if protocol == "" {
		protocol = "both" // Default to both if not determined
	}

	if srcPort == 0 || dstIP == "" || dstPort == 0 {
		return nil
	}

	// Extract user comment based on whether it's a managed rule
	userComment := ""
	if strings.HasPrefix(rule.Comment, ForwardingComment) {
		// Managed rule: extract part after "nft-ui fwd <port>"
		parts := strings.Fields(rule.Comment)
		if len(parts) > 3 {
			userComment = strings.Join(parts[3:], " ")
		}
	} else {
		// Unmanaged rule: use the raw comment
		userComment = rule.Comment
	}

	return &ForwardingRule{
		ID:       fmt.Sprintf("fwd_%d", srcPort),
		SrcPort:  srcPort,
		DstIP:    dstIP,
		DstPort:  dstPort,
		Protocol: protocol,
		Comment:  userComment,
	}
}

// AddForwardingRule adds a new port forwarding rule
func (m *ForwardingManager) AddForwardingRule(srcPort int, dstIP string, dstPort int, protocol string, comment string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate inputs
	if srcPort < 1 || srcPort > 65535 {
		return fmt.Errorf("invalid source port: %d", srcPort)
	}
	if dstPort < 1 || dstPort > 65535 {
		return fmt.Errorf("invalid destination port: %d", dstPort)
	}
	if !isValidIPv4(dstIP) {
		return fmt.Errorf("invalid destination IP: %s", dstIP)
	}
	if protocol != "tcp" && protocol != "udp" && protocol != "both" {
		return fmt.Errorf("invalid protocol: %s", protocol)
	}

	// Sanitize comment
	comment = sanitizeComment(comment)

	// Ensure nat table exists
	if err := m.EnsureNatSetup(); err != nil {
		return err
	}

	// Check for duplicate source port
	preOutput, _ := m.execNFT("-j", "-a", "list", "chain", "ip", "nat", "prerouting")
	postOutput, _ := m.execNFT("-j", "-a", "list", "chain", "ip", "nat", "postrouting")
	existingRules, _ := m.parseForwardingRules(preOutput, postOutput)
	for _, r := range existingRules {
		if r.SrcPort == srcPort {
			return fmt.Errorf("source port %d is already in use", srcPort)
		}
	}

	// Also check disabled rules
	disabledRules, _ := m.loadDisabledRules()
	for _, r := range disabledRules {
		if r.SrcPort == srcPort {
			return fmt.Errorf("source port %d is already in use (disabled rule)", srcPort)
		}
	}

	// Build comment string
	fullComment := fmt.Sprintf("%s %d", ForwardingComment, srcPort)
	if comment != "" {
		fullComment = fmt.Sprintf("%s %s", fullComment, comment)
	}

	// Add prerouting DNAT rule
	if err := m.addDNATRule(srcPort, dstIP, dstPort, protocol, fullComment); err != nil {
		return fmt.Errorf("failed to add DNAT rule: %w", err)
	}

	// Add postrouting MASQUERADE rule
	if err := m.addMasqueradeRule(dstIP, dstPort, protocol, fullComment); err != nil {
		// Rollback: delete the DNAT rule
		m.deleteDNATRuleBySrcPort(srcPort)
		return fmt.Errorf("failed to add MASQUERADE rule: %w", err)
	}

	// Add output DNAT rule for local traffic
	if err := m.addOutputDNATRule(srcPort, dstIP, dstPort, protocol, fullComment); err != nil {
		// Rollback: delete previous rules
		m.deleteDNATRuleBySrcPort(srcPort)
		m.deleteMasqueradeRuleBySrcPort(srcPort)
		return fmt.Errorf("failed to add output DNAT rule: %w", err)
	}

	return nil
}

// addDNATRule adds a prerouting DNAT rule
func (m *ForwardingManager) addDNATRule(srcPort int, dstIP string, dstPort int, protocol string, comment string) error {
	var args []string

	switch protocol {
	case "tcp":
		args = []string{
			"add", "rule", "ip", "nat", "prerouting",
			"tcp", "dport", strconv.Itoa(srcPort),
			"dnat", "to", fmt.Sprintf("%s:%d", dstIP, dstPort),
			"comment", fmt.Sprintf(`"%s"`, comment),
		}
	case "udp":
		args = []string{
			"add", "rule", "ip", "nat", "prerouting",
			"udp", "dport", strconv.Itoa(srcPort),
			"dnat", "to", fmt.Sprintf("%s:%d", dstIP, dstPort),
			"comment", fmt.Sprintf(`"%s"`, comment),
		}
	default: // "both"
		args = []string{
			"add", "rule", "ip", "nat", "prerouting",
			"meta", "l4proto", "{", "tcp,", "udp", "}",
			"th", "dport", strconv.Itoa(srcPort),
			"dnat", "to", fmt.Sprintf("%s:%d", dstIP, dstPort),
			"comment", fmt.Sprintf(`"%s"`, comment),
		}
	}

	_, err := m.execNFT(args...)
	return err
}

// addMasqueradeRule adds a postrouting MASQUERADE rule
func (m *ForwardingManager) addMasqueradeRule(dstIP string, dstPort int, protocol string, comment string) error {
	var args []string

	switch protocol {
	case "tcp":
		args = []string{
			"add", "rule", "ip", "nat", "postrouting",
			"ip", "daddr", dstIP,
			"tcp", "dport", strconv.Itoa(dstPort),
			"masquerade",
			"comment", fmt.Sprintf(`"%s"`, comment),
		}
	case "udp":
		args = []string{
			"add", "rule", "ip", "nat", "postrouting",
			"ip", "daddr", dstIP,
			"udp", "dport", strconv.Itoa(dstPort),
			"masquerade",
			"comment", fmt.Sprintf(`"%s"`, comment),
		}
	default: // "both"
		args = []string{
			"add", "rule", "ip", "nat", "postrouting",
			"ip", "daddr", dstIP,
			"meta", "l4proto", "{", "tcp,", "udp", "}",
			"th", "dport", strconv.Itoa(dstPort),
			"masquerade",
			"comment", fmt.Sprintf(`"%s"`, comment),
		}
	}

	_, err := m.execNFT(args...)
	return err
}

// addOutputDNATRule adds an output chain DNAT rule for local traffic
func (m *ForwardingManager) addOutputDNATRule(srcPort int, dstIP string, dstPort int, protocol string, comment string) error {
	var args []string

	switch protocol {
	case "tcp":
		args = []string{
			"add", "rule", "ip", "nat", "output",
			"tcp", "dport", strconv.Itoa(srcPort),
			"dnat", "to", fmt.Sprintf("%s:%d", dstIP, dstPort),
			"comment", fmt.Sprintf(`"%s"`, comment),
		}
	case "udp":
		args = []string{
			"add", "rule", "ip", "nat", "output",
			"udp", "dport", strconv.Itoa(srcPort),
			"dnat", "to", fmt.Sprintf("%s:%d", dstIP, dstPort),
			"comment", fmt.Sprintf(`"%s"`, comment),
		}
	default: // "both"
		args = []string{
			"add", "rule", "ip", "nat", "output",
			"meta", "l4proto", "{", "tcp,", "udp", "}",
			"th", "dport", strconv.Itoa(srcPort),
			"dnat", "to", fmt.Sprintf("%s:%d", dstIP, dstPort),
			"comment", fmt.Sprintf(`"%s"`, comment),
		}
	}

	_, err := m.execNFT(args...)
	return err
}

// DeleteForwardingRule deletes a forwarding rule by ID
func (m *ForwardingManager) DeleteForwardingRule(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Parse source port from ID
	srcPort, err := m.parseSrcPortFromID(id)
	if err != nil {
		return err
	}

	// Check if it's a disabled rule
	disabledRules, _ := m.loadDisabledRules()
	for i, r := range disabledRules {
		if r.SrcPort == srcPort {
			// Remove from disabled rules
			disabledRules = append(disabledRules[:i], disabledRules[i+1:]...)
			return m.saveDisabledRules(disabledRules)
		}
	}

	// It's an enabled rule, delete from nftables
	if err := m.deleteDNATRuleBySrcPort(srcPort); err != nil {
		return fmt.Errorf("failed to delete DNAT rule: %w", err)
	}

	if err := m.deleteMasqueradeRuleBySrcPort(srcPort); err != nil {
		// Log warning but don't fail
		fmt.Printf("Warning: failed to delete MASQUERADE rule: %v\n", err)
	}

	if err := m.deleteOutputDNATRuleBySrcPort(srcPort); err != nil {
		// Log warning but don't fail
		fmt.Printf("Warning: failed to delete output DNAT rule: %v\n", err)
	}

	return nil
}

// EditForwardingRule modifies an existing forwarding rule
func (m *ForwardingManager) EditForwardingRule(id string, dstIP string, dstPort int, protocol string, comment string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Parse source port from ID
	srcPort, err := m.parseSrcPortFromID(id)
	if err != nil {
		return err
	}

	// Validate inputs
	if dstPort < 1 || dstPort > 65535 {
		return fmt.Errorf("invalid destination port: %d", dstPort)
	}
	if !isValidIPv4(dstIP) {
		return fmt.Errorf("invalid destination IP: %s", dstIP)
	}
	if protocol != "tcp" && protocol != "udp" && protocol != "both" {
		return fmt.Errorf("invalid protocol: %s", protocol)
	}

	comment = sanitizeComment(comment)

	// Check if it's a disabled rule
	disabledRules, _ := m.loadDisabledRules()
	for i, r := range disabledRules {
		if r.SrcPort == srcPort {
			// Update the disabled rule
			disabledRules[i].DstIP = dstIP
			disabledRules[i].DstPort = dstPort
			disabledRules[i].Protocol = protocol
			disabledRules[i].Comment = comment
			return m.saveDisabledRules(disabledRules)
		}
	}

	// It's an enabled rule - delete and recreate
	// Delete existing rules
	m.deleteDNATRuleBySrcPort(srcPort)
	m.deleteMasqueradeRuleBySrcPort(srcPort)
	m.deleteOutputDNATRuleBySrcPort(srcPort)

	// Build comment string
	fullComment := fmt.Sprintf("%s %d", ForwardingComment, srcPort)
	if comment != "" {
		fullComment = fmt.Sprintf("%s %s", fullComment, comment)
	}

	// Add new rules
	if err := m.addDNATRule(srcPort, dstIP, dstPort, protocol, fullComment); err != nil {
		return fmt.Errorf("failed to add DNAT rule: %w", err)
	}

	if err := m.addMasqueradeRule(dstIP, dstPort, protocol, fullComment); err != nil {
		m.deleteDNATRuleBySrcPort(srcPort)
		return fmt.Errorf("failed to add MASQUERADE rule: %w", err)
	}

	if err := m.addOutputDNATRule(srcPort, dstIP, dstPort, protocol, fullComment); err != nil {
		m.deleteDNATRuleBySrcPort(srcPort)
		m.deleteMasqueradeRuleBySrcPort(srcPort)
		return fmt.Errorf("failed to add output DNAT rule: %w", err)
	}

	return nil
}

// EnableForwardingRule enables a disabled forwarding rule
func (m *ForwardingManager) EnableForwardingRule(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	srcPort, err := m.parseSrcPortFromID(id)
	if err != nil {
		return err
	}

	// Find the disabled rule
	disabledRules, err := m.loadDisabledRules()
	if err != nil {
		return fmt.Errorf("failed to load disabled rules: %w", err)
	}

	var rule *ForwardingRule
	var idx int
	for i, r := range disabledRules {
		if r.SrcPort == srcPort {
			rule = &disabledRules[i]
			idx = i
			break
		}
	}

	if rule == nil {
		return fmt.Errorf("disabled rule not found: %s", id)
	}

	// Ensure nat table exists
	if err := m.EnsureNatSetup(); err != nil {
		return err
	}

	// Build comment string
	fullComment := fmt.Sprintf("%s %d", ForwardingComment, rule.SrcPort)
	if rule.Comment != "" {
		fullComment = fmt.Sprintf("%s %s", fullComment, rule.Comment)
	}

	// Create nftables rules
	if err := m.addDNATRule(rule.SrcPort, rule.DstIP, rule.DstPort, rule.Protocol, fullComment); err != nil {
		return fmt.Errorf("failed to add DNAT rule: %w", err)
	}

	if err := m.addMasqueradeRule(rule.DstIP, rule.DstPort, rule.Protocol, fullComment); err != nil {
		m.deleteDNATRuleBySrcPort(rule.SrcPort)
		return fmt.Errorf("failed to add MASQUERADE rule: %w", err)
	}

	if err := m.addOutputDNATRule(rule.SrcPort, rule.DstIP, rule.DstPort, rule.Protocol, fullComment); err != nil {
		m.deleteDNATRuleBySrcPort(rule.SrcPort)
		m.deleteMasqueradeRuleBySrcPort(rule.SrcPort)
		return fmt.Errorf("failed to add output DNAT rule: %w", err)
	}

	// Remove from disabled rules
	disabledRules = append(disabledRules[:idx], disabledRules[idx+1:]...)
	return m.saveDisabledRules(disabledRules)
}

// DisableForwardingRule disables an enabled forwarding rule
func (m *ForwardingManager) DisableForwardingRule(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	srcPort, err := m.parseSrcPortFromID(id)
	if err != nil {
		return err
	}

	// Get current enabled rules
	preOutput, _ := m.execNFT("-j", "-a", "list", "chain", "ip", "nat", "prerouting")
	postOutput, _ := m.execNFT("-j", "-a", "list", "chain", "ip", "nat", "postrouting")
	enabledRules, err := m.parseForwardingRules(preOutput, postOutput)
	if err != nil {
		return err
	}

	// Find the rule to disable
	var rule *ForwardingRule
	for i, r := range enabledRules {
		if r.SrcPort == srcPort {
			rule = &enabledRules[i]
			break
		}
	}

	if rule == nil {
		return fmt.Errorf("enabled rule not found: %s", id)
	}

	// Delete from nftables
	if err := m.deleteDNATRuleBySrcPort(srcPort); err != nil {
		return fmt.Errorf("failed to delete DNAT rule: %w", err)
	}
	m.deleteMasqueradeRuleBySrcPort(srcPort) // Ignore errors
	m.deleteOutputDNATRuleBySrcPort(srcPort) // Ignore errors

	// Save to disabled rules
	disabledRules, _ := m.loadDisabledRules()
	disabledRule := ForwardingRule{
		ID:       rule.ID,
		SrcPort:  rule.SrcPort,
		DstIP:    rule.DstIP,
		DstPort:  rule.DstPort,
		Protocol: rule.Protocol,
		Enabled:  false,
		Comment:  rule.Comment,
	}
	disabledRules = append(disabledRules, disabledRule)
	return m.saveDisabledRules(disabledRules)
}

// Helper functions

func (m *ForwardingManager) parseSrcPortFromID(id string) (int, error) {
	if !strings.HasPrefix(id, "fwd_") {
		return 0, fmt.Errorf("invalid forwarding rule ID: %s", id)
	}
	portStr := strings.TrimPrefix(id, "fwd_")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("invalid forwarding rule ID: %s", id)
	}
	return port, nil
}

func (m *ForwardingManager) deleteDNATRuleBySrcPort(srcPort int) error {
	output, err := m.execNFT("-j", "-a", "list", "chain", "ip", "nat", "prerouting")
	if err != nil {
		return err
	}

	var ruleset NFTRuleset
	if err := json.Unmarshal(output, &ruleset); err != nil {
		return err
	}

	for _, obj := range ruleset.NFTables {
		if obj.Rule == nil || obj.Rule.Chain != "prerouting" {
			continue
		}
		if strings.HasPrefix(obj.Rule.Comment, fmt.Sprintf("%s %d", ForwardingComment, srcPort)) {
			_, err := m.execNFT("delete", "rule", "ip", "nat", "prerouting", "handle", strconv.FormatInt(obj.Rule.Handle, 10))
			return err
		}
	}

	return errors.New("DNAT rule not found")
}

func (m *ForwardingManager) deleteMasqueradeRuleBySrcPort(srcPort int) error {
	output, err := m.execNFT("-j", "-a", "list", "chain", "ip", "nat", "postrouting")
	if err != nil {
		return err
	}

	var ruleset NFTRuleset
	if err := json.Unmarshal(output, &ruleset); err != nil {
		return err
	}

	for _, obj := range ruleset.NFTables {
		if obj.Rule == nil || obj.Rule.Chain != "postrouting" {
			continue
		}
		if strings.HasPrefix(obj.Rule.Comment, fmt.Sprintf("%s %d", ForwardingComment, srcPort)) {
			_, err := m.execNFT("delete", "rule", "ip", "nat", "postrouting", "handle", strconv.FormatInt(obj.Rule.Handle, 10))
			return err
		}
	}

	return errors.New("MASQUERADE rule not found")
}

func (m *ForwardingManager) deleteOutputDNATRuleBySrcPort(srcPort int) error {
	output, err := m.execNFT("-j", "-a", "list", "chain", "ip", "nat", "output")
	if err != nil {
		return err
	}

	var ruleset NFTRuleset
	if err := json.Unmarshal(output, &ruleset); err != nil {
		return err
	}

	for _, obj := range ruleset.NFTables {
		if obj.Rule == nil || obj.Rule.Chain != "output" {
			continue
		}
		if strings.HasPrefix(obj.Rule.Comment, fmt.Sprintf("%s %d", ForwardingComment, srcPort)) {
			_, err := m.execNFT("delete", "rule", "ip", "nat", "output", "handle", strconv.FormatInt(obj.Rule.Handle, 10))
			return err
		}
	}

	return errors.New("output DNAT rule not found")
}

func (m *ForwardingManager) loadDisabledRules() ([]ForwardingRule, error) {
	data, err := os.ReadFile(m.disabledForwardsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []ForwardingRule{}, nil
		}
		return nil, err
	}

	var file DisabledForwardsFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}

	// Mark all as disabled
	for i := range file.Rules {
		file.Rules[i].Enabled = false
		file.Rules[i].ID = fmt.Sprintf("fwd_%d", file.Rules[i].SrcPort)
	}

	return file.Rules, nil
}

func (m *ForwardingManager) saveDisabledRules(rules []ForwardingRule) error {
	// Ensure directory exists
	dir := filepath.Dir(m.disabledForwardsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file := DisabledForwardsFile{Rules: rules}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.disabledForwardsPath, data, 0644)
}

// isValidIPv4 validates an IPv4 address
func isValidIPv4(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	// Ensure it's IPv4 (not IPv6)
	return parsed.To4() != nil
}

// sanitizeForwardingComment removes invalid characters from comment
func sanitizeForwardingComment(s string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9\s\-_.\u4e00-\u9fff]`)
	s = re.ReplaceAllString(s, "")
	if len(s) > 100 {
		s = s[:100]
	}
	return s
}

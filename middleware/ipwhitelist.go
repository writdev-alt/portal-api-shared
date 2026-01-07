package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// IPWhitelistConfig configuration for IP whitelist middleware
type IPWhitelistConfig struct {
	AllowedIPs     []string
	AllowedCIDRs   []string
	CloudflareOnly bool
	TrustCloudflare bool
}

// IPWhitelist middleware untuk membatasi akses berdasarkan IP
func IPWhitelist(config IPWhitelistConfig) gin.HandlerFunc {
	// Parse CIDR ranges
	var allowedNetworks []*net.IPNet
	for _, cidr := range config.AllowedCIDRs {
		if cidr != "" {
			_, network, err := net.ParseCIDR(cidr)
			if err == nil {
				allowedNetworks = append(allowedNetworks, network)
			}
		}
	}

	// Cloudflare IP ranges
	cloudflareIPs := getCloudflareIPRanges()

	return func(c *gin.Context) {
		// Get real IP (prioritize Cloudflare header)
		realIP := getRealIP(c)

		// If Cloudflare only mode, check if IP is from Cloudflare
		if config.CloudflareOnly {
			if !isCloudflareIP(realIP, cloudflareIPs) {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Access denied: Request must come from Cloudflare",
				})
				c.Abort()
				return
			}
		}

		// Check if IP is in whitelist
		if !isIPAllowed(realIP, config.AllowedIPs, allowedNetworks) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied: IP not whitelisted",
			})
			c.Abort()
			return
		}

		// Set real IP in context for later use
		c.Set("real_ip", realIP)
		c.Next()
	}
}

// getRealIP extracts the real client IP from request
func getRealIP(c *gin.Context) string {
	// Cloudflare header (most reliable)
	if cfIP := c.GetHeader("CF-Connecting-IP"); cfIP != "" {
		return cfIP
	}

	// X-Forwarded-For header
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	return c.ClientIP()
}

// isIPAllowed checks if IP is in the whitelist
func isIPAllowed(ip string, allowedIPs []string, allowedNetworks []*net.IPNet) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// Check exact IP match
	for _, allowedIP := range allowedIPs {
		if allowedIP != "" && parsedIP.Equal(net.ParseIP(allowedIP)) {
			return true
		}
	}

	// Check CIDR match
	for _, network := range allowedNetworks {
		if network.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// isCloudflareIP checks if IP is from Cloudflare
func isCloudflareIP(ip string, cloudflareRanges []*net.IPNet) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, network := range cloudflareRanges {
		if network.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// getCloudflareIPRangesInternal returns Cloudflare IP ranges (internal, used by cloudflare.go)
func getCloudflareIPRangesInternal() []*net.IPNet {
	return getCloudflareIPRanges()
}

// getCloudflareIPRanges returns Cloudflare IP ranges (internal helper)
func getCloudflareIPRanges() []*net.IPNet {
	var ranges []*net.IPNet

	// Cloudflare IPv4 ranges
	ipv4Ranges := []string{
		"173.245.48.0/20",
		"103.21.244.0/22",
		"103.22.200.0/22",
		"103.31.4.0/22",
		"141.101.64.0/18",
		"108.162.192.0/18",
		"190.93.240.0/20",
		"188.114.96.0/20",
		"197.234.240.0/22",
		"198.41.128.0/17",
		"162.158.0.0/15",
		"104.16.0.0/13",
		"104.24.0.0/14",
		"172.64.0.0/13",
		"131.0.72.0/22",
	}

	// Cloudflare IPv6 ranges
	ipv6Ranges := []string{
		"2400:cb00::/32",
		"2606:4700::/32",
		"2803:f800::/32",
		"2405:b500::/32",
		"2405:8100::/32",
		"2a06:98c0::/29",
		"2c0f:f248::/32",
	}

	for _, cidr := range append(ipv4Ranges, ipv6Ranges...) {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil {
			ranges = append(ranges, network)
		}
	}

	return ranges
}

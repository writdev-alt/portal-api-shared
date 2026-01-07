package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	cloudflareRanges []*net.IPNet
	cloudflareOnce   sync.Once
)

// CloudflareIPWhitelist middleware untuk hanya allow request dari Cloudflare
func CloudflareIPWhitelist() gin.HandlerFunc {
	cloudflareOnce.Do(func() {
		cloudflareRanges = loadCloudflareIPRanges()
	})

	return func(c *gin.Context) {
		realIP := getRealIPFromContext(c)

		if !isCloudflareIPFromRanges(realIP, cloudflareRanges) {
			if cfIP := c.GetHeader("CF-Connecting-IP"); cfIP == "" {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Access denied: Request must come from Cloudflare",
				})
				c.Abort()
				return
			}
		}

		c.Set("real_ip", realIP)
		c.Set("cloudflare_ip", c.GetHeader("CF-Connecting-IP"))
		c.Next()
	}
}

func loadCloudflareIPRanges() []*net.IPNet {
	var ranges []*net.IPNet

	if filePath := os.Getenv("CLOUDFLARE_IPS_FILE"); filePath != "" {
		if data, err := os.ReadFile(filePath); err == nil {
			var config struct {
				IPv4 []string `json:"ipv4"`
				IPv6 []string `json:"ipv6"`
			}
			if json.Unmarshal(data, &config) == nil {
				for _, cidr := range append(config.IPv4, config.IPv6...) {
					_, network, err := net.ParseCIDR(cidr)
					if err == nil {
						ranges = append(ranges, network)
					}
				}
				return ranges
			}
		}
	}

	// Use function from ipwhitelist.go
	return getCloudflareIPRangesInternal()
}

func VerifyCloudflareRequest(c *gin.Context) bool {
	if cfIP := c.GetHeader("CF-Connecting-IP"); cfIP != "" {
		return true
	}
	realIP := getRealIPFromContext(c)
	return isCloudflareIPFromRanges(realIP, cloudflareRanges)
}

func GetCloudflareCountry(c *gin.Context) string {
	return c.GetHeader("CF-IPCountry")
}

func GetCloudflareRay(c *gin.Context) string {
	return c.GetHeader("CF-Ray")
}

// getRealIPFromContext extracts the real client IP from request (internal helper)
func getRealIPFromContext(c *gin.Context) string {
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

// isCloudflareIPFromRanges checks if IP is from Cloudflare (internal helper)
func isCloudflareIPFromRanges(ip string, cloudflareRanges []*net.IPNet) bool {
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

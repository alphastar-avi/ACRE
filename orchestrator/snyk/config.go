package snyk

// Config defines the configuration settings for Snyk vulnerability remediation.
type Config struct {
	TargetSeverities map[string]bool `json:"target_severities"` // Severities to remediate, e.g. {"HIGH": true, "MEDIUM": true, "CRITICAL": true}
}

// DefaultConfig returns the default configuration targeting HIGH, MEDIUM, and CRITICAL severity vulnerabilities.
func DefaultConfig() Config {
	return Config{
		TargetSeverities: map[string]bool{
			"HIGH":     true,
			"MEDIUM":   true,
			"CRITICAL": true,
		},
	}
}

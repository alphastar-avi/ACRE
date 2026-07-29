package snyk

// Config defines the configuration for Snyk vulnerability remediation.
type Config struct {
	TargetSeverities  map[string]bool `json:"target_severities"`  // Severities to remediate, e.g. {"HIGH": true}
	MaxSnykFixRetries int             `json:"max_snyk_fix_retries"` // Max attempts to fix & verify snyk (default 3)
	MaxBuildRetries   int             `json:"max_build_retries"`    // Max attempts to fix compilation build (default 3)
}

// DefaultConfig returns the default configuration targeting HIGH severity vulnerabilities.
func DefaultConfig() Config {
	return Config{
		TargetSeverities: map[string]bool{
			"HIGH":     true,
			"CRITICAL": true,
		},
		MaxSnykFixRetries: 3,
		MaxBuildRetries:   3,
	}
}

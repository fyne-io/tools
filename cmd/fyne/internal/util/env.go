package util

import "strings"

const goflagsEnvKey = "GOFLAGS"

// ExtractLdflagsFromGoFlags returns the ldflags and environment with ldflags removed from GOFLAGS.
func ExtractLdflagsFromGoFlags(env []string) (string, []string) {
	prefix := goflagsEnvKey + "="
	for i, v := range env {
		if strings.HasPrefix(v, prefix) {
			ldflags, goflags := ExtractLdFlags(strings.TrimPrefix(v, prefix))
			env[i] = prefix + goflags
			return ldflags, env
		}
	}
	return "", env
}

// ExtractLdFlags extracts ldflags from the value of GOFLAGS environment variable, returns ldflags and new GOFLAGS.
func ExtractLdFlags(goFlags string) (string, string) {
	if goFlags == "" {
		return "", ""
	}
	var ldflags, newGoFlags []string
	for _, flag := range strings.Fields(goFlags) {
		if strings.HasPrefix(flag, "-ldflags=") {
			ldflags = append(ldflags, strings.TrimPrefix(flag, "-ldflags="))
		} else {
			newGoFlags = append(newGoFlags, flag)
		}
	}
	return JoinSpace(ldflags), JoinSpace(newGoFlags)
}

package register

import (
	"regexp"
	"strings"
)

type envvar map[string]string

var envKeyPattern = regexp.MustCompile(`^SERVICE_[a-zA-Z0-9]{3,}$`)
var envValuePattern = regexp.MustCompile(`^[a-zA-Z0-9]{3,}$`)

func valid(str string, pattern *regexp.Regexp) bool {
	if pattern.MatchString(str) {
		return true
	}
	return false
}

func keyPairs(vars []string) envvar {
	env := make(envvar)
	for _, v := range vars {
		if parts := strings.Split(v, "="); len(parts) > 0 {
			key, value := parts[0], parts[1]
			env[key] = value
		}
	}
	return env
}

func serviceVars(vars []string) envvar {
	serviceVars := make(envvar)
	for key, value := range keyPairs(vars) {
		if valid(key, envKeyPattern) && valid(value, envValuePattern) {
			serviceVars[key] = value
		}
	}
	return serviceVars
}

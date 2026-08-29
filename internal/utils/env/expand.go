package env

import "os"

// ExpandAuthored expands $VAR and ${VAR} references in a string written in a flow file,
// resolving them against envMap. "$$" yields a literal "$". A variable that does not
// resolve is left as written, so a missing value shows up in the output instead of
// silently disappearing; note that the brace form is not preserved, so an unresolved
// "${FOO}" comes back as "$FOO".
//
// Values supplied by a user - command line arguments, prompt responses - are literals
// and must not be passed through this.
func ExpandAuthored(value string, envMap map[string]string) string {
	if value == "" {
		return value
	}
	return os.Expand(value, func(key string) string {
		if key == "$" {
			return "$"
		}
		if v, ok := envMap[key]; ok {
			return v
		}
		return "$" + key
	})
}

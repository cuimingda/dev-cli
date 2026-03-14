package configtemplate

import _ "embed"

//go:embed default.yaml
var defaultYAML string

func DefaultYAML() string {
	return defaultYAML
}

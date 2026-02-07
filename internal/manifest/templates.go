package manifest

import "embed"

//go:embed templates/*.yaml.tmpl
var templateFS embed.FS

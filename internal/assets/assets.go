package assets

import _ "embed"

//go:embed ui/index.html
var IndexHTML string

//go:embed ui/.gitkeep
var KeepEmpty []byte

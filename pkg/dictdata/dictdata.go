// Package dictdata provides embedded dictionary metadata.
package dictdata

import (
	_ "embed"
)

//go:embed dictionaries.json
var JSON []byte

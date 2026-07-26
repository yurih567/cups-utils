package fonts

import _ "embed"

// Builtin sans-serif (Inter) registered as family "Arial" for a clean receipt look.
const Family = "Arial"

//go:embed Inter-Regular.ttf
var Regular []byte

//go:embed Inter-Bold.ttf
var Bold []byte

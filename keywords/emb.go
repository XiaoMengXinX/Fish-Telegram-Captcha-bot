package keywords

import _ "embed"

//go:embed blacklist.txt
var Blacklist []byte

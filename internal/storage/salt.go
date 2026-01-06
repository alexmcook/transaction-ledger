package storage

import (
	"time"
)

// LOAD TESTING salt to avoid ID collisions
var salt = uint32(time.Now().UnixNano())

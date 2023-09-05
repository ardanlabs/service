// Package keys stores development public/private key pairs used by the service.
package keys

import "embed"

// DevKeysFS is a filesystem that holds development private/public key pairs
// used by the service.
//
//go:embed *
var DevKeysFS embed.FS

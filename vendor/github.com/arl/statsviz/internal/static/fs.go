package static

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"fmt"
	"io/fs"
	"sync"
)

//go:embed dist.zip
var distzip []byte

var Assets = sync.OnceValue(func() fs.FS {
	zr, err := zip.NewReader(bytes.NewReader(distzip), int64(len(distzip)))
	if err != nil {
		panic(fmt.Sprintf("error reading dist.zip: %s", err))
	}
	webFS, err := fs.Sub(zr, "dist")
	if err != nil {
		panic(fmt.Sprintf("error loading frontend assets: %s", err))
	}
	return webFS
})

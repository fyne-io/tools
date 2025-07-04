package metadata

import (
	"bufio"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Save attempts to write a FyneApp metadata to the provided writer.
// If the encoding fails an error will be returned.
func Save(f *FyneApp, w io.Writer) error {
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	enc := toml.NewEncoder(bw)
	enc.Indent = ""
	return enc.Encode(f)
}

// SaveStandard attempts to save a FyneApp metadata to the `FyneApp.toml` file in the specified dir.
// If the file cannot be written or encoding fails an error will be returned.
func SaveStandard(f *FyneApp, dir string) error {
	path := filepath.Join(dir, "FyneApp.toml")
	w, err := os.Create(path)
	if err != nil {
		return err
	}

	defer w.Close()
	return Save(f, w)
}

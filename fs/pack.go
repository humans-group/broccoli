package fs

import (
	"bytes"
	"encoding/gob"
	"sort"

	"github.com/andybalholm/brotli"
)

// Pack compresses a set of files from disk for bundled use in the generated code.
//
// This function is only supposed to be called by broccoli the tool.
func Pack(files []*File, quality int) ([]byte, error) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Fpath < files[j].Fpath
	})

	var b bytes.Buffer
	w := brotli.NewWriterLevel(&b, quality)
	if err := gob.NewEncoder(w).Encode(files); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// New decompresses the bundle byte-slice and creates a virtual file system.
// Depending on whether if optional decompression is enabled, it will or
// will not decompress the files while loading them.
//
// This function is only supposed to be called from the generated code.
func New(opt bool, bundle []byte) *Broccoli {
	var files []*File
	r := brotli.NewReader(bytes.NewBuffer(bundle))
	if err := gob.NewDecoder(r).Decode(&files); err != nil {
		panic(err)
	}

	br := &Broccoli{
		filePaths: make([]string, 0, len(files)),
		files:     map[string]*File{},
	}

	for _, f := range files {
		// files.Data already contains raw uncompressed file data (bundle decompresses the gob),
		// so mark as not compressed and attach the broccoli reference.
		f.compressed = false
		f.br = br

		br.files[f.Fpath] = f
		br.filePaths = append(br.filePaths, f.Fpath)
	}

	return br
}

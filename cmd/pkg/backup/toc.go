package backup

import (
	"context"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/mholt/archiver/v4"
)

func toc() {
	// the type that will be used to read the input stream
	format := archiver.CompressedArchive{
		Compression: archiver.Gz{},
		Archival:    archiver.Tar{},
	}

	// the list of files we want out of the archive; any
	// directories will include all their contents unless
	// we return fs.SkipDir from our handler
	// (leave this nil to walk ALL files from the archive)
	fileList := []string{"tmp/foo/bar/a/"}

	handler := func(ctx context.Context, f archiver.File) error {
		// do something with the file
		spew.Dump(f.NameInArchive)
		return nil
	}

	input, _ := os.Open("foo.tgz")

	err := format.Extract(context.Background(), input, fileList, handler)
	if err != nil {
		panic(err)
	}
}

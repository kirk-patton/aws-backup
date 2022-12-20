package backup

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/h2non/filetype"
	"github.com/mholt/archiver/v4"
)

type FileMatcher struct {
	// Any matches any file type
	Any   bool
	Video bool
}

// Parse - return a list of files matching the configured FileMatcher
func (m FileMatcher) Parse(files []archiver.File) ([]archiver.File, error) {
	// We should pass this in...
	var matched []archiver.File

	for _, f := range files {
		if f.FileInfo.IsDir() {
			matched = append(matched, f)
			continue
		}

		// determine the file type
		fh, err := os.Open(f.NameInArchive)
		if err != nil {
			return nil, fmt.Errorf("file: %s, %w", f.NameInArchive, err)
		}
		defer fh.Close()

		// We only have to pass the file header = first 261 bytes
		head := make([]byte, 261)
		_, err = fh.Read(head)
		if err != nil {
			return nil, fmt.Errorf("file: %s, %w", f.NameInArchive, err)
		}

		if m.Any {
			matched = append(matched, f)
			continue
		}

		if m.Video && filetype.IsVideo(head) {
			matched = append(matched, f)
			continue
		}

		// TODO we need a logger...
		fmt.Printf("failed match: %s\n", f.NameInArchive)
		continue

	}
	return matched, nil
}

type Tarchive struct {
	wg                *sync.WaitGroup
	pipeR             *io.PipeReader
	pipeW             *io.PipeWriter
	compressedArchive archiver.CompressedArchive
	chunkSize         int
}

func (a *Tarchive) NewTarchive(dir string) chan error {
	errCH := make(chan error)
	// set the size of the upload buffer
	if a.chunkSize == 0 {
		a.chunkSize = 1024 * 1024 * 5
	}

	go func() {
		a.pipeR, a.pipeW = io.Pipe()
		defer a.pipeW.Close()
		defer close(errCH)

		files, err := archiver.FilesFromDisk(&archiver.FromDiskOptions{}, map[string]string{
			dir: dir,
		})
		if err != nil {
			fmt.Println("files error")
			errCH <- err
			return
		}

		// filter the files and match on file type
		m := FileMatcher{Video: true}
		files, err = m.Parse(files)
		if err != nil {
			panic(err)
		}

		format := archiver.CompressedArchive{
			//Compression: archiver.Gz{},
			Archival: archiver.Tar{},
		}

		err = format.Archive(context.Background(), a.pipeW, files)
		if err != nil {
			errCH <- err
			return
		}
	}()
	return errCH
}

// ReadChunk - read data into the upload buffer
func (a *Tarchive) ReadChunk() (*bytes.Buffer, [16]byte, error) {
	b := make([]byte, a.chunkSize)
	_, err := io.ReadFull(a.pipeR, b)

	buf := bytes.NewBuffer(b)

	// TODO we should encrypt here...

	// calculate the md5sum of the chunk
	hash := md5.Sum(buf.Bytes())
	return buf, hash, err
}

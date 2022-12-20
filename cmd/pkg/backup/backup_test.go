package backup

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateVault(t *testing.T) {
	err := newGlacierService()
	assert.NoError(t, err)
}

// TestTar - create a tar.gz file
// read it in chunks and reassemble it.
// check no-error...
func TestTar(t *testing.T) {
	// set a small chunk size
	ar := &Tarchive{chunkSize: 100}

	errCH := ar.NewTarchive("/tmp/foo")

	select {
	case err := <-errCH:
		panic(err)
	default:
		break
	}

	fh, err := os.Create("/tmp/q.tar")
	assert.NoError(t, err)
	defer fh.Close()

	// _, err = io.Copy(fh, ar.pipeR)
	// assert.NoError(t, err)

	for {
		b, _, err := ar.ReadChunk()
		//fmt.Printf("md5: %x\n", sum)
		if err == io.EOF {
			break
		}
		fh.Write(b.Bytes())
	}

	fmt.Println("done")
}

// TODO
// upload a multipart archive
// then manually download and untar
// Fudge, I tar.gz'ed instead of tar-> encrypt...
// we can encrypt each chunk...

func TestTOC(t *testing.T) {
	toc()
}

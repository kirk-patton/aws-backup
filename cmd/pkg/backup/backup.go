package backup

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/glacier"
	"github.com/mholt/archiver/v4"
)

var (
	cfg       aws.Config
	svc       *glacier.Client
	err       error
	accountID = "529467942398"
	vaultName = "test-vault"
)

func newInit() error {
	cfg, err = config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-west-2"),
	)
	if err != nil {
		return fmt.Errorf("unabled to initialize aws config: %v", err)
	}
	return nil
}

func newGlacierService() error {
	err = newInit()
	if err != nil {
		return err
	}

	svc := glacier.NewFromConfig(cfg)
	ctx := context.Background()
	_, err := svc.CreateVault(ctx, &glacier.CreateVaultInput{AccountId: &accountID, VaultName: &vaultName})
	if err != nil {
		return err
	}
	vaultDesc, err := svc.DescribeVault(ctx, &glacier.DescribeVaultInput{AccountId: &accountID, VaultName: &vaultName})
	if err != nil {
		return err
	}

	fmt.Printf("%s", *vaultDesc.VaultARN)

	return nil
}

// pipe used by channel
type pipe struct {
	data *io.PipeReader
	err  error
}

// TODO
// tar a directory with the output directed to an io.Pipe
// return the pipe
// another process should read from the pipe and split up the file into chunks
// for upload to aws as a multi-part upload
func tarDirectory(dir string, ch chan pipe) {
	go func() {
		r, w := io.Pipe()
		defer w.Close()
		defer close(ch)

		// set up the pipe
		d := pipe{
			data: r,
		}

		files, err := archiver.FilesFromDisk(&archiver.FromDiskOptions{}, map[string]string{
			dir: dir,
		})
		if err != nil {
			fmt.Println("files error")
			d.err = err
			ch <- d
		}

		format := archiver.CompressedArchive{
			Compression: archiver.Gz{},
			Archival:    archiver.Tar{},
		}

		fmt.Println("about to create archive")
		fmt.Fprintf(w, "Write to pipe\n")
		err = format.Archive(context.Background(), w, files)
		if err != nil {
			fmt.Println("files error")
			d.err = err
			ch <- d
		}
		fmt.Println("finished read of archive")
	}()
}

func chFoo(d string, ch chan string) {
	go func() {
		fmt.Println("placing data on the channel")
		ch <- d
		fmt.Println("wrote data on the channel")
	}()
	return
}

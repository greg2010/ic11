package filereader

import (
	"errors"
	"io"
	"os"
)

// MultiFileReader opens multiple files on the filesystem
type MultiFileReader struct {
	readers    []io.Reader
	closeFuncs []func() error
}

func (mfr *MultiFileReader) GetReaders() []io.Reader {
	return mfr.readers
}

// See os.File.Close()
func (mfr *MultiFileReader) Close() error {
	var errs []error
	for _, closeFunc := range mfr.closeFuncs {
		err := closeFunc()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// Creates new instance of MultiFileReader. Returns error if any of the files could not be open.
func New(fnames ...string) (*MultiFileReader, error) {
	var readers []io.Reader
	var closeFuncs []func() error

	for _, fname := range fnames {
		file, err := os.Open(fname)
		if err != nil {
			return nil, err
		}

		readers = append(readers, file)
		closeFuncs = append(closeFuncs, file.Close)
	}

	return &MultiFileReader{
		readers:    readers,
		closeFuncs: closeFuncs,
	}, nil
}

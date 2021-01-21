package gtransfer

import (
	"io/ioutil"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
)

func TestTransferSuite(t *testing.T) {
	suite.Run(t, new(TransferSuite))
}

type TransferSuite struct {
	suite.Suite

	fs afero.Fs
}

func (suite *TransferSuite) SetupTest() {
	suite.fs = afero.NewMemMapFs()
}

func (suite *TransferSuite) TestTransferSimple() {
	// setup source
	from := afero.NewMemMapFs()
	foo, err := from.Create("foo.txt")
	suite.NoError(err)
	_, err = foo.WriteString("Hello, World!")
	suite.NoError(err)
	suite.NoError(foo.Close())

	to := afero.NewMemMapFs()
	srv := NewServer(":0", from)
	defer srv.Stop()
	go func() {
		suite.NoError(srv.Serve())
	}()
	<-srv.Listening()
	client := Dial(srv.Addr())
	suite.NoError(client.DownloadInto(to))

	foo, err = to.Open("foo.txt")
	suite.NoError(err)
	data, err := ioutil.ReadAll(foo)
	suite.NoError(err)
	suite.Equal("Hello, World!", string(data))
	suite.NoError(foo.Close())
}

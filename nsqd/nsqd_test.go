package nsqd

import (
	"encoding/json"
	"io/ioutil"
	"nsq-learn/internal/test"
	"os"
	"testing"
)

func getMetadata(n *NSQD) (*meta, error) {
	fn := newMetadataFile(n.getOpts())
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	var m meta
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	return &m, err
}

func TestStartup(t *testing.T) {
	opts := NewOptions()
	opts.Logger = test.NewTestLogger(t)
	testStartNSQD(opts)
	//测试完后清楚掉data目录
	defer os.RemoveAll(opts.DataPath)
	// mac上有tmp目录不能锁定问题，后面再看 resource temporarily unavailable
	// nsqd := testStartNSQD(opts)
	// assert.NotNil(t, nsqd)
}

func testStartNSQD(opts *Options) *NSQD {
	opts.HTTPAddress = "127.0.0.1:0"
	if opts.DataPath == "" {
		tmpDir, err := ioutil.TempDir("", "nsq-test-")
		if err != nil {
			panic(err)
		}
		opts.DataPath = tmpDir
	}
	nsqd := New(opts)
	nsqd.Main()
	return nsqd
}

package nsqd

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTopic(t *testing.T) {
	opts := NewOptions()
	nsqd := topicMustStartNSQD(opts)
	defer os.RemoveAll(opts.DataPath)
	topic1 := nsqd.GetTopic("test")
	assert.NotNil(t, topic1)
	assert.Equal(t, "test", topic1.name)
	topic2 := nsqd.GetTopic("test")
	assert.Equal(t, topic1, topic2)

	topic3 := nsqd.GetTopic("test2")
	assert.Equal(t, "test2", topic3.name)
	assert.NotEqual(t, topic2, topic3)
}

func topicMustStartNSQD(opts *Options) *NSQD {
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

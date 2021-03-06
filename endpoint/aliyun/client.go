package aliyun

import (
	"context"
	"net/http"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/yunify/qscamel/constants"
	"github.com/yunify/qscamel/model"
)

// Client is the client to visit aliyun oss service.
type Client struct {
	Endpoint        string `yaml:"endpoint"`
	BucketName      string `yaml:"bucket_name"`
	AccessKeyID     string `yaml:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret"`

	Path string

	client *oss.Bucket
}

// New will create a client.
func New(ctx context.Context, et uint8, _ *http.Client) (c *Client, err error) {
	t, err := model.GetTask(ctx)
	if err != nil {
		return
	}

	c = &Client{}

	e := t.Src
	if et == constants.DestinationEndpoint {
		e = t.Dst
	}

	content, err := yaml.Marshal(e.Options)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(content, c)
	if err != nil {
		return
	}

	// Set endpoint
	if c.Endpoint == "" {
		logrus.Error("Aliyun OSS's endpoint can't be empty.")
		err = constants.ErrEndpointInvalid
		return
	}

	// Set bucket name.
	if c.BucketName == "" {
		logrus.Error("Aliyun OSS's bucket name can't be empty.")
		err = constants.ErrEndpointInvalid
		return
	}

	// Set access key.
	if c.AccessKeyID == "" {
		logrus.Error("Aliyun OSS's access key id can't be empty.")
		err = constants.ErrEndpointInvalid
		return
	}

	// Set secret key.
	if c.AccessKeySecret == "" {
		logrus.Error("Aliyun OSS's access key secret can't be empty.")
		err = constants.ErrEndpointInvalid
		return
	}

	// Set prefix.
	c.Path = e.Path

	service, err := oss.New(c.Endpoint, c.AccessKeyID, c.AccessKeySecret)
	if err != nil {
		return
	}
	c.client, err = service.Bucket(c.BucketName)
	if err != nil {
		return
	}

	return
}

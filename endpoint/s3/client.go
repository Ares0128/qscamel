package s3

import (
	"context"
	"github.com/aws/aws-sdk-go/aws/corehandlers"

	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/yunify/qscamel/constants"
	"github.com/yunify/qscamel/endpoint/s3/signer/v2"
	"github.com/yunify/qscamel/model"
)

// Client is the client to visit service.
type Client struct {
	BucketName          string `yaml:"bucket_name"`
	Endpoint            string `yaml:"endpoint"`
	Region              string `yaml:"region"`
	AccessKeyID         string `yaml:"access_key_id"`
	SecretAccessKey     string `yaml:"secret_access_key"`
	DisableSSL          bool   `yaml:"disable_ssl"`
	UseAccelerate       bool   `yaml:"use_accelerate"`
	PathStyle           bool   `yaml:"path_style"`
	EnableListObjectsV2 bool   `yaml:"enable_list_objects_v2"`
	EnableSignatureV2   bool   `yaml:"enable_signature_v2"`
	DisableURICleaning  bool   `yaml:"disable_uri_cleaning"`

	Path string

	client *s3.S3
}

// New will create a new client.
func New(ctx context.Context, et uint8, hc *http.Client) (c *Client, err error) {
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

	// Set bucket name.
	if c.BucketName == "" {
		logrus.Error("AWS bucket name can't be empty.")
		err = constants.ErrEndpointInvalid
		return
	}

	// Set access key.
	if c.AccessKeyID == "" {
		logrus.Error("AWS access key id can't be empty.")
		err = constants.ErrEndpointInvalid
		return
	}

	// Set secret key.
	if c.SecretAccessKey == "" {
		logrus.Error("AWS's secret access key can't be empty.")
		err = constants.ErrEndpointInvalid
		return
	}

	// Set path.
	c.Path = e.Path

	cfg := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(c.AccessKeyID, c.SecretAccessKey, ""),
		Endpoint:         aws.String(c.Endpoint),
		Region:           aws.String(c.Region),
		DisableSSL:       aws.Bool(c.DisableSSL),
		S3UseAccelerate:  aws.Bool(c.UseAccelerate),
		S3ForcePathStyle: aws.Bool(c.PathStyle),
		HTTPClient:       hc,
		// AWS S3 SDK will clean the url automatically which will convert "abc//bcd" -> "abc/bcd"
		DisableRestProtocolURICleaning: aws.Bool(c.DisableURICleaning),
	}
	sess, err := session.NewSession(cfg)
	if err != nil {
		return
	}
	c.client = s3.New(sess)
	if c.EnableSignatureV2 {
		// s3.New will push v4.SignRequestHandler into the sign handlers.
		// In order to force client use v2.SignRequestHandler, we should remove it
		// and add the v2.SignRequestHandler after client initiation.
		c.client.Handlers.Sign.Clear()
		c.client.Handlers.Sign.PushBack(v2.Sign)
		c.client.Handlers.Sign.PushBackNamed(corehandlers.BuildContentLengthHandler)
	}
	return
}

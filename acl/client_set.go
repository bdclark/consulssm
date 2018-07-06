package acl

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ClientSet represents a collection of clients
type ClientSet struct {
	SSM       *ssm.SSM
	Consul    *consulapi.Client
	kmsKeyID  string
	overwrite bool
	insecure  bool
}

// ClientSetInput is used as input for the NewClientSet function
type ClientSetInput struct {
	ConsulTokenParam string
	KMSKeyID         string
	Overwrite        bool
	Insecure         bool
}

// NewClientSet creates a new client collection
func NewClientSet(i *ClientSetInput) (*ClientSet, error) {
	sess := session.Must(session.NewSession())
	ssmClient := ssm.New(sess)

	var c ClientSet
	c.SSM = ssmClient
	c.kmsKeyID = i.KMSKeyID
	c.overwrite = i.Overwrite
	c.insecure = i.Insecure

	consulConfig := consulapi.DefaultConfig()
	if i.ConsulTokenParam != "" {
		val, err := c.GetStringParameter(i.ConsulTokenParam, true)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to get management token from SSM parameter \"%s\"", i.ConsulTokenParam)
		}
		log.Debugf("Using Consul token from SSM parameter \"%s\"", i.ConsulTokenParam)
		consulConfig.Token = *val
	}

	consulClient, err := consulapi.NewClient(consulConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create Consul client")
	}
	c.Consul = consulClient

	return &c, nil
}

// GetStringParameter reads a SSM parameter and returns a string
func (c *ClientSet) GetStringParameter(name string, failNotFound bool) (*string, error) {
	resp, err := c.SSM.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if !failNotFound && aerr.Code() == ssm.ErrCodeParameterNotFound {
				return aws.String(""), nil
			}
		}
		return nil, err
	}

	return resp.Parameter.Value, nil
}

// PutStringParameter writes a SSM parameter as a string
func (c *ClientSet) PutStringParameter(name, value string) (err error) {
	i := ssm.PutParameterInput{
		Name:      aws.String(name),
		Value:     aws.String(value),
		Overwrite: aws.Bool(c.overwrite),
	}
	if c.insecure {
		i.Type = aws.String("String")
	} else {
		i.Type = aws.String("SecureString")
	}
	if c.kmsKeyID != "" {
		i.KeyId = aws.String(c.kmsKeyID)
	}

	log.Debugf("Setting %s parameter: %s", *i.Type, name)
	_, err = c.SSM.PutParameter(&i)
	return
}

// isLeader determines if current agent is the Consul leader
func (c *ClientSet) isLeader() (bool, error) {
	resp, err := c.Consul.Agent().Self()
	if err != nil {
		return false, err
	}

	statsConsul, ok := resp["Stats"]["consul"].(map[string]interface{})
	if !ok {
		return false, errors.New("Failed to parse /agent/self stats")
	}

	leader, ok := statsConsul["leader"].(string)
	if !ok {
		return false, errors.New("Failed to parse /agent/self consul stats")
	} else if leader == "true" {
		return true, nil
	}
	return false, nil
}

package acl

import (
	"github.com/pkg/errors"
)

// Bootstrap performs a Consul ACL bootstrap and saves the resulting token to an SSM parameter
func (c *ClientSet) Bootstrap(consulTokenParam string) (string, error) {

	if consulTokenParam == "" {
		return "", errors.New("consulTokenParam cannot be empty")
	}

	id, _, err := c.Consul.ACL().Bootstrap()
	if err != nil {
		return "", errors.Wrap(err, "Bootstrap failed")
	}

	if err := c.PutStringParameter(consulTokenParam, id); err != nil {
		return id, errors.Wrapf(err, "Bootstrap succeeded, but failed to save token SSM parameter to \"%s\"", consulTokenParam)
	}

	return id, nil
}

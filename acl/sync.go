package acl

import (
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// SyncInput is the input for the Sync function
type SyncInput struct {
	ACLDefinitionPrefix string
	ACLIDPrefix         string
	PageSize            int64
	OnlyIfConsulLeader  bool
}

// aclItem is the internal representation of an ACL
type aclItem struct {
	consulapi.ACLEntry
	Destroy bool `json:",string"`
	slug    string
}

// Sync syncronizes ACLS with AWS SSM
func (c *ClientSet) Sync(i *SyncInput) error {
	aclDefinitionPrefix := ensureTrailingSlash(i.ACLDefinitionPrefix)
	aclIDPrefix := ensureTrailingSlash(i.ACLIDPrefix)
	if aclDefinitionPrefix == "" {
		return errors.New("ACLDefinitionPrefix is required")
	}

	if i.OnlyIfConsulLeader {
		isLeader, err := c.isLeader()
		if err != nil {
			return errors.Wrap(err, "Failed to determine Consul leader")
		}
		if !isLeader {
			log.Info("Not currently the leader, nothing to do.")
			return nil
		}
	}

	pageNum := 0
	params := &ssm.GetParametersByPathInput{
		Path:           aws.String(aclDefinitionPrefix),
		Recursive:      aws.Bool(true),
		WithDecryption: aws.Bool(true),
	}
	if i.PageSize > 0 {
		params.MaxResults = aws.Int64(i.PageSize)
	}

	fn := func(output *ssm.GetParametersByPathOutput, lastPage bool) bool {
		pageNum++
		log.Debugf("GetParametersByPathPages page: %d, lastPage?: %t", pageNum, lastPage)
		for _, item := range output.Parameters {
			acl := c.parameterToACL(item, aclDefinitionPrefix, aclIDPrefix)
			c.manageACL(acl, aclDefinitionPrefix, aclIDPrefix)
		}
		return true
	}

	err := c.SSM.GetParametersByPathPages(params, fn)
	if err != nil {
		return errors.Wrapf(err, "Failed to get ACL definition parameters from prefix \"%s\"", aclDefinitionPrefix)
	}

	return nil
}

// parameterToACL is a helper for Sync and converts an SSM parameter to an aclItem
func (c *ClientSet) parameterToACL(param *ssm.Parameter, aclDefinitionPrefix, aclIDPrefix string) *aclItem {
	var acl aclItem
	parts := strings.Split(strings.TrimPrefix(*param.Name, aclDefinitionPrefix), "/")
	acl.slug = parts[len(parts)-1]
	log.Debugf("Got SSM parameter name: %s, value: %s, slug: %s", *param.Name, *param.Value, acl.slug)

	err := json.Unmarshal([]byte(*param.Value), &acl)
	if err != nil {
		log.Fatalf("Failed to parse parameter %s from %s as acl: %s", acl.slug, *param.Name, err.Error())
	}

	// if ID not provided, attempt to get it from <aclIDPrefix>/slug
	if acl.ID == "" {
		idParam := aclIDPrefix + acl.slug
		if val, err := c.GetStringParameter(idParam, false); err != nil {
			log.Fatalf("Failed to get ACL ID from SSM parameter \"%s\": %s", idParam, err.Error())
		} else {
			acl.ID = *val
		}
	}

	// Type should default to client
	if acl.Type == "" {
		acl.Type = "client"
	}

	return &acl
}

// manageACL is a helper for Sync and manages a particular ACL item
func (c *ClientSet) manageACL(acl *aclItem, aclDefinitionPrefix, aclIDPrefix string) {
	if acl.ID == "" {
		// we are working on an ACL without an ID provided

		if acl.Destroy {
			log.Warnf("Unable to destroy ACL %s (Name: \"%s\"), no ID was provided.", acl.slug, acl.Name)

		} else {
			log.Infof("Creating ACL %s (Name: \"%s\") because no ID was provided.", acl.slug, acl.Name)

			id, _, err := c.Consul.ACL().Create(&consulapi.ACLEntry{
				Name:  acl.Name,
				Type:  acl.Type,
				Rules: acl.Rules,
			}, nil)
			if err != nil {
				log.Fatalf("Failed to create ACL %s (Name: \"%s\"): %s", acl.slug, acl.Name, err.Error())
			}

			c.PutStringParameter(aclIDPrefix+acl.slug, id)

		}
	} else {
		// we are working on an ACL with an ID provided

		currentACL, _, err := c.Consul.ACL().Info(acl.ID, nil)
		if err != nil {
			log.Fatalf("Failed to get info for ACL %s (Name: \"%s\") ", acl.slug, acl.Name)
		}

		if currentACL == nil {
			// we are working on an ACL with an ID provided, but no existing ACL exists with that ID

			if acl.Destroy {
				log.Infof("Unable to destroy ACL %s (Name: \"%s\"), no ACL found with the provided ID.", acl.slug, acl.Name)

			} else {
				log.Infof("Creating ACL %s (Name: \"%s\") with provided ID.", acl.slug, acl.Name)

				_, _, err := c.Consul.ACL().Create(&consulapi.ACLEntry{
					ID:    acl.ID,
					Name:  acl.Name,
					Type:  acl.Type,
					Rules: acl.Rules,
				}, nil)
				if err != nil {
					log.Fatalf("Failed to create ACL %s (Name: \"%s\"): %s", acl.slug, acl.Name, err.Error())
				}
			}

		} else {
			// we are working on an ACL with an ID, and we have an existing ACL that matches that ID

			if acl.Destroy {
				log.Infof("Destroying ACL %s (Name: \"%s\").", acl.slug, acl.Name)

				if _, err := c.Consul.ACL().Destroy(acl.ID, nil); err != nil {
					log.Errorf("Failed to delete ACL %s (Name: \"%s\"): %s", acl.slug, acl.Name, err.Error())
				}

			} else if currentACL.Name == acl.Name && currentACL.Type == acl.Type && currentACL.Rules == acl.Rules {
				log.Infof("Skipping ACL %s (Name: \"%s\") - matches.", acl.slug, acl.Name)

			} else {
				log.Infof("Updating ACL %s (Name: \"%s\")", acl.slug, acl.Name)

				if _, err := c.Consul.ACL().Update(&consulapi.ACLEntry{
					ID:    acl.ID,
					Name:  acl.Name,
					Type:  acl.Type,
					Rules: acl.Rules,
				}, nil); err != nil {
					log.Errorf("Failed to update ACL %s (Name: \"%s\"): %s", acl.slug, acl.Name, err.Error())
				}

			}
		}
	}
}

// ensureTrailingSlash ensures the given string ends in "/"
func ensureTrailingSlash(s string) string {
	if len(s) == 0 || s[len(s)-1:] == "/" {
		return s
	}
	return s + "/"
}

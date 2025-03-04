package plugin

import (
	"fmt"

	ddbds "github.com/Vanssh-k/go-ds-dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/ipfs/kubo/plugin"
	"github.com/ipfs/kubo/repo"
	"github.com/ipfs/kubo/repo/fsrepo"
)

var Plugins = []plugin.Plugin{
	&DDBPlugin{},
}

type DDBPlugin struct{}

func (p *DDBPlugin) Name() string {
	return "ddb-datastore-plugin"
}

func (p *DDBPlugin) Version() string {
	return "0.0.1"
}

func (p *DDBPlugin) Init(env *plugin.Environment) error {
	return nil
}

func (p *DDBPlugin) DatastoreTypeName() string {
	return "ddbds"
}

func (p *DDBPlugin) DatastoreConfigParser() fsrepo.ConfigFromMap {
	return func(m map[string]interface{}) (fsrepo.DatastoreConfig, error) {
		table, ok := m["table"].(string)
		if !ok || table == "" {
			return nil, fmt.Errorf("ddbds: no table specified")
		}

		accessKey, _ := m["accessKey"].(string)
		secretKey, _ := m["secretKey"].(string)
		region, _ := m["region"].(string)

		return &DDBConfig{
			Table:     table,
			AccessKey: accessKey,
			SecretKey: secretKey,
			Region:    region,
		}, nil
	}
}

type DDBConfig struct {
	Table     string
	AccessKey string
	SecretKey string
	Region    string
}

func (c *DDBConfig) DiskSpec() fsrepo.DiskSpec {
	return fsrepo.DiskSpec{
		"table":     c.Table,
		"accessKey": c.AccessKey,
		"secretKey": c.SecretKey,
		"region": c.Region
	}
}

func (c *DDBConfig) Create(path string) (repo.Datastore, error) {
	awsConfig := &aws.Config{
		Region: aws.String(c.Region),
	}

	if c.AccessKey != "" && c.SecretKey != "" {
		awsConfig.Credentials = credentials.NewStaticCredentials(c.AccessKey, c.SecretKey)
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	ddbClient := dynamodb.New(sess)

	return ddbds.New(ddbClient, c.Table), nil
}

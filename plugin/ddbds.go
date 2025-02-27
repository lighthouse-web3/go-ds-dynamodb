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
		endpoint, ok := m["endpoint"].(string)
		if !ok || endpoint == "" {
			return nil, fmt.Errorf("ddbds: no endpoint specified")
		}

		table, ok := m["table"].(string)
		if !ok || table == "" {
			return nil, fmt.Errorf("ddbds: no table specified")
		}

		return &DDBConfig{
			Endpoint: endpoint,
			Table:    table,
		}, nil
	}
}

type DDBConfig struct {
	Endpoint string
	Table    string
}

func (c *DDBConfig) DiskSpec() fsrepo.DiskSpec {
	return fsrepo.DiskSpec{
		"endpoint": c.Endpoint,
		"table":    c.Table,
	}
}

func (c *DDBConfig) Create(path string) (repo.Datastore, error) {
	awsConfig := &aws.Config{
		Endpoint: aws.String(c.Endpoint),  // Connecting to local DynamoDB
		Region:   aws.String("us-east-1"), // Dummy region
		Credentials: credentials.NewStaticCredentials(
			"dummy", "dummy", "", // No real credentials needed for local
		),
		DisableSSL: aws.Bool(true),
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	ddbClient := dynamodb.New(sess)

	return ddbds.New(ddbClient, c.Table), nil
}

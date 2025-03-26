package plugin

import (
	"fmt"

	ddbds "github.com/Vanssh-k/go-ds-dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/ipfs/go-datastore"
	mount "github.com/ipfs/go-datastore/mount"
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
		accessKey, _ := m["accessKey"].(string)
		secretKey, _ := m["secretKey"].(string)
		region, _ := m["region"].(string)
		providersTable, _ := m["providersTable"].(string)
		pinsTable, _ := m["pinsTable"].(string)
		defaultTable, _ := m["defaultTable"].(string)

		return &DDBConfig{
			AccessKey:      accessKey,
			SecretKey:      secretKey,
			Region:        region,
			ProvidersTable: providersTable,
			PinsTable:      pinsTable,
			DefaultTable:   defaultTable,
		}, nil
	}
}

type DDBConfig struct {
	AccessKey      string
	SecretKey      string
	Region        string
	ProvidersTable string
	PinsTable      string
	DefaultTable   string
}

func (c *DDBConfig) DiskSpec() fsrepo.DiskSpec {
	return fsrepo.DiskSpec{
		"accessKey":      c.AccessKey,
		"secretKey":      c.SecretKey,
		"region":        c.Region,
		"providersTable": c.ProvidersTable,
		"pinsTable":      c.PinsTable,
		"defaultTable":   c.DefaultTable,
	}
}

func (c *DDBConfig) Create(path string) (repo.Datastore, error) {
	awsConfig := &aws.Config{
		Region: aws.String(c.Region),
	}

	if c.AccessKey != "" && c.SecretKey != "" {
		awsConfig.Credentials = credentials.NewStaticCredentials(c.AccessKey, c.SecretKey, "")
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	ddbClient := dynamodb.New(sess)

	// Mount different namespaces to dynamically configured tables
	ddbDS := mount.New([]mount.Mount{
		{
			Prefix: datastore.NewKey("/providers"),
			Datastore: ddbds.New(ddbClient, c.ProvidersTable,
				ddbds.WithPartitionkey("ContentHash"), ddbds.WithSortKey("ProviderID")),
		},
		{
			Prefix: datastore.NewKey("/pins"),
			Datastore: ddbds.New(ddbClient, c.PinsTable,
				ddbds.WithPartitionkey("Hash")),
		},
		{
			Prefix:    datastore.NewKey("/"),
			Datastore: ddbds.New(ddbClient, c.DefaultTable),
		},
	})

	return ddbDS, nil
}
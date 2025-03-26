# go-ds-dynamodb

A DynamoDB Datastore Implementation

This is an implementation of go-datastore that is backed by DynamoDB.

ddbds includes support for optimized prefix queries. When you set up your table's key schema correctly and register it with ddbds, then incoming queries that match the schema will be converted into DynamoDB queries instead of table scans, enabling high-performance, ordered, high-cardinality prefix queries.

Note that ddbds currently only stores values up to 400 KB (the DynamoDB maximum item size). This makes ddbds inappropriate for block storage. It could be extended to fall back to S3, but that is not yet implemented. Within the InterPlanetary ecosystem, it's designed for storing DHT records, IPNS records, peerstore records, etc.

## Setup

### Simple Setup with Unoptimized Queries

ddbds can be used as a simple key-value store, without optimized queries.

In this case, all datastore queries will result in full table scans using the ParallelScan API, and filtering/ordering/etc. will be performed client-side.

This is a good option if your table is small or your data and access patterns would not significantly benefit from optimized queries.

```go
var ddbClient *dynamodb.DynamoDB = ...
tableName := "datastore-table"
ddbDS := ddbds.New(ddbClient, tableName)
```

By default, the expected partition key is DSKey of type string. The name can be customized with the `WithPartitionKey()` option.

### Optimized Queries

To use optimized prefix queries, you must specify a sort key.

Also, elements written into the datastore should have at least two parts, such as `/a/b` and not `/a`.

ddbds splits the key into partition and sort keys. Examples:

- `/a` -> error (not enough parts)
- `/a/b` -> `[a, b]`
- `/a/b/c` -> `[a, b/c]`

To use optimized queries, simply specify the sort key name using the `WithSortKey()` option:

```go
var ddbClient *dynamodb.DynamoDB = ...
tableName := "datastore-table"
ddbDS := ddbds.New(
	ddbClient, 
	tableName,
	ddbds.WithPartitionKey("PartitionKey"),
	ddbds.WithSortKey("SortKey"),
)
```

### Dynamic Table Names

This implementation now supports dynamic table names, allowing users to specify table names at runtime when mounting different namespaces. This flexibility enables better separation of concerns and fine-tuned access control.

#### Variables for Table Names
To configure dynamic table names, the following variables should be passed:

- `providersTable`: Table name for `/providers` namespace.
- `pinsTable`: Table name for `/pins` namespace.
- `defaultTable`: Table name for all other data.

Example:

```go
var ddbClient *dynamodb.DynamoDB = ...
config := DDBConfig{
	providersTable: "datastore-providers",
	pinsTable: "datastore-pins",
	defaultTable: "datastore-default",
}

ddbDS := mount.New([]mount.Mount{
	{
		Prefix: ds.NewKey("/providers"),
		Datastore: ddbds.New(
			ddbClient,
			config.providersTable,
			ddbds.WithPartitionKey("ContentHash"),
			ddbds.WithSortKey("ProviderID"),
		),
	},
	{
		Prefix: ds.NewKey("/pins"),
		Datastore: ddbds.New(
			ddbClient,
			config.pinsTable,
			ddbds.WithPartitionKey("Hash"),
		),
	},
	{
		Prefix: ds.NewKey("/"),
		Datastore: ddbds.New(
			ddbClient,
			config.defaultTable,
		),
	},
})
```

## IAM Permissions

The following describes the IAM actions and the datastore methods that use them:

- **dynamodb:GetItem**
  - `Get()`
  - `GetExpiration()`
  - `GetSize()`
  - `Has()`
- **dynamodb:PutItem**
  - `Put()`
  - `PutWithTTL()`
- **dynamodb:DeleteItem**
  - `Delete()`
- **dynamodb:Scan** (if there is no sort key defined)
  - `Scan()`
- **dynamodb:Query** (if there is a sort key defined)
  - `Query()`
- **dynamodb:DescribeTable**
  - `DiskUsage()`
- **dynamodb:UpdateItem**
  - `SetTTL()`
- **dynamodb:BatchWriteItem**
  - `Batch.Commit()`

## Datastore Features

- ✅ Batching
- ✅ TTL
- ✅ Disk Usage
- ✅ Transactions
- ❌ Checked (not applicable)
- ❌ Scrubbed (not applicable)
- ❌ Garbage Collection (not applicable)

## Contributing

Contributions are welcome! This repository is part of the IPFS project and therefore governed by our contributing guidelines.

## License

SPDX-License-Identifier: Apache-2.0 OR MIT


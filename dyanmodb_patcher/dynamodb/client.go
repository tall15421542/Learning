package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	appConfig "github.com/deliveryhero/pd-dine-in-box/script/dynamodb_patcher/config"
)

type DDBClient interface {
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

type Client struct {
	ddbClient DDBClient
}

var distinctClients = make(map[string]*Client)

func distinctClientKey(region, profile string) string {
	return fmt.Sprintf("%s#%s", region, profile)
}

func NewClient(awsCfg appConfig.AWS) (*Client, error) {
	key := distinctClientKey(awsCfg.Region, awsCfg.Profile)

	if client, ok := distinctClients[key]; ok {
		return client, nil
	}

	options := []func(*config.LoadOptions) error{
		config.WithRegion(awsCfg.Region),
		config.WithSharedConfigProfile(awsCfg.Profile),
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), options...)
	if err != nil {
		return nil, err
	}

	ddbClient := dynamodb.NewFromConfig(cfg)
	client := &Client{ddbClient: ddbClient}
	distinctClients[key] = client

	return client, nil
}

func (c *Client) QueryAll(ctx context.Context, in *dynamodb.QueryInput, out interface{}) error {
	var allItems []map[string]types.AttributeValue
	var lastEvaluatedKey map[string]types.AttributeValue
	var err error

	for {
		in.ExclusiveStartKey = lastEvaluatedKey
		response, err := c.ddbClient.Query(ctx, in)
		if err != nil {
			return fmt.Errorf("fail to Query ddb: %w", err)
		}

		allItems = append(allItems, response.Items...)
		lastEvaluatedKey = response.LastEvaluatedKey

		if lastEvaluatedKey == nil {
			break
		}
	}

	err = attributevalue.UnmarshalListOfMaps(allItems, out)
	if err != nil {
		return fmt.Errorf("failed to unmarshal ddb items: %w", err)
	}

	return nil
}

func (c *Client) UpdateItem(ctx context.Context, in *dynamodb.UpdateItemInput, out interface{}) error {
	var output *dynamodb.UpdateItemOutput
	output, err := c.ddbClient.UpdateItem(ctx, in)

	if err != nil {
		return err
	}

	if output == nil {
		return fmt.Errorf("in=%+v didn't update the attribute", in)
	}

	err = attributevalue.UnmarshalMap(output.Attributes, out)
	return err
}

package tovendor

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/deliveryhero/pd-dine-in-box/script/dynamodb_patcher/config"
	"github.com/deliveryhero/pd-dine-in-box/script/dynamodb_patcher/utils"
)

type DDBRepository struct {
	ddbClient
	env          utils.Env
	globalEntity utils.GlobalEntity
	tableName    string
}

type Vendor struct {
	Code           string `dynamodbav:"vendor_code"`
	Name           string `dynamodbav:"name"`
	LocalLegalName string `dynamodbav:"local_legal_name"`
}

type ddbClient interface {
	QueryAll(ctx context.Context, in *dynamodb.QueryInput, out interface{}) error
	UpdateItem(ctx context.Context, in *dynamodb.UpdateItemInput, out interface{}) error
}

func NewDDBRepository(ge utils.GlobalEntity, cfg config.Config, client ddbClient) *DDBRepository {
	return &DDBRepository{
		ddbClient:    client,
		globalEntity: ge,
		tableName:    cfg.AWS.DynamoDBTableName,
	}
}

func (s *DDBRepository) GetAllVendors(ctx context.Context) ([]Vendor, error) {
	var vendors []Vendor
	queryInput, err := s.getAllVendorsQueryInput()
	if err != nil {
		return nil, err
	}

	err = s.ddbClient.QueryAll(ctx, queryInput, &vendors)
	if err != nil {
		return nil, fmt.Errorf("fail to query all vendors: %w", err)
	}

	return vendors, nil
}

func (s *DDBRepository) getAllVendorsQueryInput() (*dynamodb.QueryInput, error) {
	keyEx := expression.Key(pk).Equal(expression.Value(vendorPK(s.globalEntity.ID)))
	keyEx = keyEx.And(expression.Key(sk).BeginsWith(fmt.Sprintf("GEID#%s,VENDOR", s.globalEntity.ID)))

	projection := expression.NamesList(
		expression.Name("vendor_code"),
		expression.Name("name"),
		expression.Name("local_legal_name"),
	)

	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).WithProjection(projection).Build()

	if err != nil {
		return nil, err
	}

	return &dynamodb.QueryInput{
		TableName:                 aws.String(s.tableName),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
	}, nil
}

func (s *DDBRepository) UpdateLocalLegalName(ctx context.Context, vendorCode, localLegalName string) error {
	in, err := s.updateLocalLegalNameInput(vendorCode, localLegalName)
	if err != nil {
		return err
	}

	var vendor Vendor

	err = s.ddbClient.UpdateItem(ctx, in, &vendor)
	if err != nil {
		return err
	}

	log.Printf("Updated vendor attrs: %v\n", vendor)
	return nil
}

func (s *DDBRepository) updateLocalLegalNameInput(vendorCode, localLegalName string) (*dynamodb.UpdateItemInput, error) {
	update := expression.Set(expression.Name("local_legal_name"), expression.Value(localLegalName))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return nil, err
	}

	return &dynamodb.UpdateItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			pk: &types.AttributeValueMemberS{Value: vendorPK(s.globalEntity.ID)},
			sk: &types.AttributeValueMemberS{Value: vendorSK(s.globalEntity.ID, vendorCode)},
		},
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueUpdatedNew,
	}, nil
}

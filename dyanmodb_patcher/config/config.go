package config

import (
	"fmt"

	"github.com/deliveryhero/pd-dine-in-box/script/dynamodb_patcher/utils"
)

type Config struct {
	AWS
	VendorService
}

type AWS struct {
	Region            string
	Profile           string
	DynamoDBTableName string
}

type VendorService struct {
	EndpointFormatStr string
}

var prodConfig = Config{
	AWS: AWS{
		Region:            "ap-southeast-1",
		Profile:           "pd-production",
		DynamoDBTableName: "asia-prod-table-ordering",
	},
	VendorService: VendorService{
		EndpointFormatStr: "https://%s.fd-api.com/api/v1/vendor-service/vendors/%s",
	},
}

var stagingConfig = Config{
	AWS: AWS{
		Region:            "eu-central-1",
		Profile:           "pd-staging",
		DynamoDBTableName: "asia-staging-table-ordering",
	},
	VendorService: VendorService{
		EndpointFormatStr: "https://%s-st.fd-api.com/api/v1/vendor-service/vendors/%s",
	},
}

func GetByEnv(env utils.Env) (Config, error) {
	switch env {
	case utils.EnvStaging:
		{
			return stagingConfig, nil
		}
	case utils.EnvProd:
		{
			return prodConfig, nil
		}
	}
	return Config{}, fmt.Errorf("Invalid Env")
}

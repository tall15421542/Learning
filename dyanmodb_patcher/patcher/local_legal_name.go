package patcher

import (
	"context"
	"log"

	vendorSrv "github.com/deliveryhero/pd-dine-in-box/script/dynamodb_patcher/clients/vendor_service"
	"github.com/deliveryhero/pd-dine-in-box/script/dynamodb_patcher/repository/tovendor"
)

type vendorRepository interface {
	GetAllVendors(ctx context.Context) ([]tovendor.Vendor, error)
	UpdateLocalLegalName(ctx context.Context, vendorCode, localLegalName string) error
}

type LocalLegalNamePatcher struct {
	vendorRepository vendorRepository
	vendorSrvClient  *vendorSrv.Client
}

func (p *LocalLegalNamePatcher) Patch(ctx context.Context, vendor tovendor.Vendor) {
	// it is already updated by dine in worker.
	if vendor.LocalLegalName != "" {
		return
	}

	localLegalName, err := p.vendorSrvClient.GetLocalLegalName(vendor.Code)
	if err != nil {
		log.Printf("failed to get vendor local name, vendor code: %s, err: %v", vendor.Code, err)
		return
	}

	if localLegalName == "" {
		log.Printf("vendor %s does not have local legal name\n", vendor.Code)
		return
	}

	log.Printf("%s, %s, %s\n", vendor.Code, vendor.Name, localLegalName)
	p.vendorRepository.UpdateLocalLegalName(ctx, vendor.Code, localLegalName)
}

func (p *LocalLegalNamePatcher) ValidateEnvConfig() error {
	return p.vendorSrvClient.ValidateEnvConfig()
}

func NewLocalLegalNamePatcher(vendorRepo vendorRepository, vendorSrvClient *vendorSrv.Client) *LocalLegalNamePatcher {
	return &LocalLegalNamePatcher{
		vendorRepository: vendorRepo,
		vendorSrvClient:  vendorSrvClient,
	}
}

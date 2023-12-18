package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/joho/godotenv"

	vendorSrv "github.com/deliveryhero/pd-dine-in-box/script/dynamodb_patcher/clients/vendor_service"
	"github.com/deliveryhero/pd-dine-in-box/script/dynamodb_patcher/config"
	"github.com/deliveryhero/pd-dine-in-box/script/dynamodb_patcher/dynamodb"
	"github.com/deliveryhero/pd-dine-in-box/script/dynamodb_patcher/patcher"
	"github.com/deliveryhero/pd-dine-in-box/script/dynamodb_patcher/repository/tovendor"
	"github.com/deliveryhero/pd-dine-in-box/script/dynamodb_patcher/utils"
)

type Patcher interface {
	Patch(ctx context.Context, vendors tovendor.Vendor)
	ValidateEnvConfig() error
}

// declaration block for target name constants.
const (
	localLegalName = "local_legal_name"
)

// declaration block for application constants.
var (
	// allGEIDs is the entity id list that we deployed the dine-in service.
	allGEIDs = []string{
		"FP_BD",
		"FP_HK",
		"FP_MY",
		"FP_PH",
		"FP_PK",
		"FP_SG",
		"FP_TH",
		"FP_TW",
	}
)

// declaration block for flags.
var (
	envFlag               string
	geidsFlag             utils.GlobalEntitiesFlag
	targetFlag            string
	maxConcurrentTaskFlag uint
	isForAllEntitiesFlag  bool
)

func init() {
	flag.StringVar(&envFlag, "env", "staging", "staging or prod")
	flag.Var(&geidsFlag, "geid", "[Required] Comma separated list of Pandora Global Entity IDs. For example, \"FP_SG,FP_TW\". It's required when all flag is not set")
	flag.StringVar(&targetFlag, "target", "", "[Required] The target for this patch task. For example, local_legal_name.")
	flag.UintVar(&maxConcurrentTaskFlag, "n", 1, "The maximum concurrent Patch we would execute. Use sequential processing as default.")
	flag.BoolVar(&isForAllEntitiesFlag, "all", false, "Set true if you want to run the patch task for all entites in a env. It would ignore geid flag when it's set.")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	flag.Parse()

	err := validateRequiredFlags()
	if err != nil {
		log.Fatal(err)
	}

	env, err := utils.EnvFromString(envFlag)
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.GetByEnv(env)
	if err != nil {
		log.Fatalf("failed to get config: %v", err)
	}

	var geids []string
	if isForAllEntitiesFlag {
		geids = allGEIDs
	} else {
		geids = geidsFlag
	}

	var globalEntities []utils.GlobalEntity
	for _, geid := range geids {
		globalEntity, err := utils.NewGlobalEntity(geid)
		if err != nil {
			log.Fatal(err)
		}
		globalEntities = append(globalEntities, globalEntity)
	}

	for _, globalEntity := range globalEntities {
		patch(globalEntity, cfg, targetFlag, maxConcurrentTaskFlag)
	}
}

func patch(globalEntity utils.GlobalEntity, cfg config.Config, target string, maxConcurrentTask uint) {
	patchers := map[string]Patcher{}
	initializePatchers(patchers, globalEntity, cfg)

	patcher, err := getPatcherByTarget(patchers, target)
	if err != nil {
		log.Fatalf("Failed to get patcher by target: %v", err)
	}

	vendors, err := getAllVendors(globalEntity, cfg)
	if err != nil {
		log.Fatalf("Failed to get vendor list: %v", err)
	}

	var wg sync.WaitGroup
	guard := make(chan struct{}, maxConcurrentTask)

	for _, vendor := range vendors {
		wg.Add(1)
		guard <- struct{}{}

		go func(vendor tovendor.Vendor) {
			patcher.Patch(context.TODO(), vendor)
			wg.Done()
			<-guard
		}(vendor)
	}
	wg.Wait()

	log.Printf("Completed patching for %v vendors in %s", len(vendors), globalEntity.ID)
}

func validateRequiredFlags() error {
	if !isForAllEntitiesFlag && geidsFlag == nil {
		return fmt.Errorf("geid flag is required when 'all' flag is not set")
	}

	if targetFlag == "" {
		return fmt.Errorf("target flag is required")
	}

	return nil
}

func initializePatchers(patchers map[string]Patcher, globalEntity utils.GlobalEntity, cfg config.Config) {
	ddbClient, err := dynamodb.NewClient(cfg.AWS)
	if err != nil {
		log.Fatal(err)
	}

	vendorRepository := tovendor.NewDDBRepository(globalEntity, cfg, ddbClient)

	httpClient := &http.Client{}
	vendorSrvClient := vendorSrv.NewClient(globalEntity, cfg, httpClient)

	// register available patchers
	patchers[localLegalName] = patcher.NewLocalLegalNamePatcher(vendorRepository, vendorSrvClient)
}

func getPatcherByTarget(patchers map[string]Patcher, target string) (Patcher, error) {
	patcher, ok := patchers[targetFlag]
	if !ok {
		return nil, fmt.Errorf("Unsupported target: %s", targetFlag)
	}

	if err := patcher.ValidateEnvConfig(); err != nil {
		return nil, fmt.Errorf("Environment variable for target %s is not ready: %v", targetFlag, err)
	}

	return patcher, nil
}

func getAllVendors(globalEntity utils.GlobalEntity, cfg config.Config) ([]tovendor.Vendor, error) {
	ddbClient, err := dynamodb.NewClient(cfg.AWS)
	if err != nil {
		return nil, err
	}

	vendorRepository := tovendor.NewDDBRepository(globalEntity, cfg, ddbClient)
	vendors, err := vendorRepository.GetAllVendors(context.TODO())
	if err != nil {
		return nil, err
	}
	return vendors, nil
}

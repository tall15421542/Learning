package utils

import (
	"fmt"
	"strings"

	pdkit "github.com/deliveryhero/pd-go-kit"
)

type GlobalEntity struct {
	ID string
	pdkit.GlobalEntity
}

func NewGlobalEntity(geid string) (GlobalEntity, error) {
	globalEntity, ok := pdkit.GlobalEntities[geid]
	if !ok {
		return GlobalEntity{}, fmt.Errorf("Invalid GEID %s", geid)
	}
	return GlobalEntity{
		ID:           geid,
		GlobalEntity: globalEntity,
	}, nil
}

type GlobalEntitiesFlag []string

func (g *GlobalEntitiesFlag) String() string {
	return fmt.Sprint(*g)
}

func (g *GlobalEntitiesFlag) Set(value string) error {
	for _, geid := range strings.Split(value, ",") {
		*g = append(*g, geid)
	}
	return nil
}

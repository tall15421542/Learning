package tovendor

import (
	"fmt"
)

const (
	pk = "PK"
	sk = "SK"
)

func vendorPK(geid string) string {
	return fmt.Sprintf("GEID#%s", geid)
}

func vendorSK(geid, vendorCode string) string {
	return fmt.Sprintf("GEID#%s,VENDOR#%s", geid, vendorCode)
}

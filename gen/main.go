package main

import (
	cbg "github.com/whyrusleeping/cbor-gen"
	"github.com/whyrusleeping/demo-atp/records"
)

func main() {
	genCfg := cbg.Gen{
		MaxStringLength: 1_000_000,
	}

	if err := genCfg.WriteMapEncodersToFile("records/cbor_gen.go", "records", records.Profile{}, records.Comment{}); err != nil {
		panic(err)
	}
}

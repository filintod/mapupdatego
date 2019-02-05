package updater

import (
	"fmt"
	"k8s.io/helm/pkg/chartutil"
	"testing"
)

func TestLoadChart(t *testing.T) {
	c, err := chartutil.Load("testdata/starters/base")
	nv, err := chartutil.ReadValues([]byte(c.Values.Raw))
	for k, v := range nv {
		fmt.Print(k)
		coalesce(v, v, nil)

	}
	if err != nil {
		t.Error("Problems loading base chart")
	} else {
		print(nv)
	}
}

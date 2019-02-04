package updater

import (
	"k8s.io/helm/pkg/chartutil"
	"testing"
)

func TestLoadChart(t *testing.T) {
	c, err := chartutil.Load("testdata/starters/base")
	nv, err := chartutil.ReadValues([]byte(c.Values.Raw))

	if err != nil {
		t.Error("Problems loading base chart")
	} else {
		print(nv)
	}
}

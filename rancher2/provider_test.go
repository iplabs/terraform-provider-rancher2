package rancher2

import (
	"os"
	"testing"
)

func TestReadCLIConfiguration(t *testing.T) {
	tt := []struct {
		filename    string
		errExpected bool
	}{
		{
			filename:    "./testdata/cli2_valid.json",
			errExpected: false,
		},
	}
	for _, td := range tt {
		file, err := os.Open(td.filename)
		if err != nil {
			t.Errorf("unable to open configuration file %s: %v", td.filename, err)
			t.Fail()
		}
		_, err = readCLIConfiguration(file)
		if (td.errExpected && err == nil) || (!td.errExpected && err != nil) {
			t.Errorf("unexpected error result: %v", err)
		}
		file.Close()
	}
}

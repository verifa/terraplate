package runner

import (
	"sync"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
	"github.com/verifa/terraplate/parser"
)

func TestMultipleRunError(t *testing.T) {
	// Setup TerraRun
	tf := parser.DefaultTerrafile
	tf.Dir = "testData"
	r := TerraRun{
		Terrafile: &tf,
	}
	// Run the TerraRun twice, and one of those runs should result in an error
	var (
		wg  sync.WaitGroup
		err error
	)
	{
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = multierror.Append(err, r.Run(TerraRunOpts{
				validate: true,
				init:     true,
				plan:     true,
				jobs:     1,
			}))
		}()
	}
	{
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = multierror.Append(err, r.Run(TerraRunOpts{
				validate: true,
				init:     true,
				plan:     true,
				jobs:     1,
			}))
		}()
	}
	wg.Wait()
	assert.ErrorIs(t, err, ErrRunInProgress)
}

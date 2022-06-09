package notify

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo(t *testing.T) {
	// Cannot really assert repo values as it depends on the origin value and branch...
	_, err := LookupRepo()
	require.NoError(t, err)
}

func TestRepoWithEnv(t *testing.T) {
	const (
		testRepoName   = "repoName"
		testRepoBranch = "repoBranch"
	)
	os.Setenv(repoNameEnv, testRepoName)
	os.Setenv(repoBranchEnv, testRepoBranch)
	repo, err := LookupRepo()
	require.NoError(t, err)
	assert.Equal(t, testRepoName, repo.Name)
	assert.Equal(t, testRepoBranch, repo.Branch)
}

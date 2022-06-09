package notify

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/verifa/terraplate/parser"
)

var errRepoDetachedHead = errors.New("HEAD is detached")

const (
	repoNameEnv   = "TP_REPO_NAME"
	repoBranchEnv = "TP_REPO_BRANCH"

	gitDefaultRemote = "origin"
)

type repoOptFunc func(n *Repo)

func WithRepo(repo Repo) repoOptFunc {
	return func(r *Repo) {
		*r = repo
	}
}

func WithRepoName(name string) repoOptFunc {
	return func(r *Repo) {
		r.Name = name
	}
}

func WithRepoBranch(branch string) repoOptFunc {
	return func(r *Repo) {
		r.Branch = branch
	}
}

type Repo struct {
	Name   string
	Branch string
}

// LookupRepo gets the repository data for the drift detection.
// It reads from environment variables (if set) otherwise fetches the information
// from the local git repository
func LookupRepo(opts ...repoOptFunc) (*Repo, error) {
	var repo Repo
	for _, opt := range opts {
		opt(&repo)
	}

	var (
		hasRepoName   = repo.Name != ""
		hasRepoBranch = repo.Branch != ""
	)

	if !hasRepoName {
		repo.Name, hasRepoName = os.LookupEnv(repoNameEnv)
	}
	if !hasRepoBranch {
		repo.Branch, hasRepoBranch = os.LookupEnv(repoBranchEnv)
	}

	if hasRepoBranch && hasRepoName {
		return &repo, nil
	}

	gitRepo, gitErr := findGitRepo()
	if gitErr != nil {
		return nil, fmt.Errorf("reading git repo details: %w", gitErr)
	}

	if !hasRepoName {
		var remoteErr error
		repo.Name, remoteErr = repoRemote(gitRepo)
		if remoteErr != nil {
			return nil, remoteErr
		}
	}
	if !hasRepoBranch {
		var branchErr error
		repo.Branch, branchErr = repoBranch(gitRepo)
		if branchErr != nil {
			if !errors.Is(branchErr, errRepoDetachedHead) {
				return nil, branchErr
			}
			repo.Branch = "Detached HEAD"
		}
	}

	return &repo, nil
}

// findGitRepo returns the remote, branch and error (if any)
func findGitRepo() (*git.Repository, error) {
	var gitRepo string
	travErr := parser.TraverseUpDirectory(".", func(dir string) (bool, error) {
		stat, statErr := os.Stat(filepath.Join(dir, ".git"))
		if statErr != nil {
			if os.IsNotExist(statErr) {
				// This is fine, so keep traversing and no error
				return true, nil
			}

			return false, statErr
		}
		if !stat.IsDir() {
			// This would be strange, but whatever, let's continue
			return true, nil
		}
		// We have found a .git directory... Let's use it
		gitRepo = dir
		// Gracefully stop the traversal
		return false, nil
	})
	if travErr != nil {
		return nil, fmt.Errorf("looking for git repository: %w", travErr)
	}
	repo, openErr := git.PlainOpen(gitRepo)
	if openErr != nil {
		return nil, fmt.Errorf("opening git repo: %w", openErr)
	}
	return repo, nil
}

func repoRemote(repo *git.Repository) (string, error) {
	remote, remoteErr := repo.Remote(gitDefaultRemote)
	if remoteErr != nil {
		return "", fmt.Errorf("getting remote from repository: %w", remoteErr)
	}
	// Don't know if it's possible, but just to avoid a nasty bug
	if len(remote.Config().URLs) == 0 {
		return "", fmt.Errorf("git remote has no URLs")
	}
	remoteURL := remote.Config().URLs[0]

	return remoteURL, nil
}

func repoBranch(repo *git.Repository) (string, error) {
	head, headErr := repo.Head()
	if headErr != nil {
		return "", fmt.Errorf("getting HEAD from repository: %w", headErr)
	}

	if !head.Name().IsBranch() {
		return "", fmt.Errorf("getting branch from repository: %w", errRepoDetachedHead)
	}

	return head.Name().Short(), nil
}

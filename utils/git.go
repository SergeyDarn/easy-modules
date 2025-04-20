package utils

import (
	"regexp"
	"slices"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

var GIT_DEFAULT_BRANCHES = []plumbing.ReferenceName{
	plumbing.NewBranchReferenceName("main"),
	plumbing.NewBranchReferenceName("master"),
}

const GIT_URL_SEPARATOR = "#"
const GIT_NORMAL_URL = "http"

func GitClone(
	repoName string,
	repoUrl string,
	repoDirPath string,
	sshKeyPath string,
	sshKeyPassword string,
) {
	log.Debug("Cloning", "repo", repoName, "url", repoUrl)

	cleanModuleUrl, reference, commitHash, _, _ := parseGitUrl(repoUrl)
	options := &git.CloneOptions{
		URL:           cleanModuleUrl,
		ReferenceName: reference,
	}

	auth := getGitAuth(repoUrl, sshKeyPath, sshKeyPassword)
	if auth != nil {
		options.Auth = auth
	}

	repo, err := git.PlainClone(repoDirPath, false, options)
	CheckError(err, "Error while clonning repo "+repoName)

	if commitHash != "" {
		GitCheckout(repo, commitHash, "")
	}

	ref, err := repo.Head()
	CheckError(err, "Error while getting repo head")

	log.Debug("Cloning successful.", "repo", repoName, "head", ref.String())
}

func GitDirStatus(dirPath string) git.Status {
	repo, err := git.PlainOpen(dirPath)
	CheckError(err, "Error while trying to open module directory in git for gitFolderStatus")

	workTree, err := repo.Worktree()
	CheckError(err, "Error while getting workTree for gitFolderStatus")

	status, err := workTree.Status()
	CheckError(err, "Error while trying to get git status")

	return status
}

func GitCheckout(
	repo *git.Repository,
	commitHash string,
	branchOrTag plumbing.ReferenceName,
) {
	workTree, err := repo.Worktree()
	CheckError(err, "Error while getting repo head")

	var hashObject plumbing.Hash
	var processedReference plumbing.ReferenceName = branchOrTag

	if commitHash != "" {
		hashObject = plumbing.NewHash(commitHash)
	} else if branchOrTag == "" {
		processedReference = getDefaultBranch(repo)
	}

	err = workTree.Checkout(&git.CheckoutOptions{
		Hash:   hashObject,
		Branch: processedReference,
	})
	CheckError(err, "Error while trying to checkout commitHash="+commitHash+" branchOrTag="+processedReference.String())

	log.Debug("Successful git checkout to", "commitHash", commitHash, "branchOrTag", branchOrTag)
}

func IsGitUrl(url string) bool {
	return strings.Contains(url, "git")
}

// Returns cleanUrl, reference, commitHash, branch, tag
func parseGitUrl(gitUrl string) (
	string,
	plumbing.ReferenceName,
	string,
	plumbing.ReferenceName,
	plumbing.ReferenceName,
) {
	if !IsGitUrl(gitUrl) {
		return "", "", "", "", ""
	}

	splitUrl := strings.Split(gitUrl, GIT_URL_SEPARATOR)
	cleanUrl := splitUrl[0]
	hasReference := len(splitUrl) > 1

	if !hasReference {
		return cleanUrl, "", "", "", ""
	}

	baseReference := splitUrl[1]
	reference, commitHash, branch, tag := prepareGitReference(baseReference)

	return cleanUrl, reference, commitHash, branch, tag
}

// Returns reference, commitHash, branch, tag
func prepareGitReference(baseReference string) (
	plumbing.ReferenceName,
	string,
	plumbing.ReferenceName,
	plumbing.ReferenceName,
) {
	var reference plumbing.ReferenceName
	var commitHash string
	var branch plumbing.ReferenceName
	var tag plumbing.ReferenceName

	tagRegexp, _ := regexp.Compile(`\d(\..*)+`)

	if tagRegexp.MatchString(baseReference) {
		tag = plumbing.NewTagReferenceName(baseReference)
		reference = tag
	} else if plumbing.IsHash(baseReference) {
		commitHash = baseReference
	} else {
		branch = plumbing.NewBranchReferenceName(baseReference)
		reference = branch
	}

	log.Debug("Preparing reference", "tag", tag, "commitHash", commitHash, "branch", branch)

	return reference, commitHash, branch, tag
}

func getDefaultBranch(repo *git.Repository) plumbing.ReferenceName {
	references, err := repo.References()
	CheckError(err, "Error while trying to get repo remotes when getting default branch")

	var defaultBranch plumbing.ReferenceName
	err = references.ForEach(func(reference *plumbing.Reference) error {
		if slices.Contains(GIT_DEFAULT_BRANCHES, reference.Name()) {
			defaultBranch = reference.Name()
		}

		return nil
	})
	CheckError(err, "Error while iterating remotes when getting default branch")

	log.Debug("Got default branch", "branch", defaultBranch)

	return defaultBranch
}

func getGitAuth(repoUrl string, sshKeyPath string, sshKeyPassword string) *ssh.PublicKeys {
	if strings.Contains(repoUrl, GIT_NORMAL_URL) {
		return nil
	}

	auth, err := ssh.NewPublicKeysFromFile("git", sshKeyPath, sshKeyPassword)
	CheckError(err, "Error while creating git clone auth")

	return auth
}

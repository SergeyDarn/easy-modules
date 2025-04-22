package utils

import (
	"regexp"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

var GIT_DEFAULT_BRANCHES = []plumbing.ReferenceName{
	plumbing.NewBranchReferenceName("main"),
	plumbing.NewBranchReferenceName("master"),
}

const (
	GIT_URL_SEPARATOR = string('#')
	GIT_NORMAL_URL    = "http"
	TAG_REGEXP        = `\d(\..*)+`
)

const (
	REPO_COLOR   = lipgloss.Color("#f98b6c")
	URL_COLOR    = lipgloss.Color("#4a26fd")
	TAG_COLOR    = lipgloss.Color("#ef4fa6")
	HASH_COLOR   = lipgloss.Color("#1E88E5")
	BRANCH_COLOR = lipgloss.Color("#ff9e22")
)

func GitClone(
	repoName string,
	repoUrl string,
	repoDirPath string,
	sshKeyPath string,
	sshKeyPassword string,
) {
	repoOutput := prepareGitColorOutput("repo="+repoName, REPO_COLOR) + " "
	urlOutput := prepareGitColorOutput("url="+repoUrl, URL_COLOR)
	log.Debug("Cloning " + repoOutput + urlOutput)

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
	CheckError(err, "Error while clonning "+repoOutput)

	if commitHash != "" {
		GitCheckout(repo, commitHash, "")
	}

	ref, err := repo.Head()
	CheckError(err, "Error while getting repo head")
	headOutput := prepareGitColorOutput("head="+ref.String(), HASH_COLOR)

	successOuput := PrepareColorOutput("Cloning successful ", SUCCESS_COLOR)
	log.Debug(successOuput + repoOutput + headOutput)
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
	commitOutput := "commitHash=" + commitHash
	branchOrTagOutput := "branchOrTag=" + string(branchOrTag)
	CheckError(err, "Error while trying to checkout "+commitOutput+" "+branchOrTagOutput)

	commitColorOutput := prepareGitColorOutput(commitOutput, HASH_COLOR)
	branchOrTagColorOutput := prepareGitColorOutput(branchOrTagOutput, BRANCH_COLOR)
	log.Debug("Successful git checkout to " + commitColorOutput + " " + branchOrTagColorOutput)
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

	tagRegexp, _ := regexp.Compile(TAG_REGEXP)

	if tagRegexp.MatchString(baseReference) {
		tag = plumbing.NewTagReferenceName(baseReference)
		reference = tag
	} else if plumbing.IsHash(baseReference) {
		commitHash = baseReference
	} else {
		branch = plumbing.NewBranchReferenceName(baseReference)
		reference = branch
	}

	tagOutput := prepareGitColorOutput("tag="+string(tag), TAG_COLOR) + " "
	commitOutput := prepareGitColorOutput("commitHash="+commitHash, HASH_COLOR) + " "
	branchOutput := prepareGitColorOutput("branch="+string(branch), BRANCH_COLOR) + " "
	log.Debug("Parsed from git url " + tagOutput + commitOutput + branchOutput)

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

	log.Debug("Got default " + prepareGitColorOutput("branch "+string(defaultBranch), BRANCH_COLOR))

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

func prepareGitColorOutput(output string, color lipgloss.Color) string {
	return PrepareColorOutput(output, color)
}

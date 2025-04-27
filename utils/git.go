package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

const (
	GIT_URL_SEPARATOR = "#"
	GIT_NORMAL_URL    = "http"
	TAG_REGEXP        = `\d(\..*)+`
)

const (
	REPO_COLOR   = lipgloss.Color("#f98b6c")
	URL_COLOR    = lipgloss.Color("#4a26fd")
	TAG_COLOR    = lipgloss.Color("#ef4fa6")
	HASH_COLOR   = lipgloss.Color("#1e88e5")
	BRANCH_COLOR = lipgloss.Color("#03dac6")
)

func GitClone(
	repoName string,
	repoUrl string,
	repoDirPath string,
) {
	repoLog := prepareGitColorOutput("repo="+repoName, REPO_COLOR)
	urlLog := prepareGitColorOutput("url="+repoUrl, URL_COLOR)
	log.Debugf("Cloning %s %s", repoLog, urlLog)

	cleanModuleUrl, commitHash, branch, tag := parseGitUrl(repoUrl)
	reference := branch
	if tag != "" {
		reference = tag
	}

	options := &git.CloneOptions{
		URL:           cleanModuleUrl,
		ReferenceName: reference,
	}

	auth := getGitAuth(repoUrl)
	if auth != nil {
		options.Auth = auth
	}

	repo, err := git.PlainClone(repoDirPath, false, options)
	CheckError(err, "Error while clonning "+repoLog)

	if commitHash != "" {
		GitCheckoutToCommit(repo, repoName, commitHash)
	}

	headName := GetHeadShortName(repo, commitHash != "", tag != "")
	headColor := getGitColor(commitHash != "", tag != "")

	headLog := prepareGitColorOutput("head="+headName, headColor)
	successLog := PrepareSuccessOutput("Cloning successful")
	log.Debugf("%s %s %s", successLog, repoLog, headLog)
}

func GitDirStatus(dirPath string) git.Status {
	repo, err := git.PlainOpen(dirPath)
	CheckError(err, "Error while trying to open module directory "+dirPath+" in git for gitFolderStatus")

	workTree, err := repo.Worktree()
	CheckError(err, "Error while getting workTree for gitFolderStatus")

	status, err := workTree.Status()
	CheckError(err, "Error while trying to get git status")

	return status
}

func GitCheckoutToCommit(
	repo *git.Repository,
	repoName string,
	commitHash string,
) {
	if commitHash == "" {
		return
	}

	workTree, err := repo.Worktree()
	CheckError(err, fmt.Sprintf("Error while getting repo %s worktree before checkout", repoName))

	err = workTree.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(commitHash),
	})

	repoLog := "repo=" + repoName
	commitLog := "commitHash=" + commitHash
	CheckError(err, fmt.Sprintf("Error while trying to checkout %s %s", repoLog, commitLog))

	repoColorLog := prepareGitColorOutput(repoLog, REPO_COLOR)
	commitColorLog := prepareGitColorOutput(commitLog, HASH_COLOR)
	log.Debugf("Sucessful checkout for %s to %s", repoColorLog, commitColorLog)
}

func IsGitUrl(url string) bool {
	return strings.Contains(url, "git")
}

func GetHeadShortName(repo *git.Repository, isCommit bool, isTag bool) string {
	head, err := repo.Head()
	CheckError(err, "Error while getting repo head")

	if isCommit {
		return head.Hash().String()
	}

	if isTag {
		headTag := GetHeadTag(repo)
		if headTag == nil {
			ThrowError("Couldn't find tag for head " + head.String())
		}

		return headTag.Name().Short()
	}

	return head.Name().Short()
}

func GetHeadTag(repo *git.Repository) *plumbing.Reference {
	head, err := repo.Head()
	CheckError(err, "Error while getting repo head (GetHeadTag)")

	tags, err := repo.Tags()
	CheckError(err, "Error while getting repo tags (GetHeadTag)")

	var headTag *plumbing.Reference

	tags.ForEach(func(tag *plumbing.Reference) error {
		tagHash, err := repo.ResolveRevision(plumbing.Revision(tag.Name()))
		CheckError(err, "Error while resolving revision (GetHeadTag)")

		if *tagHash == head.Hash() {
			headTag = tag
		}

		return nil
	})

	return headTag
}

// Returns cleanUrl, commitHash, branch, tag
func parseGitUrl(gitUrl string) (
	string,
	string,
	plumbing.ReferenceName,
	plumbing.ReferenceName,
) {
	if !IsGitUrl(gitUrl) {
		return "", "", "", ""
	}

	splitUrl := strings.Split(gitUrl, GIT_URL_SEPARATOR)
	cleanUrl := splitUrl[0]
	hasReference := len(splitUrl) > 1

	if !hasReference {
		return cleanUrl, "", "", ""
	}

	if (splitUrl[1] == "") || (len(splitUrl) > 2) {
		ThrowError("Cannot properly parse url " + gitUrl + " Aborting.")
	}

	baseReference := splitUrl[1]
	commitHash, branch, tag := prepareGitReference(baseReference)

	return cleanUrl, commitHash, branch, tag
}

// Returns commitHash, branch, tag
func prepareGitReference(baseReference string) (
	string,
	plumbing.ReferenceName,
	plumbing.ReferenceName,
) {
	var commitHash string
	var branch plumbing.ReferenceName
	var tag plumbing.ReferenceName

	tagRegexp, _ := regexp.Compile(TAG_REGEXP)

	if tagRegexp.MatchString(baseReference) {
		tag = plumbing.NewTagReferenceName(baseReference)
	} else if plumbing.IsHash(baseReference) {
		commitHash = baseReference
	} else {
		branch = plumbing.NewBranchReferenceName(baseReference)
	}

	baseLog := "Parsed from git url "

	if commitHash != "" {
		commitLog := prepareGitColorOutput("commitHash="+commitHash, HASH_COLOR)
		log.Debug(baseLog + commitLog)
	}

	if branch != "" {
		branchLog := prepareGitColorOutput("branch="+branch.Short(), BRANCH_COLOR)
		log.Debug(baseLog + branchLog)
	}

	if tag != "" {
		tagLog := prepareGitColorOutput("tag="+tag.Short(), TAG_COLOR) + " "
		log.Debug(baseLog + tagLog)
	}

	return commitHash, branch, tag
}

func getGitAuth(repoUrl string) *ssh.PublicKeys {
	if strings.Contains(repoUrl, GIT_NORMAL_URL) {
		return nil
	}

	sshKeyPath := GetEnv(ENV_SSH_KEY_PATH)
	sshKeyPassword := GetEnv(ENV_SSH_KEY_PASSWORD)

	auth, err := ssh.NewPublicKeysFromFile("git", sshKeyPath, sshKeyPassword)
	CheckError(err, "Error while creating git clone auth")

	return auth
}

func prepareGitColorOutput(output string, color lipgloss.Color) string {
	return PrepareColorOutput(output, color)
}

func getGitColor(isHash bool, isTag bool) lipgloss.Color {
	if isHash {
		return HASH_COLOR
	}

	if isTag {
		return TAG_COLOR
	}

	return BRANCH_COLOR
}

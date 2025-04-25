package utils

import (
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
)

func TestIsGitUrl(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"git", true},
		{"htts://google.com/", false},
		{"randomString", false},
		{"cool-js-module", false},
		{"https://github.com/SergeyDarn/scrape-search-ai.git", true},
		{"git@github.com:SergeyDarn/scrape-search-ai.git", true},
	}

	for _, test := range tests {
		testName := test.url

		t.Run(testName, func(t *testing.T) {
			res := IsGitUrl(test.url)

			if res != test.want {
				t.Errorf("Expected %t, but got %t", test.want, res)
			}
		})

		IsGitUrl(test.url)
	}
}

func TestParseGitUrl(t *testing.T) {
	type want struct {
		cleanUrl   string
		commitHash string
		branch     plumbing.ReferenceName
		tag        plumbing.ReferenceName
		error      bool
	}

	tests := []struct {
		name   string
		gitUrl string
		want   want
	}{
		{"Invalid url", "htts://google.com/", want{}},
		{"No reference http url", "https://github.com/SergeyDarn/scrape-search-ai.git", want{
			cleanUrl: "https://github.com/SergeyDarn/scrape-search-ai.git",
		}},
		{"No reference ssh url", "git@github.com:SergeyDarn/scrape-search-ai.git", want{
			cleanUrl: "git@github.com:SergeyDarn/scrape-search-ai.git",
		}},
		{"Http url with invalid reference", "https://github.com/SergeyDarn/scrape-search-ai#", want{
			cleanUrl: "https://github.com/SergeyDarn/scrape-search-ai",
		}},
		{"Ssh url with invalid reference", "git@github.com:SergeyDarn/scrape-search-ai.git#", want{
			cleanUrl: "git@github.com:SergeyDarn/scrape-search-ai.git",
		}},

		{"Branch 1", "git@github.com:SergeyDarn/scrape-search-ai.git#1", want{
			cleanUrl: "git@github.com:SergeyDarn/scrape-search-ai.git",
			branch:   plumbing.NewBranchReferenceName("1"),
		}},
		{"Branch 123", "git@github.com:SergeyDarn/scrape-search-ai.git#123", want{
			cleanUrl: "git@github.com:SergeyDarn/scrape-search-ai.git",
			branch:   plumbing.NewBranchReferenceName("123"),
		}},
		{"Branch dev", "git@github.com:SergeyDarn/scrape-search-ai.git#dev", want{
			cleanUrl: "git@github.com:SergeyDarn/scrape-search-ai.git",
			branch:   plumbing.NewBranchReferenceName("dev"),
		}},

		{"Branch 1fh1hfgdjjsadfg2", "git@github.com:SergeyDarn/scrape-search-ai.git#1fh1hfgdjjsadfg2", want{
			cleanUrl: "git@github.com:SergeyDarn/scrape-search-ai.git",
			branch:   plumbing.NewBranchReferenceName("1fh1hfgdjjsadfg2"),
		}},
		{"Commit Hash", "git@github.com:SergeyDarn/scrape-search-ai.git#b7620f64a115b85eca08504cb9b364e594c9f8df", want{
			cleanUrl:   "git@github.com:SergeyDarn/scrape-search-ai.git",
			commitHash: "b7620f64a115b85eca08504cb9b364e594c9f8df",
		}},

		{"Tag 1.4.0", "git@github.com:SergeyDarn/scrape-search-ai.git#1.4.0", want{
			cleanUrl: "git@github.com:SergeyDarn/scrape-search-ai.git",
			tag:      plumbing.NewTagReferenceName("1.4.0"),
		}},
		{"Tag 1.5", "git@github.com:SergeyDarn/scrape-search-ai.git#1.5", want{
			cleanUrl: "git@github.com:SergeyDarn/scrape-search-ai.git",
			tag:      plumbing.NewTagReferenceName("1.5"),
		}},
		{"Tag 1.10.0.1", "git@github.com:SergeyDarn/scrape-search-ai.git#1.10.0.1", want{
			cleanUrl: "git@github.com:SergeyDarn/scrape-search-ai.git",
			tag:      plumbing.NewTagReferenceName("1.10.0.1"),
		}},
		{"Tag 1.2.3_fix", "git@github.com:SergeyDarn/scrape-search-ai.git#1.2.3_fix", want{
			cleanUrl: "git@github.com:SergeyDarn/scrape-search-ai.git",
			tag:      plumbing.NewTagReferenceName("1.2.3_fix"),
		}},

		{"Icorrect tag 1.2.3#", "git@github.com:SergeyDarn/scrape-search-ai.git#1.2.3#", want{
			error: true,
		}},
		{"Icorrect url with multiple #", "git@g#ithub.com:Serg#eyDarns#ear.git#11.3#", want{
			error: true,
		}},
	}

	for _, test := range tests {
		testName := test.name
		if testName == "" {
			testName = test.gitUrl
		}

		t.Run(testName, func(t *testing.T) {
			if test.want.error {
				TestPanic(t, testName, func() { parseGitUrl(test.gitUrl) })
				return
			}

			cleanUrl, commitHash, branch, tag := parseGitUrl(test.gitUrl)
			fail := false

			if cleanUrl != test.want.cleanUrl {
				t.Logf("Expected cleanUrl %s, but got %s", test.want.cleanUrl, cleanUrl)
				fail = true
			}

			if commitHash != test.want.commitHash {
				t.Logf("Expected commitHash %s, but got %s", test.want.commitHash, commitHash)
				fail = true
			}

			if branch != test.want.branch {
				t.Logf("Expected branch %s, but got %s", test.want.branch, branch)
				fail = true
			}

			if tag != test.want.tag {
				t.Logf("Expected tag %s, but got %s", test.want.tag, tag)
				fail = true
			}

			if fail {
				t.Fail()
			}
		})
	}
}

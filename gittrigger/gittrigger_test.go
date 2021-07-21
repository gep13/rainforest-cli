package gitTrigger

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func deleteFakeRepo(t *testing.T, directory string) {
	err := os.RemoveAll(directory)
	if err != nil {
		t.Fatalf("Couldn't delete testing directory: %v.", err)
	}
}

func makeFakeRepoWithCommit(t *testing.T, commitMsg string) string {
	// create empty directory for repo
	dir, err := ioutil.TempDir("", "git_testing_directory")
	if err != nil {
		t.Fatal("Couldn't create directory for testing needs:", err)
	}
	os.Chdir(dir)
	if err != nil {
		deleteFakeRepo(t, dir)
		t.Fatal("Couldn't navigate to the directory for testing needs:", err)
	}

	// initialize repo
	cmd := exec.Command("git", "init")
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil || !strings.Contains(out.String(), "Initialized") {
		deleteFakeRepo(t, dir)
		t.Fatal("Couldn't initialize git repo for testing needs.")
	}

	// set username and email in local repo
	cmd = exec.Command("git", "config", "user.name", "'Rainforest QA'")
	out.Reset()
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		deleteFakeRepo(t, dir)
		t.Fatal("Couldn't set the username in repo.")
	}
	cmd = exec.Command("git", "config", "user.email", "'test@rainforestqa.com'")
	out.Reset()
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		deleteFakeRepo(t, dir)
		t.Fatal("Couldn't set the email in repo.")
	}
	cmd = exec.Command("git", "config", "commit.gpgSign", "false")
	out.Reset()
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		deleteFakeRepo(t, dir)
		t.Fatal("Couldn't set the email in repo.")
	}

	// create empty commit
	cmd = exec.Command("git", "commit", "--allow-empty", "-m", commitMsg)
	out.Reset()
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil || !strings.Contains(out.String(), commitMsg) {
		deleteFakeRepo(t, dir)
		t.Fatal("Couldn't commit to the test repo.")
	}

	return dir
}

func addFakeGitRemote(t *testing.T, remote_name string, remote_url string) {
	// create empty commit
	cmd := exec.Command("git", "remote", "add", remote_name, remote_url)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		t.Fatal("Couldn't add the remote in the test repo.")
	}
}

func TestNewGitTrigger(t *testing.T) {
	const commitMsg = "foo barred baz"
	dir := makeFakeRepoWithCommit(t, commitMsg)
	defer deleteFakeRepo(t, dir)
	git, err := NewGitTrigger()
	if err != nil {
		t.Error("Unexpected error when doing newGitTrigger()")
	}
	if git.LastCommit != commitMsg {
		t.Errorf("inproperly initialized gitTrigger with newGitTrigger: %v, expected: %v", git.LastCommit, commitMsg)
	}
}

func TestGetLatestCommit(t *testing.T) {
	const commitMsg = "test commit in a test repo"
	fakeGit := gitTrigger{Trigger: "@rainforest"}
	dir := makeFakeRepoWithCommit(t, commitMsg)
	defer deleteFakeRepo(t, dir)
	err := fakeGit.getLatestCommit()
	if err != nil {
		t.Error("Unexpected error when doing getLatestCommit()")
	}
	if fakeGit.LastCommit != commitMsg {
		t.Errorf("got wrong commit from getLatestCommit got: %v, expected: %v", fakeGit.LastCommit, commitMsg)
	}
}

func TestGetRemote(t *testing.T) {
	const expectedRemote = "git@github.com:rainforestapp/rainforest-cli.git"
	fakeGit := gitTrigger{Trigger: "@rainforest"}
	dir := makeFakeRepoWithCommit(t, "lol")
	addFakeGitRemote(t, "lol", expectedRemote)
	defer deleteFakeRepo(t, dir)
	remote, err := fakeGit.GetRemote()
	if err != nil {
		t.Errorf("Unexpected error when doing GetRemote(): %v", err)
	}
	if remote != expectedRemote {
		t.Errorf("got wrong remote from GetRemote got: %v, expected: %v", remote, expectedRemote)
	}
}

func TestGetRemoteTwoRemotes(t *testing.T) {
	const expectedRemote1 = "git@github.com:rainforestapp/rainforest-cli1.git"
	const expectedRemote2 = "git@github.com:rainforestapp/rainforest-cli2.git"
	fakeGit := gitTrigger{Trigger: "@rainforest"}
	dir := makeFakeRepoWithCommit(t, "lol")
	addFakeGitRemote(t, "lol1", expectedRemote1)
	addFakeGitRemote(t, "lol2", expectedRemote2)
	defer deleteFakeRepo(t, dir)
	remote, err := fakeGit.GetRemote()
	if err != nil {
		t.Errorf("Unexpected error when doing GetRemote(): %v", err)
	}
	if remote != expectedRemote1 {
		t.Errorf("got wrong remote from GetRemote got: %v, expected: %v", remote, expectedRemote1)
	}
}

func TestGetRemoteMissing(t *testing.T) {
	fakeGit := gitTrigger{Trigger: "@rainforest"}
	dir := makeFakeRepoWithCommit(t, "lol")
	defer deleteFakeRepo(t, dir)
	remote, err := fakeGit.GetRemote()
	if err == nil {
		t.Errorf("Expected GetRemote() to error, but it didn't")
	}
	if remote != "" {
		t.Errorf("got wrong remote from GetRemote got: %v, expected: ''", remote)
	}
}

func TestCheckTrigger(t *testing.T) {
	fakeGit := gitTrigger{Trigger: "@rainforest"}
	var testCases = []struct {
		fakeCommit string
		want       bool
	}{
		{
			fakeCommit: "Testing testing",
			want:       false,
		},
		{
			fakeCommit: "Testing @rainforest testing",
			want:       true,
		},
		{
			fakeCommit: "@rainfnf",
			want:       false,
		},
	}

	for _, tCase := range testCases {
		fakeGit.LastCommit = tCase.fakeCommit
		got := fakeGit.CheckTrigger()
		if !reflect.DeepEqual(tCase.want, got) {
			t.Errorf("checkTrigger returned %+v, want %+v", got, tCase.want)
		}
	}
}

func TestGetTags(t *testing.T) {
	fakeGit := gitTrigger{Trigger: "@rainforest"}
	var testCases = []struct {
		fakeCommit string
		want       []string
	}{
		{
			fakeCommit: "Testing testing",
			want:       []string{},
		},
		{
			fakeCommit: "@rainforest #foo, #bar",
			want:       []string{"foo", "bar"},
		},
		{
			fakeCommit: "@rainforest #foo #bar-baz #qwe_asd",
			want:       []string{"foo", "bar-baz", "qwe_asd"},
		},
	}

	for _, tCase := range testCases {
		fakeGit.LastCommit = tCase.fakeCommit
		got := fakeGit.GetTags()
		if !reflect.DeepEqual(tCase.want, got) {
			t.Errorf("getTags returned %+v, want %+v", got, tCase.want)
		}
	}
}

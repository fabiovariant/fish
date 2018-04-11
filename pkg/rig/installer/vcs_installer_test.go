package installer

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Masterminds/vcs"

	"github.com/fish"
	"github.com/fish/pkg/rig"
)

var _ Installer = new(VCSInstaller)

type testRepo struct {
	local, remote, current string
	tags, branches         []string
	err                    error
	vcs.Repo
}

func (r *testRepo) LocalPath() string           { return r.local }
func (r *testRepo) Remote() string              { return r.remote }
func (r *testRepo) Update() error               { return r.err }
func (r *testRepo) Get() error                  { return r.err }
func (r *testRepo) IsReference(string) bool     { return false }
func (r *testRepo) Tags() ([]string, error)     { return r.tags, r.err }
func (r *testRepo) Branches() ([]string, error) { return r.branches, r.err }
func (r *testRepo) UpdateVersion(version string) error {
	r.current = version
	return r.err
}

func TestVCSInstallerSuccess(t *testing.T) {
	dh, err := ioutil.TempDir("", "fish-home-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dh)

	home := fish.Home(dh)
	if err := os.MkdirAll(home.Rigs(), 0755); err != nil {
		t.Fatalf("Could not create %s: %s", home.Rigs(), err)
	}

	source := "https://github.com/org/fish-env"
	testRepoPath, _ := filepath.Abs("testdata/fish-food")
	repo := &testRepo{
		local: testRepoPath,
		tags:  []string{"0.1.0", "0.1.1"},
	}

	i, err := New(source, "~0.1.0", home)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// ensure a VCSInstaller was returned
	vcsInstaller, ok := i.(*VCSInstaller)
	if !ok {
		t.Error("expected a VCSInstaller")
	}

	// set the testRepo in the VCSInstaller
	vcsInstaller.Repo = repo

	if err := Install(i); err != nil {
		t.Error(err)
	}
	if repo.current != "0.1.1" {
		t.Errorf("expected version '0.1.1', got %q", repo.current)
	}
	expectedPath := home.Path("Rigs", "github.com", "org", "fish-env")
	if i.Path() != expectedPath {
		t.Errorf("expected path '%s', got %q", expectedPath, i.Path())
	}

	// Install again to test rig exists error
	if err := Install(i); err != nil {
		t.Errorf("expected nil when adding a rig that already exists, got %v", err)
	}
}

func TestVCSInstallerUpdate(t *testing.T) {
	dh, err := ioutil.TempDir("", "fish-home-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dh)

	home := fish.Home(dh)
	if err := os.MkdirAll(home.Rigs(), 0755); err != nil {
		t.Fatalf("Could not create %s: %s", home.Rigs(), err)
	}

	source := "https://github.com/fishworks/fish-food"
	i, err := New(source, "", home)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// ensure a VCSInstaller was returned
	_, ok := i.(*VCSInstaller)
	if !ok {
		t.Error("expected a VCSInstaller")
	}

	if err := Update(i); err == nil {
		t.Error("expected error for rig does not exist, got none")
	} else if err.Error() != "rig does not exist" {
		t.Errorf("expected error for rig does not exist, got (%v)", err)
	}

	// Install rig before update
	if err := Install(i); err != nil {
		t.Error(err)
	}

	// Test FindSource method for positive result
	rigInfo, err := FindSource(i.Path(), home)
	if err != nil {
		t.Error(err)
	}

	repoRemote := rigInfo.(*VCSInstaller).Repo.Remote()
	if repoRemote != source {
		t.Errorf("invalid source found, expected %q got %q", source, repoRemote)
	}

	// Update rig
	if err := Update(i); err != nil {
		t.Error(err)
	}

	// Test update failure
	os.Remove(filepath.Join(i.Path(), "LICENSE"))
	// Testing update for error
	if err := Update(i); err == nil {
		t.Error("expected error for rig modified, got none")
	} else if err != rig.ErrRepoDirty {
		t.Errorf("expected error for rig modified, got (%v)", err)
	}

}

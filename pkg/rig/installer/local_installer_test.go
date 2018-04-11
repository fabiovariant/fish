package installer

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/fish"
)

var _ Installer = new(LocalInstaller)

func TestLocalInstaller(t *testing.T) {
	dh, err := ioutil.TempDir("", "fish-home-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dh)

	home := fish.Home(dh)
	if err := os.MkdirAll(home.Rigs(), 0755); err != nil {
		t.Fatalf("Could not create %s: %s", home.Rigs(), err)
	}

	source := "testdata/fish-food"
	i, err := New(source, "", home)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if err := Install(i); err != nil {
		t.Error(err)
	}

	expectedPath := home.Path("Rigs", "fish-food")
	if i.Path() != expectedPath {
		t.Errorf("expected path '%s', got %q", expectedPath, i.Path())
	}
}

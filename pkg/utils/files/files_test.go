package files

import (
	"os"
	"path"
	"testing"
)

func TestFileIndexing(t *testing.T) {
	w := Create("C:\\Users\\alexw\\repos\\minimalist-shell\\bin", false, true)
	t.Log(w.Files())

	if len(w.Files()) > 0 {
		t.Fatalf("Inhuman indexing")
	}
	w.WaitForIndex()
	if len(w.Files()) <= 0 {
		t.Fatalf("Expected files.")
	}
}

func TestFileUpdates(t *testing.T) {
	dir := t.TempDir()
	w := Create(dir, false, true)
	w.WaitForIndex()
	if len(w.Files()) > 0 {
		t.Fatalf("Expected no files")
	}
	os.WriteFile(path.Join(dir, "test.exe"), []byte("Hello, exe"), os.FileMode(777))

	w.WaitForIndex()
	if len(w.Files()) != 1 {
		t.Fatal("Expected exactly one file, got", len(w.Files()))
	}
	t.Log(w.Files())
}

// Recursive watching not implemented
// func TestRecursiveFileUpdates(t *testing.T) {
// }

package commands

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestSanitiseName(t *testing.T) {
	file := "example.png"
	prefix := "file"

	assert.Equal(t, "fileExamplePng", sanitiseName(file, prefix))
}

func TestSanitiseName_Exported(t *testing.T) {
	file := "example.png"
	prefix := "Public"

	assert.Equal(t, "PublicExamplePng", sanitiseName(file, prefix))
}

func TestSanitiseName_Special(t *testing.T) {
	file := "a longer name (with-syms).png"
	prefix := "file"

	assert.Equal(t, "fileALongerNameWithSymsPng", sanitiseName(file, prefix))
}

func TestWriteResource(t *testing.T) {
	f, err := os.CreateTemp("", "*.go")
	if err != nil {
		t.Fatal("Unable to create temp file:", err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	writeHeader("test", f)
	writeResource("testdata/bundle/content.txt", "contentTxt", f)

	const header = fileHeader + "\n\npackage test\n\nimport (\n\t_ \"embed\"\n\t\"fyne.io/fyne/v2\"\n)\n\n"
	const expected = header + "//go:embed testdata/bundle/content.txt\nvar contentTxtData []byte\nvar contentTxt = &fyne.StaticResource{\n\tStaticName:    \"testdata/bundle/content.txt\",\n\tStaticContent: contentTxtData,\n}\n"

	// Seek file to start so we can read the written data.
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatal("Unable to seek temp file:", err)
	}

	content, err := io.ReadAll(f)
	if err != nil {
		t.Fatal("Unable to read temp file:", err)
	}

	assert.Equal(t, expected, string(content))
}

func TestBundleGlobDump(t *testing.T) {
	app := &cli.App{
		Name:        "fyne",
		Usage:       "execute test",
		Description: "writes the bundle.go and bundle_test.go to stdout",
		Commands: []*cli.Command{
			Bundle(),
		},
	}
	assert.NoError(t, app.Run([]string{"fyne", "bundle", "bundle*.go"}))
}

func BenchmarkWriteResource(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		f, _ := os.CreateTemp("", "*.go")
		b.StartTimer()

		writeHeader("test", f)
		writeResource("testdata/bundle/content.txt", "contentTxt", f)

		b.StopTimer()
		f.Close()
		os.Remove(f.Name())
		b.StartTimer()
	}
}

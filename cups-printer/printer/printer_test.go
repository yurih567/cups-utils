package printer_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"cups-printer/job"
	"cups-printer/printer"
)

func TestPrintSmokeToFile(t *testing.T) {
	out := filepath.Join(t.TempDir(), "job.bin")

	xml := []byte(`
<Page width="80mm" padding="8">
  <Column gap="4" align="center">
    <Text align="center" size="16" weight="bold">Smoke Test</Text>
  </Column>
</Page>
`)

	err := printer.Print(context.Background(), job.Job{
		XML:    xml,
		Model:  "epson",
		Dest:   "file:" + out,
		Beep:   1,
		Drawer: true,
		Cut:    true,
		Feed:   2,
		DPI:    203,
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 32 {
		t.Fatalf("payload too small: %d bytes", len(data))
	}
	if data[0] != 0x1b || data[1] != '@' {
		t.Fatalf("missing init prefix: %v", data[:4])
	}
}

package transport

import (
	"fmt"
	"os/exec"
)

// Cups sends raw bytes to a CUPS queue via lp -o raw.
type Cups struct {
	Queue string
}

func (c Cups) String() string {
	return "cups:" + c.Queue
}

func (c Cups) Send(payload []byte) error {
	cmd := exec.Command("lp", "-d", c.Queue, "-o", "raw")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("cups %s: stdin: %w", c.Queue, err)
	}

	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		return fmt.Errorf("cups %s: start lp: %w", c.Queue, err)
	}

	if _, err := stdin.Write(payload); err != nil {
		_ = stdin.Close()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return fmt.Errorf("cups %s: write: %w", c.Queue, err)
	}
	if err := stdin.Close(); err != nil {
		_ = cmd.Wait()
		return fmt.Errorf("cups %s: close stdin: %w", c.Queue, err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("cups %s: lp failed: %w", c.Queue, err)
	}
	return nil
}

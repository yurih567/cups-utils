package job_test

import (
	"testing"

	"cups-printer/job"
)

func TestJobDefaultsAreZeroValues(t *testing.T) {
	var j job.Job
	if j.Cut || j.Drawer || j.Beep != 0 {
		t.Fatalf("unexpected zero job: %+v", j)
	}
}

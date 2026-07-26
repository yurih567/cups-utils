package printer

import (
	"context"
	"fmt"
	"strings"

	"cups-printer/assemble"
	"cups-printer/job"
	"cups-printer/render"
	"cups-printer/transport"

	drivers "cups-drivers"
	_ "cups-drivers/models"
)

// Print renders the job XML and sends the payload to the configured destination.
func Print(ctx context.Context, j job.Job) error {
	_ = ctx

	model := strings.ToLower(strings.TrimSpace(j.Model))
	if model == "" {
		model = "generic"
	}
	drv, err := drivers.Get(model)
	if err != nil {
		return err
	}

	if len(j.XML) == 0 {
		return fmt.Errorf("printer: empty template xml")
	}
	if strings.TrimSpace(j.Dest) == "" {
		return fmt.Errorf("printer: destination is required")
	}

	feedInBody := 0
	body, err := render.Render(j.XML, render.Options{
		Driver:    drv,
		AssetsDir: j.Assets,
		DPI:       j.DPI,
		FeedLines: feedInBody,
	})
	if err != nil {
		return err
	}

	feed := max(j.Feed, 0)
	payload := assemble.Payload(body, drv, assemble.Options{
		Feed:       feed,
		Cut:        j.Cut,
		PartialCut: j.PartialCut,
		Beep:       j.Beep,
		Drawer:     j.Drawer,
		DrawerPin:  j.DrawerPin,
	})

	t, err := transport.Parse(j.Dest)
	if err != nil {
		return err
	}
	if err := t.Send(payload); err != nil {
		return fmt.Errorf("printer: send to %s: %w", t.String(), err)
	}
	return nil
}

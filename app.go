package main

import (
	"encoding/json"
	"fmt"
	"github.com/oniontree-org/go-oniontree"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"net/http"
	"unicode"
)

const Version = "0.1"

type Application struct {
	ot  *oniontree.OnionTree
	app *cli.App
}

func (a *Application) handleOnionTreeOpen() cli.BeforeFunc {
	return func(c *cli.Context) error {
		ot, err := oniontree.Open(c.String("C"))
		if err != nil {
			return fmt.Errorf("failed to open OnionTree repository: %s", err)
		}
		a.ot = ot
		return nil
	}
}

func (a *Application) handleSyncCommand() cli.ActionFunc {
	normalizeName := func(title string) string {
		return fmt.Sprintf("SecureDrop: %s", title)
	}
	normalizeID := func(name string) string {
		isMn := func(r rune) bool {
			return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
		}
		t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
		name, _, _ = transform.String(t, name)
		return fmt.Sprintf("securedrop-%s", name)
	}
	return func(c *cli.Context) error {
		client := &http.Client{
			Timeout: c.Duration("timeout"),
		}

		req, err := http.NewRequest("GET", c.String("url"), nil)
		if err != nil {
			return fmt.Errorf("failed to create new request: %s", err)
		}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to establish connection with the API: %s", err)
		}
		defer resp.Body.Close()

		// Decode JSON payload
		var arr []interface{}
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&arr); err != nil {
			return fmt.Errorf("failed to decode response data: %s", err)
		}

		for _, obj := range arr {
			instance, ok := obj.(map[string]interface{})
			if !ok {
				return fmt.Errorf("unexpected data format encountered")
			}

			id := normalizeID(instance["slug"].(string))
			service := oniontree.NewService(id)
			service.Name = normalizeName(instance["title"].(string))
			service.Description = instance["organization_description"].(string)
			service.SetURLs([]string{fmt.Sprintf("http://%s", instance["onion_address"].(string))})

			if err := a.ot.AddService(service); err != nil {
				if _, ok := err.(*oniontree.ErrIdExists); !ok {
					return fmt.Errorf("failed to add new service: %s", err)
				}
				if err := a.ot.UpdateService(service); err != nil {
					return fmt.Errorf("failed to update service: %s", err)
				}
			}

			tags := make([]oniontree.Tag, len(c.StringSlice("tag")))
			for i, tag := range c.StringSlice("tag") {
				tags[i] = oniontree.Tag(tag)
			}

			if len(tags) > 0 {
				if err := a.ot.TagService(id, tags); err != nil {
					return fmt.Errorf("failed to create new tags: %s", err)
				}
			}
		}
		return nil
	}
}

func (a *Application) Run(args []string) error {
	return a.app.Run(args)
}

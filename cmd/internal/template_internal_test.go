package internal

import (
	"testing"

	"github.com/flowexec/flow/v2/types/executable"
)

func tmplWith(name, location string) *executable.Template {
	t := &executable.Template{}
	t.SetContext(name, location)
	return t
}

func TestMergeTemplates_RegisteredWinsOnCollision(t *testing.T) {
	registered := executable.TemplateList{
		tmplWith("webapp", "/registered/webapp.flow.tmpl"),
	}
	discovered := executable.TemplateList{
		tmplWith("webapp", "/discovered/webapp.flow.tmpl"),
		tmplWith("service", "/discovered/service.flow.tmpl"),
	}

	merged := mergeTemplates(registered, discovered)

	if len(merged) != 2 {
		t.Fatalf("mergeTemplates returned %d templates, want 2", len(merged))
	}

	byName := map[string]string{}
	for _, tmpl := range merged {
		byName[tmpl.Name()] = tmpl.Location()
	}

	if got := byName["webapp"]; got != "/registered/webapp.flow.tmpl" {
		t.Errorf("registered template should win on collision; got webapp=%q", got)
	}
	if got := byName["service"]; got != "/discovered/service.flow.tmpl" {
		t.Errorf("discovered-only template should be present; got service=%q", got)
	}
}

func TestMergeTemplates_SortedByName(t *testing.T) {
	discovered := executable.TemplateList{
		tmplWith("zeta", "/z.flow.tmpl"),
		tmplWith("alpha", "/a.flow.tmpl"),
		tmplWith("mid", "/m.flow.tmpl"),
	}

	merged := mergeTemplates(nil, discovered)

	want := []string{"alpha", "mid", "zeta"}
	for i, tmpl := range merged {
		if tmpl.Name() != want[i] {
			t.Fatalf("merged[%d].Name() = %q, want %q", i, tmpl.Name(), want[i])
		}
	}
}

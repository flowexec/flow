package cache

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/flowexec/tuikit"
	"github.com/flowexec/tuikit/types"
	"github.com/flowexec/tuikit/views"
	"gopkg.in/yaml.v3"
)

type cacheData struct {
	Cache map[string]string `json:"cache" yaml:"cache"`
}

func (d *cacheData) Items() []*types.EntityInfo {
	items := make([]*types.EntityInfo, 0, len(d.Cache))
	for key, value := range d.Cache {
		items = append(items, &types.EntityInfo{
			Header:    key,
			SubHeader: value,
			ID:        key,
		})
	}
	return items
}

func (d *cacheData) Singular() string {
	return "Entry"
}

func (d *cacheData) Plural() string {
	return "Entries"
}

func (d *cacheData) YAML() (string, error) {
	data, err := yaml.Marshal(d)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (d *cacheData) JSON() (string, error) {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func NewCacheListView(
	container *tuikit.Container,
	cache map[string]string,
) tuikit.View {
	keys := make([]string, 0, len(cache))
	for k := range cache {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	columns := []views.TableColumn{
		{Title: fmt.Sprintf("Entries (%d)", len(keys)), Percentage: 40},
		{Title: "Value", Percentage: 60},
	}
	rows := make([]views.TableRow, 0, len(keys))
	for _, k := range keys {
		rows = append(rows, views.TableRow{Data: []string{k, cache[k]}})
	}
	return views.NewTable(container.RenderState(), columns, rows, views.TableDisplayMini)
}

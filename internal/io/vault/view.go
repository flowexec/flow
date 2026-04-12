package vault

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/flowexec/tuikit"
	"github.com/flowexec/tuikit/types"
	"github.com/flowexec/tuikit/views"
	extVault "github.com/flowexec/vault"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/internal/vault"
)

type vaultEntity struct {
	Name string `json:"name" yaml:"name"`
	Path string `json:"path" yaml:"path"`
	Type string `json:"type" yaml:"type"`

	Data map[string]interface{} `json:"data" yaml:"data"`
}

func (v *vaultEntity) YAML() (string, error) {
	data, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (v *vaultEntity) JSON() (string, error) {
	data, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (v *vaultEntity) Markdown() string {
	md := fmt.Sprintf(
		"# [Vault] %s\n\n**Path:** %s\n\n**Type:** %s\n\n",
		v.Name, v.Path, v.Type,
	)

	if v.Data != nil {
		md += "## Data\n\n"
		keys := maps.Keys(v.Data)
		slices.Sort(keys)
		for _, key := range keys {
			md += fmt.Sprintf("**%s:** %v\n\n", key, v.Data[key])
		}
	}

	return md
}

func NewVaultView(
	container *tuikit.Container,
	vaultName string,
) tuikit.View {
	v, err := vaultFromName(vaultName)
	if err != nil || v == nil {
		return views.NewErrorView(err, container.RenderState().Theme)
	}
	return views.NewEntityView(container.RenderState(), v, types.EntityFormatDocument)
}

type vaultCollection struct {
	Vaults []*vaultEntity `json:"vaults" yaml:"vaults"`
}

func (vc *vaultCollection) Singular() string {
	return "vault"
}

func (vc *vaultCollection) Plural() string {
	return "vaults"
}

func (vc *vaultCollection) Items() []*types.EntityInfo {
	items := make([]*types.EntityInfo, len(vc.Vaults))
	for i, v := range vc.Vaults {
		items[i] = &types.EntityInfo{
			Header:    v.Name,
			SubHeader: v.Path,
			ID:        v.Name,
		}
	}
	return items
}

func (vc *vaultCollection) YAML() (string, error) {
	data, err := yaml.Marshal(vc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal vaults: %w", err)
	}
	return string(data), nil
}

func (vc *vaultCollection) JSON() (string, error) {
	data, err := json.MarshalIndent(vc, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal vaults: %w", err)
	}
	return string(data), nil
}

func NewVaultListView(
	container *tuikit.Container,
	vaultNames []string,
) tuikit.View {
	vaults := &vaultCollection{Vaults: make([]*vaultEntity, 0, len(vaultNames))}
	for _, name := range vaultNames {
		v, err := vaultFromName(name)
		if err != nil || v == nil {
			return views.NewErrorView(
				fmt.Errorf("vault '%s' error: %w", name, err),
				container.RenderState().Theme,
			)
		}
		vaults.Vaults = append(vaults.Vaults, v)
	}
	if len(vaults.Vaults) == 0 {
		return views.NewErrorView(fmt.Errorf("no vaults found"), container.RenderState().Theme)
	}

	columns := []views.TableColumn{
		{Title: fmt.Sprintf("Vaults (%d)", len(vaults.Vaults)), Percentage: 40},
		{Title: "Path", Percentage: 60},
	}
	rows := make([]views.TableRow, 0, len(vaults.Vaults))
	for _, v := range vaults.Vaults {
		rows = append(rows, views.TableRow{Data: []string{v.Name, v.Path}})
	}
	table := views.NewTable(container.RenderState(), columns, rows, views.TableDisplayMini)
	table.SetOnSelect(func(_ int) error {
		row := table.GetSelectedRow()
		if row == nil || len(row.Data()) == 0 {
			return fmt.Errorf("no vault selected")
		}
		return container.SetView(NewVaultView(container, row.Data()[0]))
	})
	return table
}

func vaultFromName(vaultName string) (*vaultEntity, error) {
	cfg, vlt, err := vault.VaultFromName(vaultName)
	if err != nil {
		return nil, err
	}
	data := make(map[string]interface{})
	data["created"] = vlt.Metadata().Created
	data["lastModified"] = vlt.Metadata().LastModified

	v := &vaultEntity{
		Name: vlt.ID(),
		Type: string(cfg.Type),
		Data: data,
	}

	switch cfg.Type {
	case extVault.ProviderTypeAES256:
		v.Path = cfg.Aes.StoragePath
		data["sources"] = cfg.Aes.KeySource
	case extVault.ProviderTypeAge:
		v.Path = cfg.Age.StoragePath
		data["sources"] = cfg.Age.IdentitySources
		data["recipients"] = cfg.Age.Recipients
	}

	return v, nil
}

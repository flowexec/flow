package vault

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/flowexec/tuikit"
	"github.com/flowexec/tuikit/views"
	extVault "github.com/flowexec/vault"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/v2/internal/io/common"
	"github.com/flowexec/flow/v2/internal/vault"
)

type vaultEntity struct {
	Name string `json:"name" yaml:"name"`
	Path string `json:"path" yaml:"path"`
	Type string `json:"type" yaml:"type"`
	// Error is set when the vault could not be opened. It is listed anyway, so
	// one unusable vault does not hide every other one.
	Error string `json:"error,omitempty" yaml:"error,omitempty"`

	Data map[string]interface{} `json:"data" yaml:"data"`
}

// unopenableVault is the placeholder for a vault whose config loads but whose
// provider will not open -- a missing key, a revoked identity, a config written
// by an older version.
func unopenableVault(name string, err error) *vaultEntity {
	return &vaultEntity{Name: name, Type: "unknown", Error: err.Error(), Data: map[string]interface{}{}}
}

func (v *vaultEntity) YAML() (string, error) {
	data, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (v *vaultEntity) JSON() (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
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

	var footer string
	if v.Path != "" {
		footer = fmt.Sprintf("_Located in %s_", common.ShortenPath(v.Path))
	}

	opts := views.DetailContentOpts{
		Title:    v.Name,
		Subtitle: "Vault",
		Metadata: []views.DetailField{
			{Key: "Type", Value: v.Type},
		},
		Body:   vaultBodyMarkdown(v),
		Footer: footer,
		Entity: v,
	}

	return views.NewDetailContentView(container.RenderState(), opts)
}

type vaultCollection struct {
	Vaults []*vaultEntity `json:"vaults" yaml:"vaults"`
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
	vaults := make([]*vaultEntity, 0, len(vaultNames))
	for _, name := range vaultNames {
		v, err := vaultFromName(name)
		if err != nil || v == nil {
			v = unopenableVault(name, err)
		}
		vaults = append(vaults, v)
	}
	if len(vaults) == 0 {
		return views.NewErrorView(fmt.Errorf("no vaults found"), container.RenderState().Theme)
	}

	sort.Slice(vaults, func(i, j int) bool {
		return vaults[i].Name < vaults[j].Name
	})

	columns := []views.TableColumn{
		{Title: fmt.Sprintf("Vaults (%d)", len(vaults)), Percentage: 35},
		{Title: "Type", Percentage: 25},
		{Title: "Path", Percentage: 40},
	}
	rows := make([]views.TableRow, 0, len(vaults))
	for _, v := range vaults {
		rows = append(rows, views.TableRow{
			Data: []string{v.Name, v.Type, common.ShortenPath(v.Path)},
		})
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

func vaultBodyMarkdown(v *vaultEntity) string {
	if v.Data == nil {
		return ""
	}

	var sections []string

	if created, ok := v.Data["created"]; ok {
		sections = append(sections, fmt.Sprintf("**Created:** %v", created))
	}
	if modified, ok := v.Data["lastModified"]; ok {
		sections = append(sections, fmt.Sprintf("**Last Modified:** %v", modified))
	}

	if sources, ok := v.Data["sources"]; ok {
		sections = append(sections, formatSourceList("Sources", sources))
	}
	if recipients, ok := v.Data["recipients"]; ok {
		sections = append(sections, formatSourceList("Recipients", recipients))
	}

	return strings.Join(sections, "\n\n")
}

// formatSourceList formats KeySource, IdentitySource slices, or plain strings
// into a readable markdown list.
func formatSourceList(label string, value any) string {
	switch v := value.(type) {
	case []extVault.KeySource:
		if len(v) == 0 {
			return ""
		}
		md := fmt.Sprintf("**%s**\n", label)
		for _, src := range v {
			md += fmt.Sprintf("- `%s`", src.Type)
			if src.Path != "" {
				md += fmt.Sprintf(" — %s", src.Path)
			} else if src.Name != "" {
				md += fmt.Sprintf(" — $%s", src.Name)
			}
			md += "\n"
		}
		return md
	case []extVault.IdentitySource:
		if len(v) == 0 {
			return ""
		}
		md := fmt.Sprintf("**%s**\n", label)
		for _, src := range v {
			md += fmt.Sprintf("- `%s`", src.Type)
			if src.Path != "" {
				md += fmt.Sprintf(" — %s", src.Path)
			} else if src.Name != "" {
				md += fmt.Sprintf(" — $%s", src.Name)
			}
			md += "\n"
		}
		return md
	case string:
		return fmt.Sprintf("**%s:** `%s`", label, v)
	default:
		return fmt.Sprintf("**%s:** %v", label, v)
	}
}

func vaultFromName(vaultName string) (*vaultEntity, error) {
	cfg, vlt, err := vault.VaultFromName(vaultName)
	if err != nil {
		return nil, err
	}
	data := make(map[string]interface{})
	// A metadata failure must not make the vault unviewable. For an external
	// vault it means the backend CLI is missing or the session has expired,
	// which is exactly when a user wants to look at the vault's configuration.
	// Report it as a field instead of failing the whole view.
	if metadata, mdErr := vlt.Metadata(); mdErr != nil {
		data["metadataError"] = mdErr.Error()
	} else {
		data["created"] = metadata.Created
		data["lastModified"] = metadata.LastModified
	}

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
	case extVault.ProviderTypeExternal:
		v.Path = cfg.External.StoragePath
		// An external vault reads through to another tool, so what is worth
		// showing is how many links it holds and that it stores nothing itself.
		data["readOnly"] = true
		if links, ok := vault.AsReferenceVault(vlt); ok {
			if all, linkErr := links.Links(); linkErr == nil {
				data["links"] = len(all)
			}
		}
		if inert := cfg.External.LegacyWriteCommands(); len(inert) > 0 {
			data["inertCommands"] = inert
		}
		// Which generator produced this config, and with what values. A UI needs
		// both to browse the provider again later without asking the user to
		// restate its region or 1Password vault name.
		if src := cfg.External.Source; src != nil {
			data["source"] = src
		}
	}

	return v, nil
}

package config_test

import (
	"encoding/json"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/types/config"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("Workspace path lookups", func() {
	Describe("WorkspaceForPath", func() {
		It("returns the workspace containing the path", func() {
			cfg := &config.Config{Workspaces: map[string]string{"a": "/src/a", "b": "/src/b"}}
			name, found := cfg.WorkspaceForPath("/src/a/pkg/deep")
			Expect(found).To(BeTrue())
			Expect(name).To(Equal("a"))
		})

		It("prefers the longest match when workspaces nest", func() {
			// Map iteration order used to decide this, so the nested workspace won only by luck.
			cfg := &config.Config{Workspaces: map[string]string{
				"outer": "/src/repo",
				"inner": "/src/repo/sub/worktree",
			}}
			for range 20 {
				name, found := cfg.WorkspaceForPath("/src/repo/sub/worktree/cmd")
				Expect(found).To(BeTrue())
				Expect(name).To(Equal("inner"))
			}
		})

		It("does not match a sibling sharing a name prefix", func() {
			cfg := &config.Config{Workspaces: map[string]string{"ws": "/src/ws"}}
			_, found := cfg.WorkspaceForPath("/src/wsX")
			Expect(found).To(BeFalse())
		})

		It("matches the workspace root itself", func() {
			cfg := &config.Config{Workspaces: map[string]string{"ws": "/src/ws"}}
			name, found := cfg.WorkspaceForPath("/src/ws")
			Expect(found).To(BeTrue())
			Expect(name).To(Equal("ws"))
		})

		It("returns false for an empty path or an empty config", func() {
			cfg := &config.Config{Workspaces: map[string]string{"ws": "/src/ws"}}
			_, found := cfg.WorkspaceForPath("")
			Expect(found).To(BeFalse())
			_, found = (&config.Config{}).WorkspaceForPath("/src/ws")
			Expect(found).To(BeFalse())
		})
	})

	Describe("NameForWorkspacePath", func() {
		It("matches an exact root only", func() {
			cfg := &config.Config{Workspaces: map[string]string{"ws": "/src/ws"}}
			name, found := cfg.NameForWorkspacePath("/src/ws")
			Expect(found).To(BeTrue())
			Expect(name).To(Equal("ws"))

			_, found = cfg.NameForWorkspacePath("/src/ws/sub")
			Expect(found).To(BeFalse())
		})

		It("tolerates unclean paths", func() {
			cfg := &config.Config{Workspaces: map[string]string{"ws": "/src/ws/"}}
			_, found := cfg.NameForWorkspacePath("/src/ws/./")
			Expect(found).To(BeTrue())
		})
	})

	Describe("CurrentWorkspaceName", func() {
		It("falls back to the configured workspace in fixed mode", func() {
			cfg := &config.Config{
				Workspaces:       map[string]string{"pinned": "/src/pinned"},
				CurrentWorkspace: "pinned",
				WorkspaceMode:    config.ConfigWorkspaceModeFixed,
			}
			name, err := cfg.CurrentWorkspaceName()
			Expect(err).NotTo(HaveOccurred())
			Expect(name).To(Equal("pinned"))
		})

		It("errors when nothing is configured", func() {
			cfg := &config.Config{WorkspaceMode: config.ConfigWorkspaceModeFixed}
			_, err := cfg.CurrentWorkspaceName()
			Expect(err).To(MatchError(ContainSubstring("current workspace not found")))
		})
	})

	Describe("SetDefaults", func() {
		It("picks the first workspace by sort order, not map order", func() {
			for range 20 {
				cfg := &config.Config{Workspaces: map[string]string{
					"zulu": "/z", "alpha": "/a", "mike": "/m",
				}}
				cfg.SetDefaults()
				Expect(cfg.CurrentWorkspace).To(Equal("alpha"))
			}
		})
	})
})

var _ = Describe("MouseEnabled", func() {
	It("is on when unset", func() {
		Expect((&config.Config{}).MouseEnabled()).To(BeTrue())
		Expect((&config.Config{Interactive: &config.Interactive{Enabled: true}}).MouseEnabled()).To(BeTrue())
	})

	It("follows the interactive.mouse setting", func() {
		on, off := true, false
		Expect((&config.Config{Interactive: &config.Interactive{Mouse: &on}}).MouseEnabled()).To(BeTrue())
		Expect((&config.Config{Interactive: &config.Interactive{Mouse: &off}}).MouseEnabled()).To(BeFalse())
	})
})

var _ = Describe("Validate custom verbs", func() {
	It("accepts valid custom verbs", func() {
		cfg := &config.Config{CustomVerbs: []string{"status", "health"}}
		Expect(cfg.Validate()).To(Succeed())
	})

	It("rejects an invalid custom verb", func() {
		cfg := &config.Config{CustomVerbs: []string{"Status"}}
		Expect(cfg.Validate()).To(MatchError(ContainSubstring("invalid custom verb")))
	})

	It("rejects a custom verb that duplicates a built-in", func() {
		cfg := &config.Config{CustomVerbs: []string{"build"}}
		Expect(cfg.Validate()).To(MatchError(ContainSubstring("already a built-in verb")))
	})

	It("rejects duplicate custom verbs", func() {
		cfg := &config.Config{CustomVerbs: []string{"status", "status"}}
		Expect(cfg.Validate()).To(MatchError(ContainSubstring("duplicate custom verb")))
	})
})

var _ = Describe("LogRetentionPolicy", func() {
	It("uses the defaults when unset", func() {
		policy := (&config.Config{}).LogRetentionPolicy()
		Expect(policy.MaxFiles).To(Equal(config.DefaultLogRetentionMaxCount))
		Expect(policy.MaxAge).To(Equal(config.DefaultLogRetentionMaxAge))
		Expect(policy.MaxTotalSize).To(Equal(int64(config.DefaultLogRetentionMaxSizeMB) * 1024 * 1024))
	})

	It("overrides only the fields that are set, keeping 0 as unlimited", func() {
		maxAge, maxCount := 24*time.Hour, 0
		cfg := &config.Config{LogRetention: &config.LogRetention{MaxAge: &maxAge, MaxCount: &maxCount}}
		policy := cfg.LogRetentionPolicy()
		Expect(policy.MaxAge).To(Equal(24 * time.Hour))
		Expect(policy.MaxFiles).To(BeZero())
		Expect(policy.MaxTotalSize).To(Equal(int64(config.DefaultLogRetentionMaxSizeMB) * 1024 * 1024))
		Expect(cfg.LogRetentionSummary()).To(Equal("max age 24h0m0s, max size 500 MB, max count unlimited"))
	})

	It("rejects negative limits", func() {
		size := -1
		cfg := &config.Config{LogRetention: &config.LogRetention{MaxSizeMB: &size}}
		Expect(cfg.Validate()).To(MatchError(ContainSubstring("maxSizeMB must not be negative")))
	})
})

var _ = Describe("Config JSON", func() {
	It("writes durations as the strings the config file and schema use", func() {
		maxAge, noAge := 168*time.Hour, time.Duration(0)
		cfg := config.Config{DefaultTimeout: 5 * time.Minute, LogRetention: &config.LogRetention{MaxAge: &maxAge}}
		out, err := json.Marshal(&cfg)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(out)).To(ContainSubstring(`"defaultTimeout":"5m0s"`))
		Expect(string(out)).To(ContainSubstring(`"maxAge":"168h0m0s"`))

		cfg.LogRetention.MaxAge = &noAge
		out, err = json.Marshal(cfg)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(out)).To(ContainSubstring(`"maxAge":"0s"`))
	})

	It("leaves unset durations out", func() {
		out, err := json.Marshal(config.Config{})
		Expect(err).NotTo(HaveOccurred())
		Expect(string(out)).NotTo(ContainSubstring("defaultTimeout"))
		Expect(string(out)).NotTo(ContainSubstring("logRetention"))
	})
})

var _ = Describe("Container config", func() {
	It("runs each runtime under its own name until configured", func() {
		var cfg *config.Config
		Expect(cfg.ContainerRuntimePreference()).To(Equal(config.ContainerRuntimeAuto))
		cmd, args := cfg.ContainerCommand(config.ContainerRuntimePodman)
		Expect(cmd).To(Equal("podman"))
		Expect(args).To(BeEmpty())
	})

	It("stores a replacement command and flags, and removes them when set back", func() {
		cfg := &config.Config{}
		Expect(cfg.SetContainerCommand(config.ContainerRuntimeDocker, "finch", nil)).To(Succeed())
		Expect(cfg.SetContainerCommand(config.ContainerRuntimePodman, "podman", []string{"--connection", "vm"})).
			To(Succeed())

		cmd, _ := cfg.ContainerCommand(config.ContainerRuntimeDocker)
		Expect(cmd).To(Equal("finch"))
		cmd, args := cfg.ContainerCommand(config.ContainerRuntimePodman)
		Expect(cmd).To(Equal("podman"))
		Expect(args).To(Equal([]string{"--connection", "vm"}))
		Expect(cfg.Container.Runtimes.Podman.Command).To(BeEmpty(), "the family's own name isn't stored")

		Expect(cfg.SetContainerCommand(config.ContainerRuntimeDocker, "docker", nil)).To(Succeed())
		Expect(cfg.SetContainerCommand(config.ContainerRuntimePodman, "podman", nil)).To(Succeed())
		Expect(cfg.Container.Runtimes).To(BeNil())
	})

	It("rejects an unknown runtime, an empty command and an empty flag", func() {
		cfg := &config.Config{}
		Expect(cfg.SetContainerCommand("containerd", "nerdctl", nil)).To(HaveOccurred())
		Expect(cfg.SetContainerCommand(config.ContainerRuntimeDocker, " ", nil)).To(HaveOccurred())

		cfg.Container = &config.Container{Runtime: "lxc"}
		Expect(cfg.Validate()).To(MatchError(ContainSubstring("invalid container runtime")))

		cfg.Container = &config.Container{Runtimes: &config.ContainerRuntimes{
			Docker: &config.ContainerRuntimeCommand{Args: []string{""}},
		}}
		Expect(cfg.Validate()).To(MatchError(ContainSubstring("empty flag")))
	})
})

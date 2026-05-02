//go:build e2e

package tests_test

import (
	stdCtx "context"
	"encoding/json"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/tests/utils"
)

var _ = Describe("vault/secrets e2e", Ordered, func() {
	var (
		ctx *utils.Context
		run *utils.CommandRunner
	)

	BeforeAll(func() {
		ctx = utils.NewContext(stdCtx.Background(), GinkgoTB())
		run = utils.NewE2ECommandRunner()
	})

	BeforeEach(func() {
		utils.ResetTestContext(ctx, GinkgoTB())
	})

	AfterEach(func() {
		ctx.Finalize()
	})

	When("creating a new vault (flow vault create)", func() {
		It("should return the generated key", func() {
			stdOut := ctx.StdOut()
			keyEnv := "FLOW_TEST_VAULT_KEY"
			Expect(run.Run(ctx.Context, "vault", "create", "test", "--key-env", keyEnv, "--output", "json")).
				To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())

			var envelope struct {
				Result struct {
					Data struct {
						GeneratedKey string `json:"generatedKey"`
					} `json:"data"`
				} `json:"result"`
			}
			Expect(json.Unmarshal([]byte(out), &envelope)).To(Succeed())
			Expect(envelope.Result.Data.GeneratedKey).NotTo(BeEmpty())
			Expect(os.Setenv(keyEnv, envelope.Result.Data.GeneratedKey)).To(Succeed())
		})

		It("should create vault with custom path", func() {
			stdOut := ctx.StdOut()
			tmpdir, err := os.MkdirTemp("", "flow-vault-test")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(tmpdir)

			Expect(run.Run(
				ctx.Context, "vault", "create", "test2", "--type", "aes256", "--path", tmpdir, "--output", "json",
			)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())

			var envelope struct {
				Result struct {
					Message string `json:"message"`
					Data    struct {
						Name string `json:"name"`
						Type string `json:"type"`
					} `json:"data"`
				} `json:"result"`
			}
			Expect(json.Unmarshal([]byte(out), &envelope)).To(Succeed())
			Expect(envelope.Result.Message).To(ContainSubstring("test2"))
			Expect(envelope.Result.Data.Type).To(Equal("aes256"))
		})
	})

	It("Should remove the created vault", func() {
		reader, writer, err := os.Pipe()
		Expect(err).NotTo(HaveOccurred())
		_, err = writer.Write([]byte("yes\n"))
		Expect(err).ToNot(HaveOccurred())

		ctx.SetIO(reader, ctx.StdOut())
		Expect(run.Run(ctx.Context, "vault", "remove", "test2")).To(Succeed())
		out, err := readFileContent(ctx.StdOut())
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("Vault 'test2' deleted"))
	})

	When("switching vaults (flow vault switch)", func() {
		It("should switch to demo vault successfully", func() {
			Expect(run.Run(ctx.Context, "vault", "switch", "demo")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Vault set to demo"))
		})
	})

	When("getting vault information (flow vault get)", func() {
		It("should get demo vault in YAML format", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "vault", "get", "demo")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("name: demo"))
			Expect(out).To(ContainSubstring("type: demo"))
		})
	})

	When("getting a vault that does not exist (flow vault get doesnotexist)", func() {
		It("fatals with a failure to get vault message", func() {
			ctx.ExpectFailure()
			err := run.Run(ctx.Context, "vault", "get", "doesnotexist")
			Expect(err).To(HaveOccurred())
			Expect(ctx.ExitCalls()).To(ContainElement(ContainSubstring("Failed to get vault doesnotexist")))
		})
	})

	When("setting a secret with both --file and a value (flow secret set)", func() {
		It("fatals with a mutually exclusive input error", func() {
			ctx.ExpectFailure()
			err := run.Run(ctx.Context, "secret", "set", "message", "inline-value", "--file", "doesnotexist")
			Expect(err).To(HaveOccurred())
			Expect(ctx.ExitCalls()).To(ContainElement(ContainSubstring("either a filename OR a value")))
		})
	})

	When("listing vaults (flow vault list)", func() {
		It("should list vaults in YAML format", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "vault", "list")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("vaults:"))
		})
	})
	When("setting a secret (flow secret set)", func() {
		// NOTE: these tests are using the demo vault so values aren't actually set in the vault
		It("should save into the vault", func() {
			Expect(run.Run(ctx.Context, "secret", "set", "message", "my-value")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Secret message set in vault"))
		})

		It("should read from file and save into the vault with the --file flag", func() {
			dir := GinkgoTB().TempDir()
			GinkgoTB().Setenv("SECRET_DIR", dir)
			err := os.WriteFile(filepath.Join(dir, "secret.txt"), []byte("file data"), 0755)
			Expect(err).NotTo(HaveOccurred())
			Expect(run.Run(ctx.Context, "secret", "set", "message", "--file=$SECRET_DIR/secret.txt")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Secret message set in vault"))
		})
	})

	When("getting a secret (flow secret get)", func() {
		It("should return the secret value", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "secret", "get", "message", "--plaintext")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Thanks for trying flow!"))
		})
	})

	When("listing secrets (flow secret list)", func() {
		It("should return the list of secrets", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "secret", "list")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("message"))
		})
	})

	When("deleting a secret (flow secret remove)", func() {
		It("should remove the secret from the vault", func() {
			reader, writer, err := os.Pipe()
			Expect(err).NotTo(HaveOccurred())
			_, err = writer.Write([]byte("yes\n"))
			Expect(err).ToNot(HaveOccurred())

			ctx.SetIO(reader, ctx.StdOut())
			Expect(run.Run(ctx.Context, "secret", "remove", "message")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Secret 'message' deleted from vault"))
		})
	})
})

//go:build e2e

package tests_test

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/flowexec/tuikit"
	"github.com/muesli/termenv"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "End-to-end Test Suite")
}

const PrintEnvVar = "PRINT_TEST_STDOUT"

func readFileContent(f *os.File) (string, error) {
	out, err := os.ReadFile(f.Name())
	if err != nil {
		return "", err
	}
	outStr := string(out)
	if truthy, _ := strconv.ParseBool(os.Getenv(PrintEnvVar)); truthy {
		fmt.Println(outStr)
	}
	return outStr, nil
}

func newTUIContainer(ctx context.Context) *tuikit.Container {
	app := &tuikit.Application{Name: "flow-test"}
	container, err := tuikit.NewContainer(ctx, app)
	Expect(err).NotTo(HaveOccurred())
	return container
}

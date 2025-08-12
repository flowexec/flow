package main

import (
	stdCtx "context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra/doc"

	"github.com/flowexec/flow/cmd"
	"github.com/flowexec/flow/internal/context"
)

const (
	DocsDir = "docs"
	cliDir  = "cli"
)

var (
	oldCliRoot = filepath.Join(DocsDir, cliDir, "flow.md")
	newCliRoot = filepath.Join(DocsDir, cliDir, "README.md")
)

func main() {
	fmt.Println("generating CLI docs...")
	bkgCtx, cancelFunc := stdCtx.WithCancel(stdCtx.Background())
	ctx := context.NewContext(bkgCtx, cancelFunc, os.Stdin, os.Stdout)
	defer ctx.Finalize()

	rootCmd := cmd.NewRootCmd(ctx)
	cmd.RegisterSubCommands(ctx, rootCmd)
	rootCmd.DisableAutoGenTag = true
	if err := doc.GenMarkdownTree(rootCmd, filepath.Join(rootDir(), DocsDir, cliDir)); err != nil {
		panic(err)
	}
	if err := os.Rename(oldCliRoot, newCliRoot); err != nil {
		panic(err)
	}

	fmt.Println("generating markdown docs...")
	generateMarkdownDocs()

	fmt.Println("generating schema files...")
	generateJSONSchemas()
}

func rootDir() string {
	_, filename, _, _ := runtime.Caller(0)
	// ./tools/docsgen/schema.go -> ./
	return filepath.Dir(filepath.Dir(filepath.Dir(filepath.Base(filename))))
}

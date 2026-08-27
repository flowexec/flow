package main

import (
	stdCtx "context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra/doc"

	"github.com/flowexec/flow/v2/pkg/cli"
	"github.com/flowexec/flow/v2/pkg/context"
)

const (
	DocsDir = "docs"
	cliDir  = "cli"
)

func main() {
	fmt.Println("generating CLI docs...")
	bkgCtx, cancelFunc := stdCtx.WithCancel(stdCtx.Background())
	ctx := context.NewContext(bkgCtx, cancelFunc, context.WithStdIn(os.Stdin), context.WithStdOut(os.Stdout))
	defer ctx.Finalize()

	rootCmd := cli.BuildRootCommand(ctx)
	cli.RegisterAllCommands(ctx, rootCmd)
	rootCmd.DisableAutoGenTag = true
	cliOut := filepath.Join(rootDir(), DocsDir, cliDir)
	if err := doc.GenMarkdownTree(rootCmd, cliOut); err != nil {
		panic(err)
	}
	if err := polishCLIDocs(cliOut); err != nil {
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

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	kpu "github.com/Diaphteiros/kpu/cmd"
	"github.com/spf13/cobra/doc"
	"sigs.k8s.io/yaml"
)

func main() {
	if len(os.Args) < 2 {
		panic("documentation folder path required as argument")
	}
	if err := doc.GenMarkdownTree(kpu.RootCmd, os.Args[1]); err != nil {
		panic(err)
	}

	if err := handleReplacements(); err != nil {
		panic(fmt.Errorf("error handling replacements: %w", err))
	}
}

func handleReplacements() error {
	_, mainfilepath, _, _ := runtime.Caller(0)
	curPath := filepath.Dir(mainfilepath)
	raw, err := os.ReadFile(filepath.Join(curPath, "replace.yaml"))
	if err != nil {
		if os.IsNotExist(err) {
			// nothing to do
			fmt.Println("No replace.yaml found, skipping replacement handling.")
			return nil
		}
		return err
	}

	config := &ReplaceCfg{}
	if err := yaml.Unmarshal(raw, config); err != nil {
		return err
	}

	docsPath := filepath.Join(curPath, "..", "..", "docs")
	for filename, replacements := range config.Replacements {
		path := filepath.Join(docsPath, filename)
		fmt.Printf("Processing replacements for %s ...\n", path)
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", path, err)
		}
		for _, rep := range replacements {
			content = bytes.ReplaceAll(content, []byte(rep.Replace), []byte(rep.With))
		}
		if err := os.WriteFile(path, content, os.ModePerm); err != nil {
			return fmt.Errorf("error writing file %s: %w", path, err)
		}
	}

	return nil
}

type ReplaceCfg struct {
	Replacements map[string][]Replacement `json:"replacements"`
}

type Replacement struct {
	Replace string `json:"replace"`
	With    string `json:"with"`
}

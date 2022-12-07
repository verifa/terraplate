package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra/doc"
	"github.com/verifa/terraplate/cmd"
)

func main() {
	genCLIDocs()
}

func genCLIDocs() {
	err := doc.GenMarkdownTreeCustom(
		cmd.RootCmd,
		"./docs/commands",
		func(s string) string {
			filename := filepath.Base(s)
			name := filename[:len(filename)-len(filepath.Ext(filename))]
			name = strings.Join(strings.Split(name, "_"), " ")
			return fmt.Sprintf(`---
# # AUTOMATICALLY GENERATED BY COBRA (DO NOT EDIT)
title: "%s"
---
`, name)
		},
		func(s string) string {
			return s
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

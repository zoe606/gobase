// Command init creates a project with one HTTP engine.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"go-boilerplate/pkg/project"
	"go-boilerplate/pkg/scaffold"
)

func main() {
	engine := flag.String("engine", "gin", "HTTP engine: gin, stdlib, or fiber")
	profile := flag.String("profile", "full", "project profile: full or minimal")
	output := flag.String("output", "", "new project directory")
	module := flag.String("module", "", "new Go module name")
	name := flag.String("app-name", "", "application name")
	source := flag.String("source", ".", "gobase template checkout")
	tidy := flag.Bool("tidy", true, "resolve dependencies after generation")
	flag.Parse()
	if *output == "" {
		fmt.Fprintln(os.Stderr, "-output is required")
		os.Exit(1)
	}
	if err := scaffold.Generate(scaffold.Config{Source: *source, Output: *output, Module: *module, Name: *name, Engine: project.Engine(*engine), Profile: project.Profile(*profile)}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *tidy {
		command := exec.CommandContext(context.Background(), "go", "mod", "tidy")
		command.Dir = *output
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		if err := command.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Project generated; dependency resolution failed:", err)
			os.Exit(1)
		}
	}
	fmt.Printf("Created %s using %s (%s profile).\n", *output, *engine, *profile)
}

// Command ag is the AGS3D project build tool.
//
// Usage:
//
//	ag build                     # parse changed .agscript files, emit GDScript
//	ag run                       # build + launch Godot editor
//	ag validate                  # static analysis (broken refs, unset flags)
//	ag export --platform <name>  # build + Godot export pipeline
//	ag new <name>                # scaffold a new project
//	ag viz tokens <file>         # print token stream (VIZ-01)
//	ag viz ast <file>            # print AST tree (VIZ-02)
//	ag viz blocking <file>       # print blocking call annotations (VIZ-03)
//	ag viz emit <file>           # print side-by-side AGS-spirit ↔ GDScript (VIZ-04)
//	ag viz <file>                # run all viz stages
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ags3d/ag/internal/emitter"
	"github.com/ags3d/ag/internal/parser"
	"github.com/ags3d/ag/internal/project"
	"github.com/ags3d/ag/internal/scanner"
	"github.com/ags3d/ag/internal/viz"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	var err error
	switch os.Args[1] {
	case "build":
		err = cmdBuild(os.Args[2:])
	case "run":
		err = cmdRun(os.Args[2:])
	case "validate":
		err = cmdValidate(os.Args[2:])
	case "export":
		err = cmdExport(os.Args[2:])
	case "new":
		err = cmdNew(os.Args[2:])
	case "viz":
		err = cmdViz(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "ag: unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ag %s: %v\n", os.Args[1], err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `Usage: ag COMMAND [args]

Commands:
  build                    parse changed source files, emit GDScript
  run                      build then launch Godot editor
  validate                 static analysis on the project
  export --platform NAME   build and export (windows|mac|linux|web|ios|android)
  new NAME                 scaffold a new AGS3D project
  viz tokens  FILE         print token stream (line/col/kind/lexeme)
  viz ast     FILE         print AST tree
  viz blocking FILE        print blocking call annotations
  viz emit    FILE         print side-by-side AGS-spirit ↔ GDScript
  viz         FILE         run all viz stages`)
}

// -------------------------------------------------------------------
// ag viz
// -------------------------------------------------------------------

func cmdViz(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: ag viz [tokens|ast|blocking|emit] <file>")
	}

	// Determine sub-command and file.
	// Accept:  ag viz <file>          — run all stages
	//          ag viz tokens  <file>  — single stage
	stage := "all"
	file := args[0]
	if len(args) == 2 {
		stage = args[0]
		file = args[1]
	} else if len(args) > 2 {
		return fmt.Errorf("usage: ag viz [tokens|ast|blocking|emit] <file>")
	}

	src, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	content := string(src)

	switch stage {
	case "tokens":
		viz.Tokens(os.Stdout, file, content)
	case "ast":
		viz.AST(os.Stdout, file, content)
	case "blocking":
		viz.Blocking(os.Stdout, file, content)
	case "emit":
		viz.Emit(os.Stdout, file, content)
	case "all":
		viz.Tokens(os.Stdout, file, content)
		fmt.Fprintln(os.Stdout)
		viz.AST(os.Stdout, file, content)
		fmt.Fprintln(os.Stdout)
		viz.Blocking(os.Stdout, file, content)
		fmt.Fprintln(os.Stdout)
		viz.Emit(os.Stdout, file, content)
	default:
		return fmt.Errorf("unknown viz stage %q — expected tokens, ast, blocking, or emit", stage)
	}
	return nil
}

// requireProject finds the project root or exits with an error.
func requireProject() (string, *project.Manifest) {
	root, ok := project.Find(".")
	if !ok {
		fmt.Fprintln(os.Stderr, "ag: no game.agp found in current directory or any parent")
		os.Exit(1)
	}
	m, err := project.Load(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ag:", err)
		os.Exit(1)
	}
	return root, m
}

// -------------------------------------------------------------------
// ag build
// -------------------------------------------------------------------

func cmdBuild(_ []string) error {
	root, _ := requireProject()
	return build(root)
}

func build(root string) error {
	files, err := project.Scan(root)
	if err != nil {
		return err
	}
	manifest, err := project.LoadManifest(root)
	if err != nil {
		return err
	}
	changed := project.Changed(files, manifest)
	if len(changed) == 0 {
		fmt.Println("ag build: nothing to do (no changed source files)")
		return nil
	}

	fmt.Printf("ag build: %d file(s) to process\n", len(changed))

	generatedDir := filepath.Join(root, ".engine", "generated")
	if err := os.MkdirAll(generatedDir, 0755); err != nil {
		return err
	}

	em := emitter.New()
	var errs []error

	for _, src := range changed {
		data, err := os.ReadFile(src.Path)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		s := scanner.New(src.Rel, string(data))
		p := parser.New(s)
		f, parseErrs := p.Parse(src.Rel)
		for _, pe := range parseErrs {
			fmt.Fprintln(os.Stderr, pe)
			errs = append(errs, pe)
		}
		if len(parseErrs) > 0 {
			continue
		}
		result, err := em.Emit(f)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: emit: %w", src.Rel, err))
			continue
		}

		// Write .gd output
		outPath := filepath.Join(generatedDir, src.Rel+".gd")
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			errs = append(errs, err)
			continue
		}
		if err := os.WriteFile(outPath, []byte(result.GDScript), 0644); err != nil {
			errs = append(errs, err)
			continue
		}

		project.RecordMtimes([]project.SourceFile{src}, manifest)
		fmt.Printf("  %s → %s\n", src.Rel, filepath.Join(".engine/generated", src.Rel+".gd"))
	}

	if saveErr := project.SaveManifest(root, manifest); saveErr != nil {
		errs = append(errs, saveErr)
	}

	if len(errs) > 0 {
		return fmt.Errorf("%d error(s) during build", len(errs))
	}
	fmt.Println("ag build: done")
	return nil
}

// -------------------------------------------------------------------
// ag run
// -------------------------------------------------------------------

func cmdRun(_ []string) error {
	root, _ := requireProject()
	if err := build(root); err != nil {
		return err
	}
	godot, err := findGodot()
	if err != nil {
		return err
	}
	engineDir := filepath.Join(root, ".engine")
	if err := os.MkdirAll(engineDir, 0755); err != nil {
		return err
	}
	projectFile := filepath.Join(engineDir, "project.godot")
	if _, err := os.Stat(projectFile); errors.Is(err, os.ErrNotExist) {
		stub := "; Engine configuration file.\n[application]\nconfig/name=\"AGS3D Game\"\n"
		if err := os.WriteFile(projectFile, []byte(stub), 0644); err != nil {
			return err
		}
	}
	fmt.Printf("ag run: launching %s --editor --path %s\n", godot, engineDir)
	cmd := exec.Command(godot, "--editor", "--path", engineDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func findGodot() (string, error) {
	if env := os.Getenv("GODOT"); env != "" {
		if path, err := exec.LookPath(env); err == nil {
			return path, nil
		}
	}
	for _, name := range []string{"godot", "godot4", "Godot"} {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	// Fall back to repo-local build artefact
	exe, _ := os.Executable()
	repoRoot := filepath.Join(filepath.Dir(exe), "..", "..", "..")
	matches, _ := filepath.Glob(filepath.Join(repoRoot, "bin", "godot.linuxbsd.editor.*"))
	if len(matches) > 0 {
		return matches[0], nil
	}
	return "", fmt.Errorf("Godot binary not found — set the GODOT environment variable or ensure 'godot' is on PATH")
}

// -------------------------------------------------------------------
// ag validate
// -------------------------------------------------------------------

func cmdValidate(_ []string) error {
	// TODO(T12): run analysis.Analyze over all parsed files, print diagnostics.
	fmt.Println("ag validate: static analysis not yet implemented (T12+)")
	return nil
}

// -------------------------------------------------------------------
// ag export
// -------------------------------------------------------------------

func cmdExport(args []string) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	platform := fs.String("platform", "", "export target (windows|mac|linux|web|ios|android)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *platform == "" {
		return fmt.Errorf("--platform is required")
	}
	// TODO(T18): build then invoke Godot's export pipeline.
	fmt.Printf("ag export: export pipeline for %q not yet implemented (T18)\n", *platform)
	return nil
}

// -------------------------------------------------------------------
// ag new
// -------------------------------------------------------------------

func cmdNew(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: ag new <name>")
	}
	name := args[0]
	dest, err := filepath.Abs(name)
	if err != nil {
		return err
	}
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("%q already exists", dest)
	}
	if err := project.Scaffold(dest, name); err != nil {
		return err
	}
	fmt.Printf("ag new: created %q — run 'cd %s && ag run' to get started\n", name, name)
	return nil
}

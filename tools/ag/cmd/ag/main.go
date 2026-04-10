// Command ag is the AGS3D project build tool.
//
// Usage:
//
//	ag build [--force]           # parse changed .agscript files, emit GDScript
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
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ags3d/ag/internal/aganim"
	"github.com/ags3d/ag/internal/char"
	"github.com/ags3d/ag/internal/cut"
	"github.com/ags3d/ag/internal/dlg"
	"github.com/ags3d/ag/internal/emitter"
	"github.com/ags3d/ag/internal/gui"
	"github.com/ags3d/ag/internal/item"
	"github.com/ags3d/ag/internal/loc"
	"github.com/ags3d/ag/internal/loc/tui"
	"github.com/ags3d/ag/internal/parser"
	"github.com/ags3d/ag/internal/project"
	"github.com/ags3d/ag/internal/room"
	"github.com/ags3d/ag/internal/scanner"
	"github.com/ags3d/ag/internal/scene"
	"github.com/ags3d/ag/internal/validate"
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
	case "ls":
		err = cmdLs(os.Args[2:])
	case "loc":
		err = cmdLoc(os.Args[2:])
	case "voice":
		err = cmdVoice(os.Args[2:])
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
  ls                       list all discovered source files
  export --platform NAME   build and export (windows|mac|linux|web|ios|android)
  export --locale LANG     export dialogue strings as PO/CSV for translation
                           flags: --format po|csv  --diff (changed/untranslated only)
  export --voicescript     export per-character voice actor scripts from .agcut files
                           flags: --character NAME  --locale LANG (include translations)
  new NAME                 scaffold a new AGS3D project
  viz tokens      FILE     print token stream (line/col/kind/lexeme)
  viz ast         FILE     print AST tree (text)
  viz ast-dot     FILE     print AST as Graphviz DOT  (pipe to: dot -Tsvg -o ast.svg)
  viz symbols     FILE     print symbol table (text)
  viz symbols-dot FILE     print symbol table as Graphviz DOT
  viz blocking    FILE     print blocking call annotations
  viz emit        FILE     print side-by-side AGS-spirit ↔ GDScript
  viz             FILE     run all viz stages`)
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
	case "ast-dot":
		viz.ASTDot(os.Stdout, file, content)
	case "symbols":
		viz.Symbols(os.Stdout, file, content)
	case "symbols-dot":
		viz.SymbolsDot(os.Stdout, file, content)
	case "blocking":
		viz.Blocking(os.Stdout, file, content)
	case "emit":
		viz.Emit(os.Stdout, file, content)
	case "all":
		viz.Tokens(os.Stdout, file, content)
		fmt.Fprintln(os.Stdout)
		viz.AST(os.Stdout, file, content)
		fmt.Fprintln(os.Stdout)
		viz.Symbols(os.Stdout, file, content)
		fmt.Fprintln(os.Stdout)
		viz.Blocking(os.Stdout, file, content)
		fmt.Fprintln(os.Stdout)
		viz.Emit(os.Stdout, file, content)
	default:
		return fmt.Errorf("unknown viz stage %q — expected tokens, ast, ast-dot, symbols, symbols-dot, blocking, or emit", stage)
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

func cmdBuild(args []string) error {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	force := fs.Bool("force", false, "rebuild all files regardless of mtime")
	trace := fs.Bool("trace", false, "emit print() debug traces around blocking calls")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, _ := requireProject()
	return build(root, *force, *trace)
}

func build(root string, force bool, trace bool) error {
	all, err := project.Scan(root)
	if err != nil {
		return err
	}

	var scripts, rooms, chars, items, guis, dialogues, cutscenes []project.SourceFile
	for _, f := range all {
		switch f.Ext {
		case ".agscript":
			scripts = append(scripts, f)
		case ".agroom":
			rooms = append(rooms, f)
		case ".agchar":
			chars = append(chars, f)
		case ".agitem":
			items = append(items, f)
		case ".agui":
			guis = append(guis, f)
		case ".agdlg":
			dialogues = append(dialogues, f)
		case ".agcut":
			cutscenes = append(cutscenes, f)
		}
	}

	manifest, err := project.LoadManifest(root)
	if err != nil {
		return err
	}

	var changedScripts, changedRooms, changedChars, changedItems, changedGUIs, changedDialogues, changedCutscenes []project.SourceFile
	if force {
		changedScripts, changedRooms, changedChars, changedItems, changedGUIs, changedDialogues, changedCutscenes = scripts, rooms, chars, items, guis, dialogues, cutscenes
	} else {
		changedScripts = project.Changed(scripts, manifest)
		changedRooms = project.Changed(rooms, manifest)
		changedChars = project.Changed(chars, manifest)
		changedItems = project.Changed(items, manifest)
		changedGUIs = project.Changed(guis, manifest)
		changedDialogues = project.Changed(dialogues, manifest)
		changedCutscenes = project.Changed(cutscenes, manifest)
	}

	total := len(changedScripts) + len(changedRooms) + len(changedChars) + len(changedItems) + len(changedGUIs) + len(changedDialogues) + len(changedCutscenes)
	if total == 0 {
		fmt.Println("ag build: nothing to do (no changed source files)")
		return nil
	}

	fmt.Printf("ag build: %d file(s) to process\n", total)

	generatedDir := filepath.Join(root, ".engine", "generated")
	if err := os.MkdirAll(generatedDir, 0755); err != nil {
		return err
	}

	var errs []error

	// --- .agscript → .gd (GDScript transpilation) ---
	em := emitter.New()
	em.Trace = trace
	for _, src := range changedScripts {
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

		outPath := filepath.Join(generatedDir, src.Rel+".gd")
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			errs = append(errs, err)
			continue
		}
		if err := os.WriteFile(outPath, []byte(result.GDScript), 0644); err != nil {
			errs = append(errs, err)
			continue
		}

		// Write .agmap sidecar source map (T17):
		// [[gdscript_line, "rel/path.agscript", agscript_line], ...]
		agmapPath := strings.TrimSuffix(outPath, ".gd") + ".agmap"
		if mapJSON, jsonErr := json.Marshal(result.SourceMap); jsonErr == nil {
			_ = os.WriteFile(agmapPath, mapJSON, 0644)
		}

		project.RecordMtimes([]project.SourceFile{src}, manifest)
		fmt.Printf("  %s → %s\n", src.Rel, filepath.Join(".engine/generated", src.Rel+".gd"))
	}

	// --- .agroom → .tscn (room scene generation) ---
	for _, src := range changedRooms {
		data, err := os.ReadFile(src.Path)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		rd, err := room.ParseRoom(src.Rel, string(data))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			errs = append(errs, err)
			continue
		}
		// Derive companion script path (same name, .agscript extension).
		scriptRelPath := strings.TrimSuffix(src.Rel, ".agroom") + ".agscript"
		// Detect optional .glb visual mesh beside the .agroom file.
		glbPath := strings.TrimSuffix(src.Path, ".agroom") + ".glb"
		glbRelPath := ""
		if _, statErr := os.Stat(glbPath); statErr == nil {
			glbRelPath = strings.TrimSuffix(src.Rel, ".agroom") + ".glb"
		}
		tscnText := scene.GenerateRoomScene(rd, scriptRelPath, glbRelPath)

		// Write .tscn beside the .agroom source file.
		outPath := strings.TrimSuffix(src.Path, ".agroom") + ".tscn"
		if err := os.WriteFile(outPath, []byte(tscnText), 0644); err != nil {
			errs = append(errs, err)
			continue
		}

		project.RecordMtimes([]project.SourceFile{src}, manifest)
		outRel := strings.TrimSuffix(src.Rel, ".agroom") + ".tscn"
		fmt.Printf("  %s → %s\n", src.Rel, outRel)
	}

	// --- .agchar → .tscn (character scene generation) ---
	for _, src := range changedChars {
		data, err := os.ReadFile(src.Path)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		cd, err := char.ParseChar(src.Rel, string(data))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			errs = append(errs, err)
			continue
		}
		// T-CUT28: load .aganim sidecar if the .agchar references a .glb mesh.
		var animFile *aganim.AnimFile
		if cd.Mesh != "" {
			sidecarPath := aganim.SidecarPath(cd.Mesh)
			// Resolve relative to the project root (src.Path parent → root).
			// cd.Mesh is project-relative (e.g. "characters/player/player.glb").
			// src.Path is absolute; walk up to find the project root.
			absDir := filepath.Dir(src.Path)
			// Try the directory containing the .agchar first, then one level up.
			for _, candidate := range []string{
				filepath.Join(absDir, filepath.Base(sidecarPath)),
				filepath.Join(filepath.Dir(absDir), sidecarPath),
				sidecarPath,
			} {
				if af, err2 := aganim.ParseFile(candidate); err2 == nil {
					animFile = af
					break
				}
			}
		}

		tscnText := scene.GenerateCharScene(cd, animFile)

		// Write .tscn beside the .agchar source file.
		outPath := strings.TrimSuffix(src.Path, ".agchar") + ".tscn"
		if err := os.WriteFile(outPath, []byte(tscnText), 0644); err != nil {
			errs = append(errs, err)
			continue
		}

		project.RecordMtimes([]project.SourceFile{src}, manifest)
		outRel := strings.TrimSuffix(src.Rel, ".agchar") + ".tscn"
		fmt.Printf("  %s → %s\n", src.Rel, outRel)
	}

	// --- .agui → .tscn (GUI CanvasLayer scene generation) ---
	for _, src := range changedGUIs {
		data, err := os.ReadFile(src.Path)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		gd, err := gui.ParseGUI(src.Rel, string(data))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			errs = append(errs, err)
			continue
		}
		tscnText := scene.GenerateGUIScene(gd)

		outPath := strings.TrimSuffix(src.Path, ".agui") + ".tscn"
		if err := os.WriteFile(outPath, []byte(tscnText), 0644); err != nil {
			errs = append(errs, err)
			continue
		}

		project.RecordMtimes([]project.SourceFile{src}, manifest)
		outRel := strings.TrimSuffix(src.Rel, ".agui") + ".tscn"
		fmt.Printf("  %s → %s\n", src.Rel, outRel)
	}

	// --- .agitem (data-only, no scene output — parse to validate) ---
	for _, src := range changedItems {
		data, err := os.ReadFile(src.Path)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if _, err := item.ParseItem(src.Rel, string(data)); err != nil {
			fmt.Fprintln(os.Stderr, err)
			errs = append(errs, err)
			continue
		}
		project.RecordMtimes([]project.SourceFile{src}, manifest)
		fmt.Printf("  %s (item — data only, no scene generated)\n", src.Rel)
	}

	// --- .agdlg → .engine/generated/dialogue/*.json (dialogue compilation) ---
	// All changed .agdlg files are re-parsed together so cross-file jump
	// references resolve correctly across the changed set. When --force is not
	// set, unchanged files are not re-emitted but are still parsed for the
	// link pass so references into them resolve correctly.
	if len(changedDialogues) > 0 {
		// Parse all dialogue files for the link pass.
		var dlgFiles []*dlg.DialogueFile
		for _, src := range dialogues {
			df, err := dlg.ParseFile(src.Path)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				errs = append(errs, err)
			} else {
				dlgFiles = append(dlgFiles, df)
			}
		}
		if len(errs) == 0 || len(dlgFiles) > 0 {
			lp, linkErr := dlg.Link(dlgFiles)
			if linkErr != nil {
				fmt.Fprintln(os.Stderr, linkErr)
				errs = append(errs, linkErr)
			} else {
				// Structural validation.
				valErrs := dlg.Validate(lp)
				for _, ve := range valErrs {
					fmt.Fprintln(os.Stderr, ve)
					errs = append(errs, ve)
				}
				if len(valErrs) == 0 {
					// Emit only the changed files.
					changedSet := make(map[string]bool, len(changedDialogues))
					for _, src := range changedDialogues {
						changedSet[src.Path] = true
					}
					var toEmit []*dlg.DialogueFile
					for _, df := range dlgFiles {
						if changedSet[df.Path] {
							toEmit = append(toEmit, df)
						}
					}
					changedLP := &dlg.LinkedProject{
						Files:        toEmit,
						NodesByTitle: lp.NodesByTitle,
					}
					dlgOutDir := filepath.Join(root, ".engine", "generated", "dialogue")
					for _, emitErr := range dlg.EmitProject(changedLP, dlgOutDir) {
						fmt.Fprintln(os.Stderr, emitErr)
						errs = append(errs, emitErr)
					}
					if len(errs) == 0 {
						project.RecordMtimes(changedDialogues, manifest)
						for _, src := range changedDialogues {
							base := strings.TrimSuffix(filepath.Base(src.Path), ".agdlg") + ".json"
							fmt.Printf("  %s → .engine/generated/dialogue/%s\n", src.Rel, base)
						}
					}
				}
			}
		}
	}

	// --- .agcut → .engine/generated/cutscenes/*.json (cutscene compilation) ---
	// All cutscene files are parsed together for multi-file validation; only
	// changed files are re-emitted.
	if len(changedCutscenes) > 0 {
		var cutFiles []*cut.CutsceneFile
		for _, src := range cutscenes {
			cf, err := cut.ParseFile(src.Path)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				errs = append(errs, err)
			} else {
				cutFiles = append(cutFiles, cf)
			}
		}
		if len(cutFiles) > 0 {
			// Structural validation (errors block build).
			valErrs := cut.ValidateProjectCutscenes(cutFiles, nil)
			for _, ve := range valErrs {
				fmt.Fprintln(os.Stderr, ve)
				errs = append(errs, ve)
			}
			// Sequencing validation + warnings (non-blocking).
			for _, cf := range cutFiles {
				for _, se := range cut.ValidateSequence(cf) {
					fmt.Fprintf(os.Stderr, "warning: %s\n", se)
				}
				for _, w := range cut.WarnCutscene(cf, nil, nil) {
					fmt.Fprintf(os.Stderr, "warning: %s\n", w)
				}
			}
			if len(valErrs) == 0 {
				changedSet := make(map[string]bool, len(changedCutscenes))
				for _, src := range changedCutscenes {
					changedSet[src.Path] = true
				}
				var toEmit []*cut.CutsceneFile
				for _, cf := range cutFiles {
					if changedSet[cf.Path] {
						toEmit = append(toEmit, cf)
					}
				}
				cutOutDir := filepath.Join(root, ".engine", "generated", "cutscenes")
				for _, emitErr := range cut.EmitCutscenes(toEmit, cutOutDir) {
					fmt.Fprintln(os.Stderr, emitErr)
					errs = append(errs, emitErr)
				}
				if len(errs) == 0 {
					project.RecordMtimes(changedCutscenes, manifest)
					for _, src := range changedCutscenes {
						base := strings.TrimSuffix(filepath.Base(src.Path), ".agcut") + ".json"
						fmt.Printf("  %s → .engine/generated/cutscenes/%s\n", src.Rel, base)
					}
				}
			}
		}
	}

	// --- locale .agstrings → .engine/generated/locale/<code>.json (T-LOC07) ---
	compiledLocales, locErr := loc.CompileAllLocales(root)
	if locErr != nil {
		fmt.Fprintf(os.Stderr, "ag build: compile locales: %v\n", locErr)
		errs = append(errs, locErr)
	} else if len(compiledLocales) > 0 {
		localeOutDir := filepath.Join(root, ".engine", "generated", "locale")
		if mkErr := os.MkdirAll(localeOutDir, 0755); mkErr != nil {
			errs = append(errs, mkErr)
		} else {
			for code, cl := range compiledLocales {
				jsonBytes, writeErr := loc.WriteCompiledLocale(cl)
				if writeErr != nil {
					errs = append(errs, writeErr)
					continue
				}
				outPath := filepath.Join(localeOutDir, code+".json")
				if wErr := os.WriteFile(outPath, jsonBytes, 0644); wErr != nil {
					errs = append(errs, wErr)
					continue
				}
				rel, _ := filepath.Rel(root, outPath)
				fmt.Printf("  %s → %s\n", filepath.Join("locale", code+".agstrings"), rel)
			}
		}
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
// ag ls
// -------------------------------------------------------------------

func cmdLs(args []string) error {
	fs := flag.NewFlagSet("ls", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: ag ls [project]")
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	root := "."
	if fs.NArg() >= 1 {
		root = fs.Arg(0)
	}

	all, err := project.Scan(root)
	if err != nil {
		return err
	}

	var scripts, rooms, chars, items, guis, dialogues, cutscenes []string
	for _, f := range all {
		switch f.Ext {
		case ".agscript":
			scripts = append(scripts, f.Rel)
		case ".agroom":
			rooms = append(rooms, f.Rel)
		case ".agchar":
			chars = append(chars, f.Rel)
		case ".agitem":
			items = append(items, f.Rel)
		case ".agui":
			guis = append(guis, f.Rel)
		case ".agdlg":
			dialogues = append(dialogues, f.Rel)
		case ".agcut":
			cutscenes = append(cutscenes, f.Rel)
		}
	}

	printSection("rooms", rooms)
	printSection("characters", chars)
	printSection("items", items)
	printSection("dialogues", dialogues)
	printSection("cutscenes", cutscenes)
	printSection("guis", guis)
	printSection("scripts", scripts)

	total := len(scripts) + len(rooms) + len(chars) + len(items) + len(guis) + len(dialogues) + len(cutscenes)
	fmt.Fprintf(os.Stderr, "\n%d source file(s) in %q\n", total, root)
	return nil
}

func printSection(name string, files []string) {
	if len(files) == 0 {
		return
	}
	fmt.Printf("\n[%s]\n", name)
	for _, f := range files {
		fmt.Printf("  %s\n", f)
	}
}

// -------------------------------------------------------------------
// ag run
// -------------------------------------------------------------------

func cmdRun(_ []string) error {
	root, _ := requireProject()
	if err := build(root, false, false); err != nil {
		return err
	}
	godot, err := findGodot()
	if err != nil {
		return err
	}
	// Use the project root as the Godot project path when project.godot is
	// present there (the normal case for game_prototype and ag-new projects).
	godotProject := root
	if _, err := os.Stat(filepath.Join(root, "project.godot")); errors.Is(err, os.ErrNotExist) {
		// Fallback: look in .engine/ for legacy or generated project layouts.
		godotProject = filepath.Join(root, ".engine")
	}
	fmt.Printf("ag run: launching %s --editor --path %s\n", godot, godotProject)
	cmd := exec.Command(godot, "--editor", "--path", godotProject)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func findGodot() (string, error) {
	// 1. Explicit override via environment variable.
	if env := os.Getenv("GODOT"); env != "" {
		if path, err := exec.LookPath(env); err == nil {
			return path, nil
		}
	}
	// 2. Standard names on PATH.
	for _, name := range []string{"godot", "godot4", "Godot"} {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	// 3. Repo-local build artefact: bin/ sibling of the ag binary's directory.
	//    ag lives at tools/ag/ag (or bin/ after install), so walk up to find bin/.
	exe, _ := os.Executable()
	for dir := filepath.Dir(exe); dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
		matches, _ := filepath.Glob(filepath.Join(dir, "bin", "godot.linuxbsd.editor.*"))
		if len(matches) > 0 {
			return matches[0], nil
		}
	}
	return "", fmt.Errorf("Godot binary not found — set the GODOT environment variable or ensure 'godot' is on PATH")
}

// -------------------------------------------------------------------
// ag validate
// -------------------------------------------------------------------

func cmdValidate(_ []string) error {
	var issues []validate.Issue
	var err error

	// If stdin is a pipe, read newline-separated file paths from it.
	// This lets users do: find . -name "*.agdlg" | ag validate
	// All provided files are analysed together so cross-file checks work.
	if stdinIsPipe() {
		files, readErr := readFilesFromStdin()
		if readErr != nil {
			return readErr
		}
		issues, err = validate.ValidateFiles(files)
	} else {
		root, manifest := requireProject()
		issues, err = validate.ValidateProject(root, manifest)
	}

	if err != nil {
		return err
	}
	if len(issues) == 0 {
		fmt.Println("ag validate: no issues found")
		return nil
	}
	errorCount := 0
	for _, issue := range issues {
		fmt.Fprintln(os.Stderr, issue)
		if issue.Severity == "error" {
			errorCount++
		}
	}
	if errorCount > 0 {
		return fmt.Errorf("%d error(s)", errorCount)
	}
	return nil
}

// stdinIsPipe reports whether os.Stdin is connected to a pipe or redirect
// rather than an interactive terminal.
func stdinIsPipe() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice == 0
}

// readFilesFromStdin reads newline-separated file paths from stdin and returns
// a []project.SourceFile. Paths are made absolute; Rel is relative to cwd.
// Blank lines and lines starting with '#' are ignored.
func readFilesFromStdin() ([]project.SourceFile, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	var files []project.SourceFile
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		abs, err := filepath.Abs(line)
		if err != nil {
			return nil, fmt.Errorf("invalid path %q: %w", line, err)
		}
		ext := filepath.Ext(abs)
		rel, err := filepath.Rel(cwd, abs)
		if err != nil {
			rel = abs
		}
		files = append(files, project.SourceFile{Path: abs, Rel: rel, Ext: ext})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return files, nil
}

// -------------------------------------------------------------------
// ag export
// -------------------------------------------------------------------

func cmdExport(args []string) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	platform := fs.String("platform", "", "export target (windows|mac|linux|web|ios|android)")
	locale := fs.String("locale", "", "export dialogue strings for locale (e.g. en, es, he)")
	format := fs.String("format", "po", "output format for --locale: po or csv")
	diff := fs.Bool("diff", false, "only emit strings missing or changed since last export")
	voicescript := fs.Bool("voicescript", false, "export per-character voice actor scripts from .agcut files")
	charFilter := fs.String("character", "", "limit voicescript to one character name")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *voicescript {
		return cmdExportVoicescript(*locale, *charFilter)
	}

	if *locale != "" {
		return cmdExportLocale(*locale, *format, *diff)
	}

	if *platform == "" {
		root, _ := requireProject()
		written, err := loc.ExportLocaleForProject(root)
		if err != nil {
			return err
		}
		for _, rel := range written {
			fmt.Printf("ag export: %s\n", rel)
		}
		fmt.Printf("ag export: %d locale(s) written\n", len(written))
		return nil
	}
	// TODO(T18): build then invoke Godot's export pipeline.
	fmt.Printf("ag export: export pipeline for %q not yet implemented (T18)\n", *platform)
	return nil
}

func cmdExportLocale(lang, format string, diffOnly bool) error {
	if lang == "" {
		return fmt.Errorf("--locale requires a language code (e.g. en, fr)")
	}
	if format != "po" && format != "csv" && format != "agstrings" {
		return fmt.Errorf("--format must be po, csv, or agstrings")
	}

	root, _ := requireProject()
	all, err := project.Scan(root)
	if err != nil {
		return err
	}

	var dlgFiles []*dlg.DialogueFile
	for _, f := range all {
		if f.Ext != ".agdlg" {
			continue
		}
		df, err := dlg.ParseFile(f.Path)
		if err != nil {
			return fmt.Errorf("export locale: parse %s: %w", f.Rel, err)
		}
		dlgFiles = append(dlgFiles, df)
	}
	if len(dlgFiles) == 0 {
		fmt.Println("ag export: no .agdlg files found")
		return nil
	}

	lp, linkErr := dlg.Link(dlgFiles)
	if linkErr != nil {
		return fmt.Errorf("export locale: %w", linkErr)
	}

	entries := dlg.CollectLocEntries(lp)
	if len(entries) == 0 {
		fmt.Println("ag export: no localizable strings found")
		return nil
	}

	// Determine output path and load existing translations for --diff.
	localeDir := filepath.Join(root, "locale")
	if err := os.MkdirAll(localeDir, 0755); err != nil {
		return fmt.Errorf("export locale: create locale dir: %w", err)
	}

	var ext, outPath string
	switch format {
	case "csv":
		ext = ".csv"
	case "agstrings":
		ext = ".agstrings"
	default:
		ext = ".po"
	}
	outPath = filepath.Join(localeDir, lang+ext)

	// agstrings format uses the loc package's Diff/Apply pipeline.
	if format == "agstrings" {
		return cmdExportLocaleAgstrings(entries, lang, outPath, diffOnly)
	}

	existing := map[string]string{}
	if diffOnly {
		if raw, readErr := os.ReadFile(outPath); readErr == nil {
			switch format {
			case "csv":
				existing = dlg.ParseCSVTranslations(string(raw))
			default:
				existing = dlg.ParsePOTranslations(string(raw))
			}
		}
		// If the file doesn't exist yet, existing stays empty — all entries emitted.
	}

	var content string
	switch format {
	case "csv":
		content = dlg.ExportCSV(entries, existing, diffOnly)
	default:
		content = dlg.ExportPO(entries, existing, diffOnly)
	}

	if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("export locale: write %s: %w", outPath, err)
	}

	rel, _ := filepath.Rel(root, outPath)
	fmt.Printf("ag export: %d strings → %s\n", len(entries), rel)
	return nil
}

// cmdExportLocaleAgstrings exports dialogue strings in .agstrings format using
// the loc package's Diff/Apply pipeline for incremental update tracking.
// If an existing .agstrings file is present, Diff() marks new keys as empty
// and removed keys as orphans. Apply() merges the diff into the base file.
func cmdExportLocaleAgstrings(entries []dlg.LocEntry, lang, outPath string, diffOnly bool) error {
	// Build the current key list from the dialogue entries.
	currentKeys := make([]string, 0, len(entries))
	for _, e := range entries {
		currentKeys = append(currentKeys, e.LocKey)
	}

	// Load existing .agstrings file (if any).
	var base *loc.StringsFile
	if raw, readErr := os.ReadFile(outPath); readErr == nil {
		var parseErr error
		base, parseErr = loc.Parse(outPath, string(raw))
		if parseErr != nil {
			base = nil // treat as missing
		}
	}
	if base == nil {
		base = &loc.StringsFile{Meta: loc.Meta{BaseLocale: "en", Locale: lang}}
	}

	// Diff: detect new / removed / unchanged keys.
	diff := loc.Diff(base, currentKeys)

	// Apply diff to produce the updated file.
	updated := loc.Apply(base, diff)

	if diffOnly {
		// In diff-only mode, keep only untranslated, stale, or orphan entries.
		var filtered []loc.Entry
		for _, e := range updated.Entries {
			if e.Value == "" || e.Stale || e.Orphan {
				filtered = append(filtered, e)
			}
		}
		updated.Entries = filtered
	}

	content := loc.Write(updated)
	if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("export locale: write %s: %w", outPath, err)
	}
	rel, _ := filepath.Rel(filepath.Dir(filepath.Dir(outPath)), outPath)
	fmt.Printf("ag export: %d strings → %s\n", len(entries), rel)
	return nil
}

func cmdExportVoicescript(localeFilter, charFilter string) error {
	root, _ := requireProject()
	all, err := project.Scan(root)
	if err != nil {
		return err
	}

	var cutFiles []*cut.CutsceneFile
	for _, f := range all {
		if f.Ext != ".agcut" {
			continue
		}
		cf, err := cut.ParseFile(f.Path)
		if err != nil {
			return fmt.Errorf("voicescript: parse %s: %w", f.Rel, err)
		}
		cutFiles = append(cutFiles, cf)
	}
	if len(cutFiles) == 0 {
		fmt.Println("ag export: no .agcut files found")
		return nil
	}

	lines := cut.CollectVoiceLines(cutFiles)
	if len(lines) == 0 {
		fmt.Println("ag export: no <<line>> commands found in cutscene files")
		return nil
	}

	// Optionally load translations from locale/<lang>.po.
	var translations map[string]string
	if localeFilter != "" {
		poPath := filepath.Join(root, "locale", localeFilter+".po")
		if raw, readErr := os.ReadFile(poPath); readErr == nil {
			translations = dlg.ParsePOTranslations(string(raw))
		}
	}

	groups := cut.RenderVoicescripts(lines, translations, charFilter)
	if len(groups) == 0 {
		fmt.Println("ag export: no lines matched the specified filters")
		return nil
	}

	outRoot := filepath.Join(root, "voicescripts")
	for key, content := range groups {
		outPath := filepath.Join(outRoot, key+".md")
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("voicescript: mkdir %s: %w", filepath.Dir(outPath), err)
		}
		if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("voicescript: write %s: %w", outPath, err)
		}
		rel, _ := filepath.Rel(root, outPath)
		fmt.Printf("ag export: voicescripts/%s\n", rel)
	}
	return nil
}

// -------------------------------------------------------------------
// ag new
// -------------------------------------------------------------------

func cmdLoc(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: ag loc check|find|filter|report|import|tui [args]")
	}
	switch args[0] {
	case "check":
		return cmdLocCheck(args[1:])
	case "find":
		return cmdLocFind(args[1:])
	case "filter":
		return cmdLocFilter(args[1:])
	case "report":
		return cmdLocReport(args[1:])
	case "import":
		return cmdLocImport(args[1:])
	case "tui":
		return cmdLocTUI(args[1:])
	default:
		return fmt.Errorf("ag loc: unknown subcommand %q (check|find|filter|report|import|tui)", args[0])
	}
}

// ag voice — voice file coverage tracking

func cmdVoice(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: ag voice coverage [args]")
	}
	switch args[0] {
	case "coverage":
		return cmdVoiceCoverage(args[1:])
	default:
		return fmt.Errorf("ag voice: unknown subcommand %q (coverage)", args[0])
	}
}

func cmdVoiceCoverage(args []string) error {
	fs := flag.NewFlagSet("ag voice coverage", flag.ContinueOnError)
	locale := fs.String("locale", "en", "locale code")
	scan := fs.Bool("scan", false, "scan audio directory and output voice_coverage.json manifest")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: ag voice coverage <project> [--locale LANG] [--scan]")
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		fs.Usage()
		return fmt.Errorf("missing project argument")
	}

	root := fs.Arg(0)

	cutFiles, err := collectCutsceneFiles(root)
	if err != nil {
		return fmt.Errorf("scan cutscenes: %w", err)
	}

	if *scan {
		entries, err := cut.ScanVoiceDirectory(root, *locale)
		if err != nil {
			return fmt.Errorf("scan audio directory: %w", err)
		}
		data, _ := json.MarshalIndent(map[string]interface{}{
			"version": 1,
			"locale":  *locale,
			"entries": entries,
		}, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	entries, _ := cut.ScanVoiceDirectory(root, *locale)
	report := cut.BuildVoiceCoverageReport(cutFiles, entries, nil)

	var sb strings.Builder
	cut.WriteVoiceCoverageJSON(report, &sb, *locale)
	fmt.Print(sb.String())

	missing := len(report.Missing)
	covered := len(report.Covered)
	stale := len(report.Stale)
	total := missing + covered + stale
	if total == 0 {
		fmt.Fprintf(os.Stderr, "\nag voice coverage: no voice lines found in project\n")
		return nil
	}
	pct := float64(covered) / float64(total) * 100
	fmt.Fprintf(os.Stderr, "\nag voice coverage: %d/%d recorded (%.0f%%), %d missing, %d stale\n",
		covered, total, pct, missing, stale)
	return nil
}

func collectCutsceneFiles(root string) ([]*cut.CutsceneFile, error) {
	cutDir := filepath.Join(root, "cutscenes")
	info, err := os.Stat(cutDir)
	if err != nil || !info.IsDir() {
		return nil, nil
	}
	var files []*cut.CutsceneFile
	err = filepath.WalkDir(cutDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".agcut" {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		cf, err := cut.Parse(rel, string(data))
		if err != nil {
			return nil
		}
		files = append(files, cf)
		return nil
	})
	return files, err
}

func cmdLocFind(args []string) error {
	fs := flag.NewFlagSet("ag loc find", flag.ContinueOnError)
	locale := fs.String("locale", "en", "locale code")
	pattern := fs.String("pattern", "*", "glob pattern to match loc_key or source text")
	groupBy := fs.String("group-by", "", "group results by: character, node, or type")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: ag loc find <project> [--locale LANG] [--pattern GLOB] [--group-by character|node|type]")
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		fs.Usage()
		return fmt.Errorf("missing project argument")
	}

	root := fs.Arg(0)
	entries, err := loc.CollectAllLocaleEntriesWithTranslations(root, *locale)
	if err != nil {
		return fmt.Errorf("collect entries: %w", err)
	}

	if *pattern != "" && *pattern != "*" {
		entries = loc.FindLocaleEntries(entries, *pattern)
	}

	if *groupBy != "" {
		fmt.Print(loc.FormatLocaleFind(entries, *groupBy))
	} else {
		fmt.Print(loc.FormatLocaleFind(entries, ""))
	}

	fmt.Fprintf(os.Stderr, "\nag loc find: %d entries matching %q\n", len(entries), *pattern)
	return nil
}

func cmdLocFilter(args []string) error {
	fs := flag.NewFlagSet("ag loc filter", flag.ContinueOnError)
	locale := fs.String("locale", "en", "locale code")
	untranslated := fs.Bool("untranslated", false, "show only untranslated strings")
	char := fs.String("char", "", "filter by character name")
	node := fs.String("node", "", "filter by node/scene name")
	lineType := fs.String("type", "", "filter by line type (spoken, choice, narration, ui, subtitle)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: ag loc filter <project> [--locale LANG] [--untranslated] [--char NAME] [--node NAME] [--type TYPE]")
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		fs.Usage()
		return fmt.Errorf("missing project argument")
	}

	root := fs.Arg(0)
	entries, err := loc.CollectAllLocaleEntriesWithTranslations(root, *locale)
	if err != nil {
		return fmt.Errorf("collect entries: %w", err)
	}

	entries = loc.FilterLocaleEntries(entries, loc.FilterOptions{
		Untranslated: *untranslated,
		Type:         *lineType,
		Char:         *char,
		Node:         *node,
	})

	fmt.Print(loc.FormatLocaleFind(entries, ""))
	fmt.Fprintf(os.Stderr, "\nag loc filter: %d entries\n", len(entries))
	return nil
}

func cmdLocCheck(args []string) error {
	fs := flag.NewFlagSet("ag loc check", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: ag loc check <project>")
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		fs.Usage()
		return fmt.Errorf("missing project argument")
	}

	root := fs.Arg(0)
	_, loadErr := project.Load(root)
	if loadErr != nil {
		return fmt.Errorf("load project: %w", loadErr)
	}

	files, scanErr := project.Scan(root)
	if scanErr != nil {
		return fmt.Errorf("scan project: %w", scanErr)
	}

	issues, validateErr := validate.ValidateFiles(files)
	if validateErr != nil {
		return fmt.Errorf("validate: %w", validateErr)
	}

	if len(issues) == 0 {
		fmt.Println("ag loc check: no localisation issues found")
		return nil
	}

	var errors, warnings int
	for _, issue := range issues {
		if issue.Severity == "error" {
			errors++
		} else {
			warnings++
		}
		fmt.Printf("%s\n", issue)
	}

	fmt.Fprintf(os.Stderr, "\n%d errors, %d warnings\n", errors, warnings)
	if errors > 0 {
		return fmt.Errorf("validation failed")
	}
	return nil
}

func cmdLocReport(args []string) error {
	fs := flag.NewFlagSet("ag loc report", flag.ContinueOnError)
	locale := fs.String("locale", "en", "locale code")
	byCharacter := fs.String("by-character", "", "show all strings for this character name")
	byNode := fs.String("by-node", "", "show all strings for this node/scene name")
	untranslated := fs.Bool("untranslated", false, "show only untranslated strings")
	groupBy := fs.String("group-by", "", "group report by: character, node, type")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: ag loc report <project> [--locale LANG] [--by-character NAME] [--by-node NAME] [--untranslated] [--group-by character|node|type]")
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		fs.Usage()
		return fmt.Errorf("missing project argument")
	}

	root := fs.Arg(0)
	entries, err := loc.CollectAllLocaleEntriesWithTranslations(root, *locale)
	if err != nil {
		return fmt.Errorf("collect entries: %w", err)
	}

	filterOpts := loc.FilterOptions{
		Untranslated: *untranslated,
		Char:         *byCharacter,
		Node:         *byNode,
	}
	entries = loc.FilterLocaleEntries(entries, filterOpts)

	if *groupBy != "" {
		fmt.Print(loc.FormatLocaleReportGrouped(entries, *groupBy, *locale))
	} else {
		fmt.Print(loc.FormatLocaleReport(entries, *locale))
	}

	fmt.Fprintf(os.Stderr, "\nag loc report: %d strings for locale %s\n", len(entries), *locale)
	return nil
}

func cmdLocImport(args []string) error {
	fs := flag.NewFlagSet("ag loc import", flag.ContinueOnError)
	locale := fs.String("locale", "", "locale code (required)")
	file := fs.String("file", "", "import file path (required)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: ag loc import <project> --locale LANG --file PATH")
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		fs.Usage()
		return fmt.Errorf("missing project argument")
	}
	if *locale == "" {
		fs.Usage()
		return fmt.Errorf("--locale is required")
	}
	if *file == "" {
		fs.Usage()
		return fmt.Errorf("--file is required")
	}

	root := fs.Arg(0)

	ext := filepath.Ext(*file)
	var result *loc.ImportResult
	var err error
	switch ext {
	case ".po":
		result, err = loc.ImportPO(root, *locale, *file)
	case ".csv":
		result, err = loc.ImportCSV(root, *locale, *file)
	default:
		return fmt.Errorf("unsupported file format %q (use .po or .csv)", ext)
	}
	if err != nil {
		return fmt.Errorf("import: %w", err)
	}

	fmt.Printf("ag loc import: processed %d entries from %s\n", result.Imported+result.Skipped+result.Invalid, *file)
	fmt.Printf("  imported: %d\n", result.Imported)
	if result.Skipped > 0 {
		fmt.Printf("  skipped (empty): %d\n", result.Skipped)
	}
	if result.Invalid > 0 {
		fmt.Fprintf(os.Stderr, "  invalid (not in project): %d\n", result.Invalid)
		for _, e := range result.Errors {
			fmt.Fprintf(os.Stderr, "    %s\n", e)
		}
	}

	fmt.Println("ag loc import: merge complete")
	return nil
}

func cmdLocTUI(args []string) error {
	locale := ""
	remaining := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] == "--locale" && i+1 < len(args) {
			locale = args[i+1]
			i++
		} else if strings.HasPrefix(args[i], "--locale=") {
			locale = strings.TrimPrefix(args[i], "--locale=")
		} else {
			remaining = append(remaining, args[i])
		}
	}
	if len(remaining) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: ag loc tui <project> [--locale LANG]")
		return fmt.Errorf("missing project argument")
	}
	return tui.RunTUIMain(remaining[0], locale)
}

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

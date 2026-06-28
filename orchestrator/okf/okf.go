package okf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"acre/opencode"
)

// Generate runs the OKF documentation pipeline for a repository.
func Generate(repoPath string) error {
	repoAbs, err := filepath.Abs(repoPath)
	if err != nil {
		return fmt.Errorf("invalid repository path: %w", err)
	}

	repoName := filepath.Base(repoAbs)
	fmt.Printf("🔍 Scanning repository: %s\n", repoAbs)
	fmt.Printf("   Generating OKF v0.1 index under OKF/%s...\n\n", repoName)

	prompt := constructOKFPrompt(repoName)

	fmt.Println("🤖 Executing OpenCode to analyze the codebase and generate OKF files...")
	out, err := opencode.Run(prompt, repoAbs)
	if err != nil {
		return fmt.Errorf("OpenCode failed to index codebase: %w\nOutput: %s", err, out)
	}

	codebaseOKFDir := filepath.Join(repoAbs, "OKF")
	info, err := os.Stat(codebaseOKFDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("OpenCode finished but did not generate the OKF/ directory inside the repository. Output:\n%s", out)
	}

	// Move/copy OKF files to ACRE's OKF/<repoName>/
	destOKFDir := filepath.Join("OKF", repoName)
	// Try one level up if ACRE is run from orchestrator directory
	if _, statErr := os.Stat("orchestrator"); statErr == nil {
		// Currently in project root
	} else {
		// Currently in orchestrator directory, destination is "../OKF/<repoName>"
		destOKFDir = filepath.Join("..", "OKF", repoName)
	}

	if err := os.MkdirAll(destOKFDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination OKF folder: %w", err)
	}

	// Read and copy all files
	copiedCount := 0
	err = filepath.Walk(codebaseOKFDir, func(path string, fileInfo os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".md") {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			rel, _ := filepath.Rel(codebaseOKFDir, path)
			destPath := filepath.Join(destOKFDir, rel)
			
			// Ensure parent folder in destination exists
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return err
			}
			
			if err := os.WriteFile(destPath, data, 0644); err != nil {
				return err
			}
			fmt.Printf("   [OKF file created] %s\n", rel)
			copiedCount++
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to copy OKF files: %w", err)
	}

	// Clean up the temporary folder inside target codebase
	_ = os.RemoveAll(codebaseOKFDir)

	fmt.Printf("\n✨ Successfully generated %d OKF concept files for '%s'!\n", copiedCount, repoName)
	fmt.Printf("   Documentation saved to: %s\n", destOKFDir)
	return nil
}

func constructOKFPrompt(repoName string) string {
	var builder strings.Builder
	builder.WriteString("You are a senior software engineer and architect tasked with indexing the codebase in this repository and generating documentation conforming to the Open Knowledge Format (OKF) v0.1 specification.\n\n")
	builder.WriteString("### OKF v0.1 Requirements:\n")
	builder.WriteString("1. Create a directory named `OKF` at the root of the repository.\n")
	builder.WriteString("2. Every file inside the `OKF` directory must be a markdown file (`.md`) representing a single 'concept' (e.g. layers, specific flows, testing guidelines, entity models).\n")
	builder.WriteString("3. Every markdown file MUST start with a YAML frontmatter block containing these fields:\n")
	builder.WriteString("```yaml\n")
	builder.WriteString("---\n")
	builder.WriteString("type: <Concept Type (e.g., Codebase Overview, Architecture Design, Feature Flow, Developer Guide)>\n")
	builder.WriteString("title: <Descriptive Title>\n")
	builder.WriteString("description: <Short explanation of what this concept captures>\n")
	builder.WriteString("resource: <Relative path to directory or solution file it documents>\n")
	builder.WriteString("tags: [<array of relevant tags>]\n")
	builder.WriteString("timestamp: 2026-06-29T00:00:00Z\n")
	builder.WriteString("---\n")
	builder.WriteString("```\n")
	builder.WriteString("4. Cross-link between concept files using standard markdown links (e.g., `[Architecture Layers](architecture.md)` or `[Basket Flow](basket_flow.md)`).\n\n")
	
	builder.WriteString("### Mandatory Files to Generate:\n")
	builder.WriteString("- **`OKF/index.md`**: Main landing index page linking to all other documentation concepts and explaining key entry points.\n")
	builder.WriteString("- **`OKF/architecture.md`**: Detailed breakdown of the layers (e.g. Presentation, Core domain, Infrastructure) and core architectural design patterns (Repositories, Dependency Injection, Specification patterns).\n")
	builder.WriteString("- **`OKF/testing.md`**: Accurate compilation/build instructions, listing solution files (prefer solution files that build without external dependencies like Docker compose), and instructions on executing unit/functional/integration test suites.\n")
	builder.WriteString("- **Feature-specific mapping files**: Identify the core flows in the codebase (e.g. Basket/Shopping Cart flow, Order/Checkout flow, Authentication, or main business features) and create detailed files for them (e.g., `OKF/basket_flow.md`).\n\n")

	builder.WriteString("### Methodology:\n")
	builder.WriteString("- Take your time. Explore the directory structure carefully.\n")
	builder.WriteString("- Locate and read key project files, program entry points, configuration setups, interface files, and test files.\n")
	builder.WriteString("- Make sure your descriptions of directories, controllers, services, database contexts, and test scopes are highly accurate. Do not guess or hallucinate components.\n")
	builder.WriteString("- Write the complete set of markdown files inside the `OKF/` directory before exiting. Do not output code modifications to other files, only write files inside the new `OKF/` directory.\n")
	return builder.String()
}

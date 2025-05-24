package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sst/opencode/internal/app"
	"github.com/sst/opencode/internal/db"
	"github.com/sst/opencode/internal/permission"
)

func handleWatchMode(ctx context.Context, cwd string) error {
	slog.Info("Starting watch mode", "directory", cwd)

	conn, err := db.Connect()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	app, err := app.New(ctx, conn)
	if err != nil {
		slog.Error("Failed to create app", "error", err)
		return err
	}
	defer app.Shutdown()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer watcher.Close()

	err = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := filepath.Base(path)
			if strings.HasPrefix(name, ".") ||
				name == "node_modules" ||
				name == "target" ||
				name == "build" ||
				name == "dist" {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to add directories to watcher: %w", err)
	}

	slog.Info("File watcher started, watching for comments ending with 'opencode!'")
	fmt.Println("Watching for file changes... Press Ctrl+C to stop")

	fileDebounce := make(map[string]time.Time)
	var debounceMutex sync.Mutex

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				if info, err := os.Stat(event.Name); err == nil && !info.IsDir() {
					debounceMutex.Lock()
					lastProcessed, exists := fileDebounce[event.Name]
					now := time.Now()

					if !exists || now.Sub(lastProcessed) > 100*time.Millisecond {
						fileDebounce[event.Name] = now
						debounceMutex.Unlock()

						slog.Debug("Processing file for opencode comments", "file", event.Name)
						go processFileForOpenCodeComments(ctx, app, event.Name)
					} else {
						debounceMutex.Unlock()
						slog.Debug("Skipping file due to debounce", "file", event.Name, "lastProcessed", lastProcessed)
					}
				}
			}

			if event.Op&fsnotify.Create == fsnotify.Create {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					name := filepath.Base(event.Name)
					if !strings.HasPrefix(name, ".") &&
						name != "node_modules" &&
						name != "target" &&
						name != "build" &&
						name != "dist" {
						watcher.Add(event.Name)
					}
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			slog.Error("File watcher error", "error", err)
		}
	}
}

func processFileForOpenCodeComments(ctx context.Context, app *app.App, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		slog.Error("Failed to open file", "file", filePath, "error", err)
		return
	}
	defer file.Close()

	commentRegex := regexp.MustCompile(`(?m)^\s*(?://|#|/\*|\*|<!--)\s*(.+?)\s*opencode!\s*(?:\*/|-->)?\s*$`)
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		matches := commentRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			comment := strings.TrimSpace(matches[1])
			if comment != "" {
				slog.Info("Found opencode comment", "file", filePath, "line", lineNum, "comment", comment)

				fmt.Println("\nUser: " + strings.TrimSpace(comment))
				fmt.Printf("Processing: %s\nj", comment)
				err := processOpenCodeComment(ctx, app, comment, filePath, lineNum)

				if err != nil {
					slog.Error("Failed to process opencode comment", "error", err, "file", filePath, "line", lineNum)
					fmt.Printf("Error processing comment: %v\n", err)
				} else {
					fmt.Printf("Completed processing comment from %s:%d\n", filepath.Base(filePath), lineNum)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		slog.Error("Error reading file", "file", filePath, "error", err)
	}
}

func processOpenCodeComment(ctx context.Context, app *app.App, comment, filePath string, lineNum int) error {
	if app.CurrentSession == nil || app.CurrentSession.ID == "" {
		sessionTitle := fmt.Sprintf("Watch: %s:%d", filepath.Base(filePath), lineNum)
		session, err := app.Sessions.Create(ctx, sessionTitle)
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}

		permission.AutoApproveSession(ctx, session.ID)
		app.CurrentSession = &session
	}
	promptWithContext := fmt.Sprintf("File: %s (line %d)\nRequest: %s", filePath, lineNum, comment)

	err := removeOpenCodeComment(filePath, lineNum)
	if err != nil {
		slog.Warn("Failed to remove opencode comment line", "error", err, "file", filePath, "line", lineNum)
	}
	eventCh, err := app.PrimaryAgent.Run(ctx, app.CurrentSession.ID, promptWithContext, true)
	if err != nil {
		return fmt.Errorf("failed to run agent: %w", err)
	}

	for event := range eventCh {
		if event.Err() != nil {
			return fmt.Errorf("agent error: %w", event.Err())
		}
	}

	return nil
}

func removeOpenCodeComment(filePath string, lineNum int) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	if lineNum < 1 || lineNum > len(lines) {
		return fmt.Errorf("invalid line number: %d", lineNum)
	}

	lineIndex := lineNum - 1
	newLines := slices.Delete(lines, lineIndex, lineIndex+1)

	newContent := strings.Join(newLines, "\n")
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	slog.Debug("Removed opencode comment line", "file", filePath, "line", lineNum)
	return nil
}

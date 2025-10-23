package utils

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

const (
	defaultDirPermission  = 0755
	defaultFilePermission = 0644
	rootDir               = "."
	staticFilesDir        = "staticfiles"
)

//go:embed staticfiles
var embeddedFiles embed.FS

// FileSystemManager handles the extraction of embedded files to the filesystem
type FileSystemManager struct {
	embeddedFS embed.FS
	targetDir  string
}

// NewFileSystemManager creates a new FileSystemManager instance
func NewFileSystemManager() *FileSystemManager {
	return &FileSystemManager{
		embeddedFS: embeddedFiles,
		targetDir:  staticFilesDir,
	}
}

// ExtractEmbeddedFiles extracts all embedded files to the target directory
func (fsm *FileSystemManager) ExtractEmbeddedFiles() error {
	log.Printf("Starting extraction of embedded files to %s directory", fsm.targetDir)

	if err := fsm.ensureTargetDirectory(); err != nil {
		return fmt.Errorf("failed to ensure target directory: %w", err)
	}

	if err := fsm.walkAndExtractFiles(); err != nil {
		return fmt.Errorf("failed to extract embedded files: %w", err)
	}

	log.Printf("Successfully extracted all embedded files to %s directory", fsm.targetDir)
	return nil
}

// ensureTargetDirectory creates the target directory if it doesn't exist
func (fsm *FileSystemManager) ensureTargetDirectory() error {
	if err := os.MkdirAll(fsm.targetDir, defaultDirPermission); err != nil {
		return fmt.Errorf("failed to create target directory %s: %w", fsm.targetDir, err)
	}
	return nil
}

// walkAndExtractFiles walks through the embedded filesystem and extracts files
func (fsm *FileSystemManager) walkAndExtractFiles() error {
	return fs.WalkDir(fsm.embeddedFS, rootDir, fsm.processFileEntry)
}

func (fsm *FileSystemManager) processFileEntry(path string, entry fs.DirEntry, err error) error {
	if err != nil {
		return fmt.Errorf("error accessing path %s: %w", path, err)
	}

	// Skip root directory
	if path == rootDir {
		return nil
	}

	// Skip the staticfiles directory itself (we only want the files inside it)
	if entry.IsDir() && path == "staticfiles" {
		return nil
	}

	// Extract the filename from the path (remove the staticfiles/ prefix)
	fileName := filepath.Base(path)
	destPath := filepath.Join(fsm.targetDir, fileName)

	if entry.IsDir() {
		return fsm.createDirectory(destPath)
	}

	return fsm.extractFile(path, destPath)
}

func (fsm *FileSystemManager) createDirectory(dirPath string) error {
	if err := os.MkdirAll(dirPath, defaultDirPermission); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
	}
	log.Printf("Created directory: %s", dirPath)
	return nil
}

func (fsm *FileSystemManager) extractFile(sourcePath, destPath string) error {
	content, err := fs.ReadFile(fsm.embeddedFS, sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read embedded file %s: %w", sourcePath, err)
	}

	if err := fsm.ensureParentDirectory(destPath); err != nil {
		return err
	}

	if err := os.WriteFile(destPath, content, defaultFilePermission); err != nil {
		return fmt.Errorf("failed to write file %s: %w", destPath, err)
	}

	log.Printf("Extracted file: %s", destPath)
	return nil
}

func (fsm *FileSystemManager) ensureParentDirectory(filePath string) error {
	parentDir := filepath.Dir(filePath)
	if parentDir != rootDir && parentDir != fsm.targetDir {
		if err := os.MkdirAll(parentDir, defaultDirPermission); err != nil {
			return fmt.Errorf("failed to create parent directory %s: %w", parentDir, err)
		}
	}
	return nil
}

// ExtractStaticFiles extracts embedded static files to the filesystem
// This function is called during service initialization to ensure required files are available
func ExtractStaticFiles() error {
	fsm := NewFileSystemManager()
	return fsm.ExtractEmbeddedFiles()
}

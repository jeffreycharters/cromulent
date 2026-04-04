package handlers

import (
	"context"
	"cromulent/config"
	"cromulent/db"
	"fmt"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type ConfigHandler struct {
	ctx context.Context
}

func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{}
}

// SetContext is called from app.startup so the handler can use Wails runtime
// functions (file dialogs) that require the Wails context.
func (h *ConfigHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

// GetDBPath returns the currently configured DB path, or "" if not set.
func (h *ConfigHandler) GetDBPath() (string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", err
	}
	return cfg.DBPath, nil
}

// InitDB initialises the database at the path stored in config.
// Called by the frontend on startup once it knows a path is configured.
func (h *ConfigHandler) InitDB() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.DBPath == "" {
		return fmt.Errorf("no database path configured")
	}
	return db.Init(cfg.DBPath)
}

// OpenDBFilePicker opens a native file dialog and returns the chosen path.
// It does not save anything — the frontend drives the confirmation flow.
func (h *ConfigHandler) OpenDBFilePicker() (string, error) {
	if h.ctx == nil {
		return "", fmt.Errorf("context not ready")
	}
	path, err := runtime.OpenFileDialog(h.ctx, runtime.OpenDialogOptions{
		Title: "Select Cromulent database file",
		Filters: []runtime.FileFilter{
			{DisplayName: "SQLite Database (*.db)", Pattern: "*.db"},
		},
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

// SetDBPath saves path to config and reinitialises the DB.
// Logout is handled by the frontend before calling this.
func (h *ConfigHandler) SetDBPath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	cfg := &config.Config{DBPath: path}
	if err := config.Save(cfg); err != nil {
		return err
	}
	return db.Init(path)
}

func (h *ConfigHandler) OpenDBFolderPicker() (string, error) {
	if h.ctx == nil {
		return "", fmt.Errorf("context not ready")
	}
	folder, err := runtime.OpenDirectoryDialog(h.ctx, runtime.OpenDialogOptions{
		Title: "Choose location for new database",
	})
	if err != nil {
		return "", err
	}
	if folder == "" {
		return "", nil
	}
	return filepath.Join(folder, "cromulent.db"), nil
}

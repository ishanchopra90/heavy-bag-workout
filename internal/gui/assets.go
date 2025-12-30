package gui

import (
	"fmt"
	"heavybagworkout/internal/models"
	"image"
	_ "image/jpeg" // Register JPEG decoder
	_ "image/png"  // Register PNG decoder
	"os"
	"path/filepath"
)

// AssetLoader handles loading sprite images for animations
type AssetLoader struct {
	assetDir string // Base directory for assets (e.g., "assets" or "internal/gui/assets")
}

// NewAssetLoader creates a new asset loader with the specified asset directory
func NewAssetLoader(assetDir string) *AssetLoader {
	return &AssetLoader{
		assetDir: assetDir,
	}
}

// LoadSpriteImage loads a sprite image from the file system
// Tries JPG first, then falls back to PNG if JPG is not found
// Returns nil if the image cannot be loaded
func (al *AssetLoader) LoadSpriteImage(stance models.Stance, moveName string) (image.Image, error) {
	stanceDir := stance.String()

	// Try JPG first (preferred format)
	jpgFileName := fmt.Sprintf("%s.jpg", moveName)
	jpgFilePath := filepath.Join(al.assetDir, stanceDir, jpgFileName)

	file, err := os.Open(jpgFilePath)
	if err == nil {
		defer file.Close()
		img, _, decodeErr := image.Decode(file)
		if decodeErr == nil {
			return img, nil
		}
		// If decode failed, continue to try PNG
	}

	// Fallback to PNG
	pngFileName := fmt.Sprintf("%s.png", moveName)
	pngFilePath := filepath.Join(al.assetDir, stanceDir, pngFileName)

	file, err = os.Open(pngFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sprite file (tried %s and %s): %w", jpgFilePath, pngFilePath, err)
	}
	defer file.Close()

	// Decode the PNG image
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode sprite image %s: %w", pngFilePath, err)
	}

	return img, nil
}

// GetMoveFileName returns the file name for a given punch or defensive move
func GetMoveFileName(move interface{}) string {
	switch v := move.(type) {
	case models.Punch:
		switch v {
		case models.Jab:
			return "jab"
		case models.Cross:
			return "cross"
		case models.LeadHook:
			return "left_hook"
		case models.RearHook:
			return "right_hook"
		case models.LeadUppercut:
			return "left_uppercut"
		case models.RearUppercut:
			return "right_uppercut"
		}
	case models.DefensiveMove:
		switch v {
		case models.LeftSlip:
			return "left_slip"
		case models.RightSlip:
			return "right_slip"
		case models.LeftRoll:
			return "left_roll"
		case models.RightRoll:
			return "right_roll"
		case models.PullBack:
			return "pull_back"
		case models.Duck:
			return "duck"
		}
	}
	return ""
}

// LoadIdleAnimation loads idle animation frames
// For now, we'll use a single frame, but this can be extended to multiple frames
// Tries idle.jpg/idle.png first, then falls back to jab.jpg/jab.png
func (al *AssetLoader) LoadIdleAnimation(stance models.Stance) ([]image.Image, error) {
	// Try to load idle (will try .jpg first, then .png)
	img, err := al.LoadSpriteImage(stance, "idle")
	if err != nil {
		// Fallback to jab as a default pose (will try .jpg first, then .png)
		img, err = al.LoadSpriteImage(stance, "jab")
		if err != nil {
			return nil, err
		}
	}
	return []image.Image{img}, nil
}

// GetPunchFileName returns the file name for a punch based on stance
// File names are: jab.png, cross.png, left_hook.png, right_hook.png, left_uppercut.png, right_uppercut.png
// The mapping depends on stance (orthodox vs southpaw)
func GetPunchFileName(punch models.Punch, stance models.Stance) string {
	switch punch {
	case models.Jab:
		return "jab"
	case models.Cross:
		return "cross"
	case models.LeadHook:
		if stance == models.Southpaw {
			return "right_hook" // In southpaw, lead is right
		}
		return "left_hook" // In orthodox, lead is left
	case models.RearHook:
		if stance == models.Southpaw {
			return "left_hook" // In southpaw, rear is left
		}
		return "right_hook" // In orthodox, rear is right
	case models.LeadUppercut:
		if stance == models.Southpaw {
			return "right_uppercut"
		}
		return "left_uppercut"
	case models.RearUppercut:
		if stance == models.Southpaw {
			return "left_uppercut"
		}
		return "right_uppercut"
	}
	return ""
}

// GetDefensiveMoveFileName returns the file name for a defensive move
func GetDefensiveMoveFileName(move models.DefensiveMove) string {
	switch move {
	case models.LeftSlip:
		return "left_slip"
	case models.RightSlip:
		return "right_slip"
	case models.LeftRoll:
		return "left_roll"
	case models.RightRoll:
		return "right_roll"
	case models.PullBack:
		return "pull_back"
	case models.Duck:
		return "duck"
	}
	return ""
}

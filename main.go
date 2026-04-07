package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "Cromulent",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 228, G: 228, B: 228, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
			app.Auth,
			app.Setup,
			app.Config,
			app.Library,
			app.MMA,
			app.DataEntry,
			app.Limits,
			app.ChartReview,
			app.Comment,
			app.SPCRuleSet,
		},
	})
	if err != nil {
		log.Fatalf("wails: %v", err)
	}
}

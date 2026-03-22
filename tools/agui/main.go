package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:             "AG Studio",
		Width:             1400,
		Height:            860,
		MinWidth:          900,
		MinHeight:         600,
		BackgroundColour:  &options.RGBA{R: 3, G: 7, B: 18, A: 1}, // gray-950
		AssetServer:       &assetserver.Options{Assets: assets},
		OnStartup:         app.startup,
		Frameless:         false,
		EnableDefaultContextMenu: false,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

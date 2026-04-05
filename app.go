package main

import (
	"context"
	"cromulent/handlers"
)

type App struct {
	ctx         context.Context
	Auth        *handlers.AuthHandler
	Setup       *handlers.SetupHandler
	Config      *handlers.ConfigHandler
	Library     *handlers.LibraryHandler
	MMA         *handlers.MMAHandler
	DataEntry   *handlers.DataEntryHandler
	Limits      *handlers.LimitsHandler
	ChartReview *handlers.ChartReviewHandler
}

func NewApp() *App {
	return &App{
		Auth:        handlers.NewAuthHandler(),
		Setup:       handlers.NewSetupHandler(),
		Config:      handlers.NewConfigHandler(),
		Library:     handlers.NewLibraryHandler(),
		MMA:         handlers.NewMMAHandler(),
		DataEntry:   handlers.NewDataEntryHandler(),
		Limits:      handlers.NewLimitsHandler(),
		ChartReview: handlers.NewChartReviewHandler(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.Config.SetContext(ctx)
}

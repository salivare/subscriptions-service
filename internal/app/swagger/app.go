package swaggerapp

import (
	"net/http"
)

type App struct {
	jsonPath string
	uiPath   string
}

func New(jsonPath, uiPath string) *App {
	return &App{
		jsonPath: jsonPath,
		uiPath:   uiPath,
	}
}

func (a *App) Register(mux *http.ServeMux) {
	// swagger.json
	mux.HandleFunc(
		"/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, a.jsonPath)
		},
	)

	// Swagger UI
	fs := http.FileServer(http.Dir(a.uiPath))
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", fs))
}

package handler

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"time"

	"github.com/casapps/cassocial/src/server"
)

// templateCache holds parsed templates keyed by page name.
// Populated once by initTemplates() inside NewRouter.
var templateCache = map[string]*template.Template{}

// initTemplates parses all page templates together with shared layouts and partials.
func initTemplates() error {
	pages := []string{
		"home", "login", "register", "setup",
		"dashboard", "admin", "profile",
	}

	shared := []string{
		"template/layout/base.html",
		"template/partial/header.html",
		"template/partial/footer.html",
		"template/partial/link_card.html",
	}

	for _, page := range pages {
		pagePath := fmt.Sprintf("template/page/%s.html", page)
		files := append(shared, pagePath)

		tmpl, err := template.New("base").ParseFS(server.TemplateFS, files...)
		if err != nil {
			return fmt.Errorf("parse template %s: %w", page, err)
		}
		templateCache[page] = tmpl
	}

	return nil
}

// renderTemplate executes a named page template and writes the HTML response.
func renderTemplate(w http.ResponseWriter, page string, data interface{}) {
	tmpl, ok := templateCache[page]
	if !ok {
		http.Error(w, "template not found: "+page, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "template render error", http.StatusInternalServerError)
	}
}

// BaseTemplateData carries fields common to every page template.
type BaseTemplateData struct {
	SiteName            string
	AppVersion          string
	CurrentYear         int
	MetaTitle           string
	MetaDescription     string
	OgImageURL          string
	IsAuthenticated     bool
	IsAdmin             bool
	RegistrationEnabled bool
}

// newBaseData populates BaseTemplateData using DB settings and the current request context.
func newBaseData(rt *Router, r *http.Request) BaseTemplateData {
	siteName := "Cassocial"
	if v, err := rt.db.GetSetting("site_name"); err == nil && v != "" {
		siteName = v
	}

	regEnabled := true
	if v, err := rt.db.GetSetting("registration_enabled"); err == nil && v != "" {
		regEnabled = v != "false"
	}

	isAuth := false
	isAdmin := false
	if tok := rt.middleware.ExtractToken(r); tok != "" {
		if claims, err := rt.authHandlers.auth.ValidateToken(tok); err == nil {
			isAuth = true
			isAdmin = claims.Role == "admin"
		}
	}

	return BaseTemplateData{
		SiteName:            siteName,
		AppVersion:          appVersionString,
		CurrentYear:         time.Now().Year(),
		IsAuthenticated:     isAuth,
		IsAdmin:             isAdmin,
		RegistrationEnabled: regEnabled,
	}
}

// appVersionString is populated by SetAppVersion from main.
var appVersionString = "dev"

// SetAppVersion lets main.go propagate the build-time version string into templates.
func SetAppVersion(v string) {
	appVersionString = v
}

// staticHandler returns an http.Handler that serves embedded static assets.
// Routes to this handler should be prefixed with /static/.
func staticHandler() http.Handler {
	sub, err := fs.Sub(server.StaticFS, "static")
	if err != nil {
		panic("failed to sub static FS: " + err.Error())
	}
	return http.StripPrefix("/static/", http.FileServer(http.FS(sub)))
}

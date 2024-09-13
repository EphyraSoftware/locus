package auth

import (
	"context"
	"embed"
	"fmt"
	ory "github.com/ory/client-go"
	"html/template"
	"log"
	"net/http"
	"os"
)

//go:embed public/*
var public embed.FS

type App struct {
	ory     *ory.APIClient
	oryBase string
}

func StartApp(oryClient *ory.APIClient, oryBase string) error {
	app := &App{
		ory:     oryClient,
		oryBase: oryBase,
	}
	mux := http.NewServeMux()

	mux.Handle("/login", app.loginHandler())
	mux.Handle("/register", app.registerHandler())
	mux.Handle("/verification", app.verificationHandler())
	mux.Handle("/", app.sessionMiddleware(app.dashboardHandler()))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3100"
	}

	fmt.Printf("Application launched and running on http://127.0.0.1:%s\n", port)
	// start the server
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		return err
	}

	return nil
}

// save the cookies for any upstream calls to the Ory apis
func withCookies(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, "req.cookies", v)
}

func getCookies(ctx context.Context) string {
	return ctx.Value("req.cookies").(string)
}

// save the session to display it on the dashboard
func withSession(ctx context.Context, v *ory.Session) context.Context {
	return context.WithValue(ctx, "req.session", v)
}

func getSession(ctx context.Context) *ory.Session {
	return ctx.Value("req.session").(*ory.Session)
}

func (app *App) sessionMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		log.Printf("handling middleware request\n")

		// set the cookies on the ory client
		var cookies string

		// this example passes all request.Cookies
		// to `ToSession` function
		//
		// However, you can pass only the value of
		// ory_session_projectid cookie to the endpoint
		cookies = request.Header.Get("Cookie")

		// check if we have a session
		session, _, err := app.ory.FrontendAPI.ToSession(request.Context()).Cookie(cookies).Execute()
		if (err != nil && session == nil) || (err == nil && !*session.Active) {
			// this will redirect the user to the managed Ory Login UI
			http.Redirect(writer, request, fmt.Sprintf("%sself-service/login/browser", app.oryBase), http.StatusSeeOther)
			return
		}

		ctx := withCookies(request.Context(), cookies)
		ctx = withSession(ctx, session)

		// continue to the requested page (in our case the Dashboard)
		next.ServeHTTP(writer, request.WithContext(ctx))
		return
	}
}

type DashboardData struct {
	LogoutUrl string
	Session   *ory.Session
}

func (app *App) dashboardHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		tmpl, err := template.New("index.html").ParseFS(public, "public/index.html")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		cookies := getCookies(request.Context())
		req := app.ory.FrontendAPI.CreateBrowserLogoutFlow(request.Context()).Cookie(cookies)
		url, _, err := app.ory.FrontendAPI.CreateBrowserLogoutFlowExecute(req)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		session := getSession(request.Context())
		if session == nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		err = tmpl.ExecuteTemplate(writer, "index.html", DashboardData{
			LogoutUrl: url.LogoutUrl,
			Session:   session,
		})
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

type LoginData struct {
	Flow        *ory.LoginFlow
	RegisterUrl string
}

func (app *App) loginHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		flowId := request.URL.Query().Get("flow")
		if flowId == "" {
			http.Error(writer, "missing flow id", http.StatusBadRequest)
			return
		}

		cookies := request.Header.Get("Cookie")

		req := app.ory.FrontendAPI.GetLoginFlow(request.Context()).Id(flowId).Cookie(cookies)
		flow, _, err := app.ory.FrontendAPI.GetLoginFlowExecute(req)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		//json, err := json.Marshal(flow)
		//log.Printf("flow: %s\n", string(json))

		tmpl, err := template.New("login.html").ParseFS(public, "public/login.html")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		err = tmpl.ExecuteTemplate(writer, "login.html", LoginData{
			Flow:        flow,
			RegisterUrl: fmt.Sprintf("%sself-service/registration/browser", app.oryBase),
		})
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (app *App) registerHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		flowId := request.URL.Query().Get("flow")
		if flowId == "" {
			http.Error(writer, "missing flow id", http.StatusBadRequest)
			return
		}

		cookies := request.Header.Get("Cookie")

		req := app.ory.FrontendAPI.GetRegistrationFlow(request.Context()).Id(flowId).Cookie(cookies)
		flow, _, err := app.ory.FrontendAPI.GetRegistrationFlowExecute(req)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl, err := template.New("register.html").ParseFS(public, "public/register.html")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		err = tmpl.ExecuteTemplate(writer, "register.html", flow)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (app *App) verificationHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		flowId := request.URL.Query().Get("flow")
		if flowId == "" {
			http.Error(writer, "missing flow id", http.StatusBadRequest)
			return
		}

		cookies := request.Header.Get("Cookie")

		req := app.ory.FrontendAPI.GetVerificationFlow(request.Context()).Id(flowId).Cookie(cookies)
		flow, _, err := app.ory.FrontendAPI.GetVerificationFlowExecute(req)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl, err := template.New("verification.html").ParseFS(public, "public/verification.html")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		err = tmpl.ExecuteTemplate(writer, "verification.html", flow)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

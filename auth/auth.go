package auth

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	ory "github.com/ory/client-go"
	"net/http"
)

type App struct {
	Router  fiber.Router
	Ory     *ory.APIClient
	OryBase string
}

func SessionUser(c *fiber.Ctx) interface{} {
	session := c.Locals("session").(*ory.Session)
	if session == nil {
		return nil
	} else {
		return session.Identity.Traits
	}
}

func SessionId(c *fiber.Ctx) string {
	session := c.Locals("session").(*ory.Session)
	if session == nil {
		return ""
	} else {
		return session.Identity.Id
	}
}

func (a *App) Prepare(app *fiber.App) {
	a.Router.Get("/login", func(c *fiber.Ctx) error {
		flowId := c.Query("flow")
		if flowId == "" {
			return c.SendStatus(http.StatusBadRequest)
		}

		cookies, err := getCookiesFromRequest(c)
		if err != nil {
			return c.SendStatus(http.StatusBadRequest)
		}

		req := a.Ory.FrontendAPI.GetLoginFlow(c.Context()).Id(flowId).Cookie(cookies)
		flow, _, err := a.Ory.FrontendAPI.GetLoginFlowExecute(req)
		if err != nil {
			return c.SendStatus(http.StatusInternalServerError)
		}

		return c.Render("public/auth/login", fiber.Map{
			"Flow":        flow,
			"RegisterUrl": fmt.Sprintf("%sself-service/registration/browser", a.OryBase),
		})
	})

	a.Router.Get("/register", func(c *fiber.Ctx) error {
		flowId := c.Query("flow")
		if flowId == "" {
			return c.SendStatus(http.StatusBadRequest)
		}

		cookies, err := getCookiesFromRequest(c)
		if err != nil {
			return c.SendStatus(http.StatusBadRequest)
		}

		req := a.Ory.FrontendAPI.GetRegistrationFlow(c.Context()).Id(flowId).Cookie(cookies)
		flow, _, err := a.Ory.FrontendAPI.GetRegistrationFlowExecute(req)
		if err != nil {
			return c.SendStatus(http.StatusInternalServerError)
		}

		return c.Render("public/auth/register", flow)
	})

	a.Router.Get("/verification", func(c *fiber.Ctx) error {
		flowId := c.Query("flow")
		if flowId == "" {
			return c.SendStatus(http.StatusBadRequest)
		}

		cookies, err := getCookiesFromRequest(c)
		if err != nil {
			return c.SendStatus(http.StatusBadRequest)
		}

		req := a.Ory.FrontendAPI.GetVerificationFlow(c.Context()).Id(flowId).Cookie(cookies)
		flow, _, err := a.Ory.FrontendAPI.GetVerificationFlowExecute(req)
		if err != nil {
			return c.SendStatus(http.StatusInternalServerError)
		}

		return c.Render("public/auth/verification", flow)
	})

	// Mount middleware to the root of the app to protect all routes
	app.Use(func(c *fiber.Ctx) error {
		cookies := c.GetReqHeaders()["Cookie"]
		if len(cookies) == 0 {
			return c.Redirect(fmt.Sprintf("%sself-service/login/browser", a.OryBase), http.StatusSeeOther)
		}

		// check if we have a session
		session, _, err := a.Ory.FrontendAPI.ToSession(c.Context()).Cookie(cookies[0]).Execute()
		if (err != nil && session == nil) || (err == nil && !*session.Active) {
			// this will redirect the user to the managed Ory Login UI
			return c.Redirect(fmt.Sprintf("%sself-service/login/browser", a.OryBase), http.StatusSeeOther)
		}

		c.Locals("cookies", cookies[0])
		c.Locals("session", session)

		return c.Next()
	})

	a.Router.Get("/logout", func(c *fiber.Ctx) error {
		cookie := c.Locals("cookies").(string)
		req := a.Ory.FrontendAPI.CreateBrowserLogoutFlow(c.Context()).Cookie(cookie)
		url, _, err := a.Ory.FrontendAPI.CreateBrowserLogoutFlowExecute(req)
		if err != nil {
			return c.SendStatus(http.StatusInternalServerError)
		}

		return c.Redirect(url.LogoutUrl, http.StatusSeeOther)
	})
}

func getCookiesFromRequest(c *fiber.Ctx) (string, error) {
	var cookie string
	cookies := c.GetReqHeaders()["Cookie"]
	if len(cookies) > 0 {
		cookie = cookies[0]
	} else {
		return "", fmt.Errorf("missing cookies")
	}
	return cookie, nil
}

//func StartApp(oryClient *ory.APIClient, oryBase string) error {
//	app := &App{
//		Ory:     oryClient,
//		OryBase: oryBase,
//	}
//	mux := http.NewServeMux()
//
//	mux.Handle("/login", app.loginHandler())
//	mux.Handle("/register", app.registerHandler())
//	mux.Handle("/verification", app.verificationHandler())
//	mux.Handle("/", app.sessionMiddleware(app.dashboardHandler()))
//
//	port := os.Getenv("PORT")
//	if port == "" {
//		port = "3100"
//	}
//
//	fmt.Printf("Application launched and running on http://127.0.0.1:%s\n", port)
//	// start the server
//	err := http.ListenAndServe(":"+port, mux)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

// save the cookies for any upstream calls to the Ory apis
func withCookies(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, "req.cookies", v)
}

func getCookies(ctx context.Context) string {
	return ctx.Value("req.cookies").(string)
}

//func (app *App) sessionMiddleware(next http.HandlerFunc) http.HandlerFunc {
//	return func(writer http.ResponseWriter, request *http.Request) {
//		log.Printf("handling middleware request\n")
//
//		// set the cookies on the ory client
//		var cookies string
//
//		// this example passes all request.Cookies
//		// to `ToSession` function
//		//
//		// However, you can pass only the value of
//		// ory_session_projectid cookie to the endpoint
//		cookies = request.Header.Get("Cookie")
//
//		// check if we have a session
//		session, _, err := app.Ory.FrontendAPI.ToSession(request.Context()).Cookie(cookies).Execute()
//		if (err != nil && session == nil) || (err == nil && !*session.Active) {
//			// this will redirect the user to the managed Ory Login UI
//			http.Redirect(writer, request, fmt.Sprintf("%sself-service/login/browser", app.OryBase), http.StatusSeeOther)
//			return
//		}
//
//		ctx := withCookies(request.Context(), cookies)
//		ctx = withSession(ctx, session)
//
//		// continue to the requested page (in our case the Dashboard)
//		next.ServeHTTP(writer, request.WithContext(ctx))
//		return
//	}
//}

type DashboardData struct {
	LogoutUrl string
	Session   *ory.Session
}

//func (app *App) dashboardHandler() http.HandlerFunc {
//	return func(writer http.ResponseWriter, request *http.Request) {
//		tmpl, err := template.New("index.html").ParseFS(public, "public/index.html")
//		if err != nil {
//			http.Error(writer, err.Error(), http.StatusInternalServerError)
//			return
//		}
//
//		cookies := getCookies(request.Context())
//		req := app.Ory.FrontendAPI.CreateBrowserLogoutFlow(request.Context()).Cookie(cookies)
//		url, _, err := app.Ory.FrontendAPI.CreateBrowserLogoutFlowExecute(req)
//		if err != nil {
//			http.Error(writer, err.Error(), http.StatusInternalServerError)
//			return
//		}
//
//		session := getSession(request.Context())
//		if session == nil {
//			http.Error(writer, err.Error(), http.StatusInternalServerError)
//			return
//		}
//		err = tmpl.ExecuteTemplate(writer, "index.html", DashboardData{
//			LogoutUrl: url.LogoutUrl,
//			Session:   session,
//		})
//		if err != nil {
//			http.Error(writer, err.Error(), http.StatusInternalServerError)
//			return
//		}
//	}
//}

type LoginData struct {
	Flow        *ory.LoginFlow
	RegisterUrl string
}

//func (app *App) loginHandler() http.HandlerFunc {
//	return func(writer http.ResponseWriter, request *http.Request) {
//		flowId := request.URL.Query().Get("flow")
//		if flowId == "" {
//			http.Error(writer, "missing flow id", http.StatusBadRequest)
//			return
//		}
//
//		cookies := request.Header.Get("Cookie")
//
//		req := app.Ory.FrontendAPI.GetLoginFlow(request.Context()).Id(flowId).Cookie(cookies)
//		flow, _, err := app.Ory.FrontendAPI.GetLoginFlowExecute(req)
//		if err != nil {
//			http.Error(writer, err.Error(), http.StatusInternalServerError)
//			return
//		}
//
//		//json, err := json.Marshal(flow)
//		//log.Printf("flow: %s\n", string(json))
//
//		tmpl, err := template.New("login.html").ParseFS(public, "public/login.html")
//		if err != nil {
//			http.Error(writer, err.Error(), http.StatusInternalServerError)
//			return
//		}
//		err = tmpl.ExecuteTemplate(writer, "login.html", LoginData{
//			Flow:        flow,
//			RegisterUrl: fmt.Sprintf("%sself-service/registration/browser", app.OryBase),
//		})
//		if err != nil {
//			http.Error(writer, err.Error(), http.StatusInternalServerError)
//			return
//		}
//	}
//}

//func (app *App) registerHandler() http.HandlerFunc {
//	return func(writer http.ResponseWriter, request *http.Request) {
//		flowId := request.URL.Query().Get("flow")
//		if flowId == "" {
//			http.Error(writer, "missing flow id", http.StatusBadRequest)
//			return
//		}
//
//		cookies := request.Header.Get("Cookie")
//
//		req := app.Ory.FrontendAPI.GetRegistrationFlow(request.Context()).Id(flowId).Cookie(cookies)
//		flow, _, err := app.Ory.FrontendAPI.GetRegistrationFlowExecute(req)
//		if err != nil {
//			http.Error(writer, err.Error(), http.StatusInternalServerError)
//			return
//		}
//
//		tmpl, err := template.New("register.html").ParseFS(public, "public/register.html")
//		if err != nil {
//			http.Error(writer, err.Error(), http.StatusInternalServerError)
//			return
//		}
//		err = tmpl.ExecuteTemplate(writer, "register.html", flow)
//		if err != nil {
//			http.Error(writer, err.Error(), http.StatusInternalServerError)
//			return
//		}
//	}
//}

//func (app *App) verificationHandler() http.HandlerFunc {
//	return func(writer http.ResponseWriter, request *http.Request) {
//		flowId := request.URL.Query().Get("flow")
//		if flowId == "" {
//			http.Error(writer, "missing flow id", http.StatusBadRequest)
//			return
//		}
//
//		cookies := request.Header.Get("Cookie")
//
//		req := app.Ory.FrontendAPI.GetVerificationFlow(request.Context()).Id(flowId).Cookie(cookies)
//		flow, _, err := app.Ory.FrontendAPI.GetVerificationFlowExecute(req)
//		if err != nil {
//			http.Error(writer, err.Error(), http.StatusInternalServerError)
//			return
//		}
//
//		tmpl, err := template.New("verification.html").ParseFS(public, "public/verification.html")
//		if err != nil {
//			http.Error(writer, err.Error(), http.StatusInternalServerError)
//			return
//		}
//		err = tmpl.ExecuteTemplate(writer, "verification.html", flow)
//		if err != nil {
//			http.Error(writer, err.Error(), http.StatusInternalServerError)
//			return
//		}
//	}
//}

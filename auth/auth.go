package auth

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	ory "github.com/ory/client-go"
	"net/http"
)

type App struct {
	Router         fiber.Router
	Ory            *ory.APIClient
	OryBrowserBase string
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
			log.Info("No flow id")
			return c.SendStatus(http.StatusBadRequest)
		}

		cookies, err := getCookiesFromRequest(c)
		if err != nil {
			log.Errorf("No cookies: %s", err)
			return c.SendStatus(http.StatusBadRequest)
		}

		req := a.Ory.FrontendAPI.GetLoginFlow(c.Context()).Id(flowId).Cookie(cookies)
		flow, _, err := a.Ory.FrontendAPI.GetLoginFlowExecute(req)
		if err != nil {
			log.Errorf("Error getting login flow: %s", err)
			return c.SendStatus(http.StatusInternalServerError)
		}

		return c.Render("public/auth/login", fiber.Map{
			"Flow":        flow,
			"RegisterUrl": fmt.Sprintf("%sself-service/registration/browser", a.OryBrowserBase),
		})
	})

	a.Router.Get("/register", func(c *fiber.Ctx) error {
		flowId := c.Query("flow")
		if flowId == "" {
			log.Info("No flow id")
			return c.SendStatus(http.StatusBadRequest)
		}

		cookies, err := getCookiesFromRequest(c)
		if err != nil {
			log.Errorf("No cookies: %s", err)
			return c.SendStatus(http.StatusBadRequest)
		}

		req := a.Ory.FrontendAPI.GetRegistrationFlow(c.Context()).Id(flowId).Cookie(cookies)
		flow, _, err := a.Ory.FrontendAPI.GetRegistrationFlowExecute(req)
		if err != nil {
			log.Errorf("Error getting registration flow: %s", err)
			return c.SendStatus(http.StatusInternalServerError)
		}

		return c.Render("public/auth/register", flow)
	})

	a.Router.Get("/verification", func(c *fiber.Ctx) error {
		flowId := c.Query("flow")
		if flowId == "" {
			log.Info("No flow id")
			return c.SendStatus(http.StatusBadRequest)
		}

		cookies, err := getCookiesFromRequest(c)
		if err != nil {
			log.Errorf("No cookies: %s", err)
			return c.SendStatus(http.StatusBadRequest)
		}

		req := a.Ory.FrontendAPI.GetVerificationFlow(c.Context()).Id(flowId).Cookie(cookies)
		flow, _, err := a.Ory.FrontendAPI.GetVerificationFlowExecute(req)
		if err != nil {
			log.Errorf("Error getting verification flow: %s", err)
			return c.SendStatus(http.StatusInternalServerError)
		}

		return c.Render("public/auth/verification", flow)
	})

	// Mount middleware to the root of the app to protect all routes
	app.Use(func(c *fiber.Ctx) error {
		cookies := c.GetReqHeaders()["Cookie"]
		if len(cookies) == 0 {
			return c.Redirect(fmt.Sprintf("%sself-service/login/browser", a.OryBrowserBase), http.StatusSeeOther)
		}

		// check if we have a session
		session, _, err := a.Ory.FrontendAPI.ToSession(c.Context()).Cookie(cookies[0]).Execute()
		if (err != nil && session == nil) || (err == nil && !*session.Active) {
			// this will redirect the user to the managed Ory Login UI
			return c.Redirect(fmt.Sprintf("%sself-service/login/browser", a.OryBrowserBase), http.StatusSeeOther)
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
			log.Errorf("Error creating logout flow: %s", err)
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

type DashboardData struct {
	LogoutUrl string
	Session   *ory.Session
}

type LoginData struct {
	Flow        *ory.LoginFlow
	RegisterUrl string
}

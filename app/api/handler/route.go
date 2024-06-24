package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/riandyrn/otelchi"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/Housiadas/backend-system/business/domain/authbus"
)

// Routes returns applications router
func (h *Handler) Routes() *chi.Mux {
	mid := h.Web.Mid

	authenticate := mid.Bearer()
	ruleAny := mid.Authorize(authbus.RuleAny)
	ruleUserOnly := mid.Authorize(authbus.RuleUserOnly)
	ruleAdmin := mid.Authorize(authbus.RuleAdminOnly)

	ruleAuthorizeUser := mid.AuthorizeUser(authbus.RuleAdminOrSubject)
	ruleAuthorizeAdmin := mid.AuthorizeUser(authbus.RuleAdminOnly)
	ruleAuthorizeProduct := mid.AuthorizeProduct(authbus.RuleAdminOrSubject)

	apiRouter := chi.NewRouter()
	apiRouter.Use(
		mid.RequestID,
		middleware.SetHeader("Content-Type", "application/json"),
		middleware.GetHead,
		mid.Recoverer(),
	)

	// v1 routes
	apiRouter.Route("/v1", func(v1 chi.Router) {
		v1.Use(otelchi.Middleware(h.AppName, otelchi.WithChiRoutes(v1)))
		v1.Use(mid.ApiVersion("v1"))
		v1.Use(cors.Handler(cors.Options{
			AllowedOrigins: h.Cors.AllowedOrigins,
			AllowedMethods: h.Cors.AllowedMethods,
			AllowedHeaders: h.Cors.AllowedHeaders,
			ExposedHeaders: h.Cors.ExposedHeaders,
			MaxAge:         h.Cors.MaxAge,
		}))

		// Auth
		v1.Post("/auth/authenticate", h.Web.Respond.Respond(h.authenticate))
		v1.With(authenticate).Get("/auth/authorize", h.Web.Respond.Respond(h.authorize))

		// Users
		v1.With(authenticate).Route("/users", func(u chi.Router) {
			u.With(ruleAuthorizeAdmin).Get("/", h.Web.Respond.Respond(h.userQuery))
			u.With(ruleAuthorizeUser).Get("/{user_id}", h.Web.Respond.Respond(h.userQueryByID))
			u.With(ruleAdmin).Post("/users", h.Web.Respond.Respond(h.userCreate))
			u.With(ruleAdmin).Put("/role/{user_id}", h.Web.Respond.Respond(h.updateRole))
			u.With(ruleAuthorizeUser).Put("/{user_id}", h.Web.Respond.Respond(h.userUpdate))
			u.With(ruleAuthorizeUser).Delete("/{user_id}", h.Web.Respond.Respond(h.userDelete))
		})

		// Products
		v1.With(authenticate).Route("/products", func(p chi.Router) {
			p.With(ruleAny).Get("/", h.Web.Respond.Respond(h.productQuery))
			p.With(ruleUserOnly).Post("/", h.Web.Respond.Respond(h.productCreate))
			p.With(ruleAuthorizeProduct).Get("/{product_id}", h.Web.Respond.Respond(h.productQueryByID))
			p.With(ruleAuthorizeProduct).Put("/{product_id}", h.Web.Respond.Respond(h.productUpdate))
			p.With(ruleAuthorizeProduct).Delete("/{product_id}", h.Web.Respond.Respond(h.productDelete))
		})
	})

	// System Routes
	router := chi.NewRouter()
	router.Get("/readiness", h.Web.Respond.Respond(h.readiness))
	router.Get("/liveness", h.Web.Respond.Respond(h.liveness))
	router.Handle("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("./doc.json"),
	))
	router.Get("/swagger/doc.json", h.Swagger)

	router.Mount("/api", apiRouter)
	return router
}

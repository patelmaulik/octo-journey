package service

import (
	"errors"
	"net/http"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	limitermd "github.com/didip/tollbooth"
	limiter "github.com/didip/tollbooth/limiter"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/urfave/negroni"
	"patelmaulik.com/maulik/v1/dbclient"
	"patelmaulik.com/maulik/v1/models"
)

type MyPermissionType interface {
	IsAdmin() bool
}

type AccountHandler struct {
	DBClient dbclient.IDatabaseRepository
}

func CreateRoute(router *chi.Mux, h *AccountHandler) {
	// Mount the admin sub-router
	router.Mount("/account", accountRouter(h))
	router.Mount("/admin", adminRouter(h))
}

// A completely separate router for administrator routes
func adminRouter(h *AccountHandler) http.Handler {
	r := chi.NewRouter()
	// r.Use(CheckSecurityPolicy)
	w := NewNegroni()

	r.Get("/{accountId}", w.UnSecureRoute(h.getAccountById))
	r.Get("/hr/{accountId}", w.SecureRoute(h.getAccountById))
	return r
}

// A completely separate router for administrator routes
func accountRouter(h *AccountHandler) http.Handler {
	r := chi.NewRouter()
	// r.Use(CheckSecurityPolicy)

	w := NewNegroni() // setupRateLimittingMiddleware(h.getAccountById)

	r.Get("/", h.getAllAccounts)
	r.Get("/{accountId}", w.UnSecureRoute(h.getAccountById))

	return r
}

type NegroniWrapper struct {
	N *negroni.Negroni
}

func NewNegroni(handlers ...negroni.Handler) NegroniWrapper {
	w := NegroniWrapper{}
	d := append(handlers, negroni.NewLogger())
	w.N = negroni.New(d...)
	return w
}

// Just wraps the handler
func (n *NegroniWrapper) UnSecureRoute(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return n.N.With(negroni.WrapFunc(handlerFunc)).ServeHTTP
}

// Secures the route
func (n *NegroniWrapper) SecureRoute(handlerFunc http.HandlerFunc) http.HandlerFunc {
	mw := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte("secret"), nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})

	return n.N.With(negroni.HandlerFunc(mw.HandlerWithNext), negroni.WrapFunc(handlerFunc)).ServeHTTP
	//nx := negroni.New(negroni.NewLogger(), negroni.HandlerFunc(mw.HandlerWithNext), negroni.WrapFunc(handlerFunc)) // negroni.HandlerFunc(mw.HandlerWithNext),
	//n.Use(negroni.HandlerFunc(mw.HandlerWithNext))
}

func setupRateLimittingMiddleware(handler http.HandlerFunc) negroni.Handler {
	lmt := limitermd.NewLimiter(1, &limiter.ExpirableOptions{
		DefaultExpirationTTL: 1,
		ExpireJobInterval:    time.Second,
	})

	lmt.
		SetIPLookups([]string{"RemoteAddr", "X-Forwarded-For", "X-Real-IP"}).
		SetMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

	return negroni.Wrap(limitermd.LimitFuncHandler(lmt, handler))
}

// getAccountById gets account by id
func (h *AccountHandler) getAccountById(w http.ResponseWriter, r *http.Request) {

	if accountId := chi.URLParam(r, "accountId"); accountId != "" {

		var account models.Account
		var err error

		client := h.DBClient

		if account, err = client.QueryAccount(accountId); err != nil {
			render.Render(w, r, ErrNotFoundRequest(errors.New("accountId not found.")))
			return
		}

		if err := render.Render(w, r, &AccountResponse{&account}); err != nil {
			render.Render(w, r, ErrRender(err))
			return
		}

		return
	}

	render.Render(w, r, ErrInvalidRequest(errors.New("account not found.")))
	return
}

// AccountResponse type
type AccountResponse struct {
	*models.Account
}

func (rd *AccountResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	return nil
}

func (h *AccountHandler) getAllAccounts(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Accounts: Get all..."))
}

func CheckSecurityPolicy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		isAdmin, ok := ctx.Value("acl.permission").(bool)
		if !ok || !isAdmin {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     http.StatusText(http.StatusBadRequest),
		ErrorText:      err.Error(),
	}
}

func ErrNotFoundRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusNotFound,
		StatusText:     http.StatusText(http.StatusNotFound),
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

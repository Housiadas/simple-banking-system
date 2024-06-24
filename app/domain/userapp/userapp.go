// Package userapp maintains the app layer api for the user domain.
package userapp

import (
	"context"
	"errors"
	"net/mail"

	"github.com/Housiadas/backend-system/business/domain/authbus"
	"github.com/Housiadas/backend-system/business/domain/userbus"
	"github.com/Housiadas/backend-system/business/sys/errs"
	"github.com/Housiadas/backend-system/business/sys/order"
	"github.com/Housiadas/backend-system/business/sys/page"
	"github.com/Housiadas/backend-system/business/sys/validation"
	"github.com/Housiadas/backend-system/business/web"
)

// App manages the set of app layer api functions for the user domain.
type App struct {
	userBus *userbus.Business
	authbus *authbus.Auth
}

// NewApp constructs a user app API for use.
func NewApp(userBus *userbus.Business, authbus *authbus.Auth) *App {
	return &App{
		authbus: authbus,
		userBus: userBus,
	}
}

// Authenticate provides an API to authenticate the user.
func (c *App) Authenticate(ctx context.Context, authUser AuthenticateUser) (User, error) {
	addr, err := mail.ParseAddress(authUser.Email)
	if err != nil {
		return User{}, validation.NewFieldsError("email", err)
	}

	usr, err := c.userBus.Authenticate(ctx, *addr, authUser.Password)
	if err != nil {
		return User{}, err
	}

	return toAppUser(usr), nil
}

// Token provides an API token for the authenticated user.
func (c *App) Token(ctx context.Context, kid string) (Token, error) {
	if c.authbus == nil {
		return Token{}, errs.Newf(errs.Internal, "authapi not configured")
	}

	claims := web.GetClaims(ctx)

	tkn, err := c.authbus.GenerateToken(kid, claims)
	if err != nil {
		return Token{}, errs.New(errs.Internal, err)
	}

	return toToken(tkn), nil
}

// Create adds a new user to the system.
func (a *App) Create(ctx context.Context, app NewUser) (User, error) {
	nc, err := toBusNewUser(app)
	if err != nil {
		return User{}, errs.New(errs.InvalidArgument, err)
	}

	usr, err := a.userBus.Create(ctx, nc)
	if err != nil {
		if errors.Is(err, userbus.ErrUniqueEmail) {
			return User{}, errs.New(errs.Aborted, userbus.ErrUniqueEmail)
		}
		return User{}, errs.Newf(errs.Internal, "create: usr[%+v]: %s", usr, err)
	}

	return toAppUser(usr), nil
}

// Update updates an existing user.
func (a *App) Update(ctx context.Context, app UpdateUser) (User, error) {
	uu, err := toBusUpdateUser(app)
	if err != nil {
		return User{}, errs.New(errs.InvalidArgument, err)
	}

	usr, err := web.GetUser(ctx)
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "user missing in context: %s", err)
	}

	updUsr, err := a.userBus.Update(ctx, usr, uu)
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "update: userID[%s] uu[%+v]: %s", usr.ID, uu, err)
	}

	return toAppUser(updUsr), nil
}

// UpdateRole updates an existing user's role.
func (a *App) UpdateRole(ctx context.Context, app UpdateUserRole) (User, error) {
	uu, err := toBusUpdateUserRole(app)
	if err != nil {
		return User{}, errs.New(errs.InvalidArgument, err)
	}

	usr, err := web.GetUser(ctx)
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "user missing in context: %s", err)
	}

	updUsr, err := a.userBus.Update(ctx, usr, uu)
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "updaterole: userID[%s] uu[%+v]: %s", usr.ID, uu, err)
	}

	return toAppUser(updUsr), nil
}

// Delete removes a user from the system.
func (a *App) Delete(ctx context.Context) error {
	usr, err := web.GetUser(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "userID missing in context: %s", err)
	}

	if err := a.userBus.Delete(ctx, usr); err != nil {
		return errs.Newf(errs.Internal, "delete: userID[%s]: %s", usr.ID, err)
	}

	return nil
}

// Query returns a list of users with paging.
func (a *App) Query(ctx context.Context, qp QueryParams) (page.Result[User], error) {
	p, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return page.Result[User]{}, validation.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return page.Result[User]{}, err
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return page.Result[User]{}, validation.NewFieldsError("order", err)
	}

	usrs, err := a.userBus.Query(ctx, filter, orderBy, p)
	if err != nil {
		return page.Result[User]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.userBus.Count(ctx, filter)
	if err != nil {
		return page.Result[User]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return page.NewResult(toAppUsers(usrs), total, p), nil
}

// QueryByID returns a user by its Ia.
func (a *App) QueryByID(ctx context.Context) (User, error) {
	usr, err := web.GetUser(ctx)
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return toAppUser(usr), nil
}

package ginadapter

import (
	"mime"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yukihito-jokyu/topic2html/backend/apperr"
	domainauth "github.com/yukihito-jokyu/topic2html/backend/domain/auth"
	"github.com/yukihito-jokyu/topic2html/backend/observability"
	usecaseauth "github.com/yukihito-jokyu/topic2html/backend/usecase/auth"
)

const (
	oauthTransactionCookie = "__Host-topic2html_oauth_tx"
	adminSessionCookie     = "__Host-topic2html_admin_session"
	failureRedirect        = "/admin/login?reason=failed"
	oauthCookieMaxAge      = int(domainauth.OAuthTransactionLifetime / time.Second)
	sessionCookieMaxAge    = int(domainauth.SessionAbsoluteLifetime / time.Second)
)

type OAuthHandler struct {
	service usecaseauth.OAuthService
	logger  observability.EventLogger
}

func NewOAuthHandler(service usecaseauth.OAuthService, logger observability.EventLogger) *OAuthHandler {
	return &OAuthHandler{
		service: service,
		logger:  logger,
	}
}

func (h *OAuthHandler) Start(c *gin.Context) {
	if !isFormPost(c.Request) {
		h.logger.Error(c.Request.Context(), "oauth.start.rejected", apperr.New(apperr.CodeInvalidRequest))
		redirectFailure(c)

		return
	}
	returnPaths, ok := parseReturnPath(c.Request)
	if !ok {
		h.logger.Error(c.Request.Context(), "oauth.start.rejected", apperr.New(apperr.CodeInvalidRequest))
		redirectFailure(c)

		return
	}
	output, err := h.service.Start(c.Request.Context(), usecaseauth.StartInput{
		Origins:           c.Request.Header.Values("Origin"),
		ReturnPaths:       returnPaths,
		PreviousReference: transactionCookie(c.Request),
	})
	if err != nil || output.TransactionReference == "" || output.AuthorizationURL == "" {
		if err == nil {
			h.logger.Error(c.Request.Context(), "oauth.start.rejected", apperr.New(apperr.CodeInternal))
		}
		redirectFailure(c)

		return
	}
	setCookie(c, oauthTransactionCookie, output.TransactionReference, oauthCookieMaxAge)
	h.logger.Info(c.Request.Context(), "oauth.start.completed")
	c.Redirect(http.StatusSeeOther, output.AuthorizationURL)
}

func (h *OAuthHandler) Callback(c *gin.Context) {
	query := c.Request.URL.Query()
	state, stateOK := singleQueryValue(query, "state")
	code, codeOK := singleQueryValue(query, "code")
	providerError, errorOK := singleQueryValue(query, "error")
	if !stateOK || !codeOK || !errorOK {
		h.logger.Error(c.Request.Context(), "oauth.callback.rejected", apperr.New(apperr.CodeInvalidRequest))
		deleteCookie(c, oauthTransactionCookie)
		redirectFailure(c)

		return
	}
	output, err := h.service.Callback(c.Request.Context(), usecaseauth.CallbackInput{
		TransactionReference: transactionCookie(c.Request),
		Code:                 code,
		State:                state,
		ProviderError:        providerError,
	})
	deleteCookie(c, oauthTransactionCookie)
	if err != nil || output.SessionReference == "" || output.ReturnPath != "/admin" {
		if err == nil {
			h.logger.Error(c.Request.Context(), "oauth.callback.rejected", apperr.New(apperr.CodeInternal))
		}
		redirectFailure(c)

		return
	}
	setCookie(c, adminSessionCookie, output.SessionReference, sessionCookieMaxAge)
	h.logger.Info(c.Request.Context(), "oauth.callback.completed")
	c.Redirect(http.StatusSeeOther, output.ReturnPath)
}

func transactionCookie(request *http.Request) string {
	cookie, err := request.Cookie(oauthTransactionCookie)
	if err != nil {
		return ""
	}

	return cookie.Value
}

func isFormPost(request *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))

	return err == nil && mediaType == "application/x-www-form-urlencoded"
}

func parseReturnPath(request *http.Request) ([]string, bool) {
	if err := request.ParseForm(); err != nil {
		return nil, false
	}
	for key := range request.PostForm {
		if key != "return_path" {
			return nil, false
		}
	}
	values, present := request.PostForm["return_path"]
	if !present {
		return nil, true
	}
	if len(values) != 1 || values[0] == "" {
		return nil, false
	}

	return values, true
}

func singleQueryValue(values url.Values, key string) (string, bool) {
	items, present := values[key]
	if !present {
		return "", true
	}
	if len(items) != 1 {
		return "", false
	}

	return items[0], true
}

func redirectFailure(c *gin.Context) {
	c.Redirect(http.StatusSeeOther, failureRedirect)
}

func setCookie(c *gin.Context, name, value string, maxAge int) {
	c.Writer.Header().Add("Set-Cookie", (&http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}).String())
}

func deleteCookie(c *gin.Context, name string) {
	c.Writer.Header().Add("Set-Cookie", (&http.Cookie{
		Name:     name,
		MaxAge:   -1,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}).String())
}

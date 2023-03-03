package apis

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/search"
)

// bindLogsApi registers the request logs api endpoints.
func bindLogsApi(app core.App, rg *echo.Group) {
	api := logsApi{app: app}

	subGroup := rg.Group("/logs", RequireAdminAuth())
	subGroup.GET("/requests", api.requestsList)
	subGroup.GET("/requests/stats", api.requestsStats)
	subGroup.GET("/requests/:id", api.requestView)

	subGroup.GET("/errors", api.errorsList)

	subGroup.GET("/logs", api.logsList)
}

type logsApi struct {
	app core.App
}

var requestFilterFields = []string{
	"rowid", "id", "created", "updated",
	"url", "method", "status", "auth",
	"remoteIp", "userIp", "referer", "userAgent",
}

var errorFilterFields = []string{
	"rowid", "id", "created", "updated",
	"file", "line", "error", "fatal", "meta",
}

var logFilterFields = []string{
	"rowid", "id", "created", "updated",
	"level", "message", "meta",
}

func (api *logsApi) requestsList(c echo.Context) error {
	fieldResolver := search.NewSimpleFieldResolver(requestFilterFields...)

	result, err := search.NewProvider(fieldResolver).
		Query(api.app.LogsDao().RequestQuery()).
		ParseAndExec(c.QueryParams().Encode(), &[]*models.Request{})

	if err != nil {
		return NewBadRequestError("", err)
	}

	return c.JSON(http.StatusOK, result)
}

func (api *logsApi) requestsStats(c echo.Context) error {
	fieldResolver := search.NewSimpleFieldResolver(requestFilterFields...)

	filter := c.QueryParam(search.FilterQueryParam)

	var expr dbx.Expression
	if filter != "" {
		var err error
		expr, err = search.FilterData(filter).BuildExpr(fieldResolver)
		if err != nil {
			return NewBadRequestError("Invalid filter format.", err)
		}
	}

	stats, err := api.app.LogsDao().RequestsStats(expr)
	if err != nil {
		return NewBadRequestError("Failed to generate requests stats.", err)
	}

	return c.JSON(http.StatusOK, stats)
}

func (api *logsApi) requestView(c echo.Context) error {
	id := c.PathParam("id")
	if id == "" {
		return NewNotFoundError("", nil)
	}

	request, err := api.app.LogsDao().FindRequestById(id)
	if err != nil || request == nil {
		return NewNotFoundError("", err)
	}

	return c.JSON(http.StatusOK, request)
}

func (api *logsApi) errorsList(c echo.Context) error {
	fieldResolver := search.NewSimpleFieldResolver(errorFilterFields...)

	result, err := search.NewProvider(fieldResolver).
		Query(api.app.LogsDao().ErrorQuery()).
		ParseAndExec(c.QueryParams().Encode(), &[]*models.Error{})

	if err != nil {
		return NewBadRequestError("", err)
	}

	return c.JSON(http.StatusOK, result)
}

func (api *logsApi) logsList(c echo.Context) error {
	fieldResolver := search.NewSimpleFieldResolver(logFilterFields...)

	result, err := search.NewProvider(fieldResolver).
		Query(api.app.LogsDao().LogQuery()).
		ParseAndExec(c.QueryParams().Encode(), &[]*models.Log{})

	if err != nil {
		return NewBadRequestError("", err)
	}

	return c.JSON(http.StatusOK, result)
}

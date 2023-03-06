package logs

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {}

const (
	ContextAdminKey      string = "admin"
	ContextAuthRecordKey string = "authRecord"
	ContextCollectionKey string = "collection"
)

// Returns the "real" user IP from common proxy headers (or fallbackIp if none is found).
//
// The returned IP value shouldn't be trusted if not behind a trusted reverse proxy!
func realUserIp(r *http.Request, fallbackIp string) string {
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}

	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	if ipsList := r.Header.Get("X-Forwarded-For"); ipsList != "" {
		ips := strings.Split(ipsList, ",")
		// extract the rightmost ip
		for i := len(ips) - 1; i >= 0; i-- {
			ip := strings.TrimSpace(ips[i])
			if ip != "" {
				return ip
			}
		}
	}

	return fallbackIp
}

var (
	Writer io.Writer
	daodb  *daos.Dao

	LogLevel  = models.InfoLevel
	LogFormat = FormatBasic
	colored   = true
)

type outputFormat string

const (
	FormatBasic outputFormat = "basic"
	FormatJSON  outputFormat = "json"
	FormatText  outputFormat = "text"
)

type Event struct {
	level models.LogLevel
	model models.Model
	meta  types.JsonMap
}

func (e *Event) flush() {
	if e.model == nil {
		return
	}

	if Writer == nil && daodb == nil {
		return
	}

	switch m := e.model.(type) {
	case *models.Error, *models.Request, *models.Log:
		m.RefreshUpdated()
	default:
		fmt.Fprintf(Writer, "error: unknown active model\n")

		return
	}

	if e.level < LogLevel {
		return
	}

	switch m := e.model.(type) {
	case *models.Error:
		m.Meta = e.meta
	case *models.Request:
		m.Meta = e.meta
	case *models.Log:
		m.Meta = e.meta
	}

	if Writer != nil {
		switch LogFormat {
		case FormatBasic:
			var msg string

			switch m := e.model.(type) {
			case *models.Error:
				msg = m.String()
			case *models.Request:
				msg = m.String()
			case *models.Log:
				msg = m.String()
			}

			if colored {
				fmt.Fprintln(Writer, color.HiBlackString(msg))
			} else {
				fmt.Fprintln(Writer, msg)
			}
		case FormatText:
			b, err := json.Marshal(e.model)
			if err != nil {
				panic(err)
			}

			var j map[string]any
			err = json.Unmarshal(b, &j)
			if err != nil {
				panic(err)
			}

			fmt.Fprintf(Writer, "%s ", time.Now().Format(time.RFC3339Nano))

			for k, v := range j {
				if k == "meta" || k == "created" || k == "updated" || k == "id" {
					continue
				}

				sv := fmt.Sprint(v)
				if sv == "" || sv == "null" || sv == "[]" || sv == "{}" || sv == "0" {
					continue
				}

				fmt.Fprintf(Writer, `%s="%v" `, k, sv)
			}

			switch m := j["meta"].(type) {
			case map[string]any:
				for k, v := range m {
					fmt.Fprintf(Writer, `_%s="%v" `, k, v)
				}
			}

			fmt.Fprintln(Writer)

		case FormatJSON:
			err := json.NewEncoder(Writer).Encode(e.model)
			if err != nil {
				fmt.Fprintf(Writer, "error: %s\n", err.Error())

				return
			}
		default:
			fmt.Fprintf(Writer, "error: unknown log format: %s\n", LogFormat)
			return
		}
	}

	if daodb != nil {
		err := daodb.Save(e.model)
		if err != nil {
			fmt.Fprintf(Writer, "error: %s\n", err.Error())

			return
		}

		e.model = nil
	}
}

func (e *Event) setMeta(key string, value any) *Event {
	if e.meta == nil {
		e.meta = make(types.JsonMap)
	}

	e.meta[key] = fmt.Sprint(value)

	return e
}

func (e *Event) Str(key, value string) *Event {
	return e.setMeta(key, value)
}

func (e *Event) Strs(key string, value []string) *Event {
	b, err := json.Marshal(value)
	if err != nil {
		return e.setMeta(key, value)
	} else {
		return e.setMeta(key, string(b))
	}
}

func (e *Event) Interface(key string, value any) *Event {
	return e.setMeta(key, value)
}

func (e *Event) Int(key string, value int) *Event {
	return e.setMeta(key, value)
}

func (e *Event) Err(err error) *Event {
	if err == nil {
		return e
	}

	switch m := e.model.(type) {
	case *models.Error:
		_, file, line, _ := runtime.Caller(1)
		m.Error = err.Error()
		m.File = file
		m.Line = line
	default:
		e.setMeta("error", err.Error())
	}

	return e
}

func (e *Event) Time(key string, t time.Time) *Event {
	return e.setMeta(key, t.Format(types.DefaultDateLayout))
}

func (e *Event) Duration(key string, d time.Duration) *Event {
	return e.setMeta(key, d.String())
}

func (e *Event) Msg(s string) {
	e.Msgf("%s", s)
}

func (e *Event) Msgf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)

	if s == "" {
		e.flush()

		return
	}

	s = strings.TrimSuffix(s, "\n")

	switch m := e.model.(type) {
	case *models.Log:
		m.Message = s
	default:
		e.setMeta("msg", s)
	}

	e.flush()
}

func SetDao(db *daos.Dao) {
	daodb = db
}

func NoColour() {
	colored = false
}

func (e *Event) Request(c echo.Context) *Event {
	httpRequest := c.Request()
	httpResponse := c.Response()
	status := httpResponse.Status

	requestAuth := models.RequestAuthGuest
	if c.Get(ContextAuthRecordKey) != nil {
		requestAuth = models.RequestAuthRecord
	} else if c.Get(ContextAdminKey) != nil {
		requestAuth = models.RequestAuthAdmin
	}

	ip, _, _ := net.SplitHostPort(httpRequest.RemoteAddr)

	e.model = &models.Request{
		Url:       httpRequest.URL.RequestURI(),
		Method:    strings.ToLower(httpRequest.Method),
		Status:    status,
		Auth:      requestAuth,
		UserIp:    realUserIp(httpRequest, ip),
		RemoteIp:  ip,
		Referer:   httpRequest.Referer(),
		UserAgent: httpRequest.UserAgent(),
	}

	return e
}

func Trace() *Event {
	return &Event{level: models.TraceLevel, model: &models.Log{
		Level: models.TraceLevel,
	}}
}
func Debug() *Event {
	return &Event{level: models.DebugLevel, model: &models.Log{
		Level: models.DebugLevel,
	}}
}
func Info() *Event {
	return &Event{level: models.InfoLevel, model: &models.Log{
		Level: models.InfoLevel,
	}}
}
func Warn() *Event {
	return &Event{level: models.WarnLevel, model: &models.Log{
		Level: models.WarnLevel,
	}}
}

func Error() *Event {
	return &Event{level: models.ErrorLevel, model: &models.Error{}}
}

func Fatal() *Event {
	return &Event{level: models.FatalLevel, model: models.NewFatalError(false)}
}

func Panic() *Event {
	return &Event{level: models.PanicLevel, model: models.NewFatalError(true)}
}

func Println(v ...interface{}) {
	Writer.Write([]byte(fmt.Sprintln(v...)))
}

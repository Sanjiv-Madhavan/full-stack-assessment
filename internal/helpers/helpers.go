package helpers

import (
	"encoding/json"
	"full-stack-assesment/internal/scheme"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string, details ...string) {
	WriteJSON(w, status, scheme.Error{Code: status, Message: msg})
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, scheme.Error{
		Code:    status,
		Message: msg,
	})
}

func NowRFC3339() string { return time.Now().UTC().Format(time.RFC3339Nano) }

func ParseTimeOrNow(s string) time.Time {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t
	}
	return time.Now().UTC()
}

func ParseUUIDParam(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	raw := strings.TrimSpace(r.PathValue(name))
	u, err := uuid.Parse(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, name+" must be a valid uuid")
		return uuid.Nil, false
	}
	return u, true
}

func MustUUID(s string) types.UUID {
	u, err := uuid.Parse(s)
	if err != nil {
		return types.UUID(uuid.Nil)
	}
	return types.UUID(u)
}

func ClampInt(v, min, max, def int) int {
	if v == 0 {
		return def
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func ParseLimitOffset(q url.Values) (limit, offset int) {
	limit = 50
	offset = 0
	if s := q.Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			limit = ClampInt(n, 1, 200, 50)
		}
	}
	if s := q.Get("offset"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n >= 0 {
			offset = n
		}
	}
	return
}

func NormalizeStatus(s string) (string, bool) {
	s = strings.TrimSpace(strings.ToUpper(s))
	switch s {
	case "TODO", "IN_PROGRESS", "DONE":
		return s, true
	default:
		return "", false
	}
}

// Copyright 2021-2025 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package restful

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type nullIntReq struct {
	MaxAge *int32  `json:"maxAge,omitempty"`
	Note   *string `json:"note,omitempty"`
	Count  *int32  `json:"count,omitempty" nullable:"true"`
}

func TestDisallowNullIntegersGlobal(t *testing.T) {
	assert := assert.New(t)
	DisallowNullIntegers = true
	t.Cleanup(func() { DisallowNullIntegers = false })

	r := NewRouter()
	r.HandleFunc("/check", func(req nullIntReq) error { return nil }).Methods(http.MethodPost)

	rr := postJSON(t, r, `{"maxAge":null}`)
	assert.Equal(http.StatusBadRequest, rr.Code)

	rr = postJSON(t, r, `{"maxAge":0}`)
	assert.Equal(http.StatusNoContent, rr.Code)

	rr = postJSON(t, r, `{}`)
	assert.Equal(http.StatusNoContent, rr.Code)

	rr = postJSON(t, r, `{"note":null}`)
	assert.Equal(http.StatusNoContent, rr.Code)

	rr = postJSON(t, r, `{"count":null}`)
	assert.Equal(http.StatusNoContent, rr.Code)
}

func TestDisallowNullIntegersRouter(t *testing.T) {
	assert := assert.New(t)

	r := NewRouter().DisallowNullIntegers()
	r.HandleFunc("/check", func(req nullIntReq) error { return nil }).Methods(http.MethodPost)

	rr := postJSON(t, r, `{"maxAge":null}`)
	assert.Equal(http.StatusBadRequest, rr.Code)

	plain := NewRouter()
	plain.HandleFunc("/check", func(req nullIntReq) error { return nil }).Methods(http.MethodPost)
	rr = postJSON(t, plain, `{"maxAge":null}`)
	assert.Equal(http.StatusNoContent, rr.Code)
}

func TestRejectNullIntegersNested(t *testing.T) {
	type inner struct {
		N *int32 `json:"n,omitempty"`
	}
	type outer struct {
		Config inner `json:"config"`
	}
	err := rejectNullIntegers([]byte(`{"config":{"n":null}}`), &outer{})
	assert.Error(t, err)
	assert.NoError(t, rejectNullIntegers([]byte(`{"config":{}}`), &outer{}))
}

func postJSON(t *testing.T, r *Router, body string) *httptest.ResponseRecorder {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, "/check", bytes.NewReader([]byte(body)))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

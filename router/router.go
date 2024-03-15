package router

import (
	"io"
	"net/http"

	lua "github.com/yuin/gopher-lua"
)

type Router struct {
	router  *http.ServeMux
	exports map[string]lua.LGFunction
}

func New(router *http.ServeMux) *Router {
	return &Router{
		router:  router,
		exports: make(map[string]lua.LGFunction),
	}
}

func (r *Router) Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"handle": r.handleFunc,
	})
	L.SetField(mod, "name", lua.LString("vvv1"))
	L.Push(mod)
	return 1
}

func (r *Router) handleFunc(L *lua.LState) int {
	path := L.ToString(1)
	fn := L.ToFunction(2)

	r.router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		state := lua.NewState()

		state.CallByParam(lua.P{
			Fn: fn,
		}, newRequest(state, r), newWriter(state, w, r))
	})

	return 1
}

func newRequest(L *lua.LState, req *http.Request) *lua.LTable {
	tab := L.NewTable()
	bodyReader := L.NewFunction(func(L *lua.LState) int {
		data, err := io.ReadAll(req.Body)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(string(data)))
		return 1
	})

	tab.RawSetString("body", bodyReader)
	tab.RawSetString("host", lua.LString(req.Host))
	tab.RawSetString("method", lua.LString(req.Method))
	tab.RawSetString("referer", lua.LString(req.Referer()))
	tab.RawSetString("proto", lua.LString(req.Proto))
	tab.RawSetString("user_agent", lua.LString(req.UserAgent()))
	tab.RawSetString("url", lua.LString(req.URL.String()))
	if req.URL != nil && len(req.URL.Query()) > 0 {
		query := L.NewTable()
		for k, v := range req.URL.Query() {
			if len(v) > 0 {
				query.RawSetString(k, lua.LString(v[0]))
			}
		}
		tab.RawSetString("query", query)
	}

	if len(req.Header) > 0 {
		headers := L.NewTable()
		for k, v := range req.Header {
			if len(v) > 0 {
				headers.RawSetString(k, lua.LString(v[0]))
			}
		}
		tab.RawSetString("headers", headers)
	}

	tab.RawSetString("path", lua.LString(req.URL.Path))
	tab.RawSetString("raw_path", lua.LString(req.URL.RawPath))
	tab.RawSetString("raw_query", lua.LString(req.URL.RawQuery))
	tab.RawSetString("request_uri", lua.LString(req.RequestURI))
	tab.RawSetString("remote_addr", lua.LString(req.RemoteAddr))

	return tab
}

func newWriter(L *lua.LState, w http.ResponseWriter, r *http.Request) *lua.LTable {
	tab := L.NewTable()
	// 写出 body
	tab.RawSetString("write", L.NewFunction(func(state *lua.LState) int {
		data := state.CheckAny(2).String()
		count, err := w.Write([]byte(data))
		state.Push(lua.LNumber(count))
		if err != nil {
			state.Push(lua.LString(err.Error()))
			return 2
		}
		return 1
	}))
	tab.RawSetString("json", L.NewFunction(func(state *lua.LState) int {
		w.Header().Set("Content-Type", "application/json")

		data := state.CheckAny(2).String()
		count, err := w.Write([]byte(data))
		state.Push(lua.LNumber(count))
		if err != nil {
			state.Push(lua.LString(err.Error()))
			return 2
		}
		return 1
	}))
	tab.RawSetString("header", L.NewFunction(func(state *lua.LState) int {
		key := state.CheckString(2)
		val := state.CheckString(3)
		w.Header().Set(key, val)
		return 1
	}))
	tab.RawSetString("write_header", L.NewFunction(func(state *lua.LState) int {
		code := state.CheckInt(2)
		w.WriteHeader(code)
		return 1
	}))
	tab.RawSetString("done", L.NewFunction(func(state *lua.LState) int {
		return 1
	}))

	tab.RawSetString("redirect", L.NewFunction(func(state *lua.LState) int {
		url := state.CheckString(2)
		http.Redirect(w, r, url, http.StatusFound)
		return 1
	}))

	return tab
}

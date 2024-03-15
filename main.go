package main

import (
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	lua "github.com/yuin/gopher-lua"

	"github.com/sendya/webmock/router"
)

var (
	listenAddr = ":"
)

func init() {
	flag.StringVar(&listenAddr, "addr", ":8080", "lister addr")
}
func main() {

	// 创建 HTTP 路由
	r := http.NewServeMux()

	loader(r)

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGUSR1)
		for {
			<-ch

			slog.Info("Reload http handles..")
			r = http.NewServeMux()
			loader(r)
			slog.Info("Reload http handles done.")
		}
	}()

	lis := mustListen(listenAddr)
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			slog.Info("New Request", "method", req.Method, "url", req.URL.String(), "client-ip", req.RemoteAddr)
		}()
		r.ServeHTTP(w, req)
	})

	slog.Info("Start HTTP server", "addr", listenAddr)

	err := http.Serve(lis, handler)
	if err != nil {
		panic(err)
	}
}

func mustListen(addr string) net.Listener {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	return listen
}

func loader(r *http.ServeMux) {
	router := router.New(r)

	// 注册 Lua 模块
	L := lua.NewState()
	L.PreloadModule("router", router.Loader)

	// 扫描 routes 下的所有 lua 文件
	files, err := os.ReadDir("routes")
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if err := L.DoFile(filepath.Join("routes", file.Name())); err != nil {
			slog.Warn("load lua file error", "file", file.Name(), "error", err.Error())
			continue
		}
		slog.Debug("Load lua file", "file", file.Name())
	}

	L.Close()
}

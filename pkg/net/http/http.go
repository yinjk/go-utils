//@Desc
//@Date 2019-11-12 19:49
//@Author yinjk
package http

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	readTimeOut  time.Duration = 10000 * time.Second
	writeTimeOut time.Duration = 10000 * time.Second
)

type Engine struct {
	*gin.Engine
	Port   string
	server *http.Server
}
type Config struct {
	Mode         string
	Port         string
	ReadTimeOut  time.Duration
	WriteTimeOut time.Duration
}

func Default() *Engine {
	return &Engine{Engine: newGinEngine(gin.DebugMode), Port: ":8080"}
}

func NewEngine(conf Config) *Engine {
	return &Engine{
		Engine: newGinEngine(conf.Mode),
		Port:   conf.Port,
		server: nil,
	}
}

//StartUp start with goroutine
func (e *Engine) StartUp() {
	s := &http.Server{
		Addr:           e.Port,
		Handler:        e,
		ReadTimeout:    readTimeOut,
		WriteTimeout:   writeTimeOut,
		MaxHeaderBytes: 1 << 20,
	}
	e.server = s
	go func() {
		// service connections
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("listen: %s\n", err))
		}
	}()
}

//ListenAndStartUp start with block
func (e *Engine) ListenAndStartUp() {
	e.StartUp()
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Print("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := e.server.Shutdown(ctx); err != nil {
		panic(fmt.Sprintf("Server Shutdown: %s", err))
	}
	log.Print("Server exiting")
}

//创建gin框架Engine
func newGinEngine(ginMode string) *gin.Engine {
	if ginMode == gin.DebugMode || ginMode == gin.ReleaseMode || ginMode == gin.TestMode {
		gin.SetMode(ginMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	return gin.Default()
}

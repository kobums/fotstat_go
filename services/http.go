package services

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"fotstat/global/config"
	"fotstat/global/log"
	"fotstat/router"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func Http() {
	log.Info().Str("service", "HTTP").Msg("Start Service")

	app := fiber.New(fiber.Config{
		BodyLimit:             500 * 1024 * 1024,
		Prefork:               false,
		CaseSensitive:         true,
		StrictRouting:         true,
		DisableStartupMessage: true,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
	})

	sites := strings.Join(config.Cors, ", ")
	if sites != "" {
		app.Use(cors.New(cors.Config{
			AllowOrigins:     sites,
			AllowCredentials: true,
			AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
			AllowHeaders:     "Origin,Content-Type,Accept,Content-Length,Accept-Language,Accept-Encoding,Connection,Access-Control-Allow-Origin,Authorization",
		}))
	}

	if config.Log.Web {
		app.Use(fiberzerolog.New(fiberzerolog.Config{
			Logger: log.Get(),
			// 기본값의 "url" 은 쿼리스트링을 포함해 GET /api/jwt?email=&password= 의
			// 비밀번호가 로그에 평문으로 남는다 — 쿼리스트링 없는 path 만 기록한다.
			Fields: []string{fiberzerolog.FieldIP, fiberzerolog.FieldLatency, fiberzerolog.FieldStatus, fiberzerolog.FieldMethod, fiberzerolog.FieldPath, fiberzerolog.FieldError},
		}))
	}

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestCompression,
	}))


	// 로그(webdata/log)·임시파일(webdata/temp)은 UploadPath 아래에 있으므로
	// 정적 공개 경로에서 제외한다.
	// 정적 핸들러와 같은 정규화 경로(fasthttp URI.Path)로 검사해 "//", "../" 우회를 막는다.
	app.Use(func(ctx *fiber.Ctx) error {
		p := string(ctx.Context().URI().Path())
		for _, blocked := range []string{"/webdata/log", "/webdata/temp"} {
			if p == blocked || strings.HasPrefix(p, blocked+"/") {
				return ctx.SendStatus(fiber.StatusNotFound)
			}
		}
		return ctx.Next()
	})

	// app.Static("/webdata", config.DocumentRoot)
	app.Static("/webdata", config.UploadPath)

	router.SetRouter(app)

	app.Get("/*", func(ctx *fiber.Ctx) error {
		return ctx.SendFile(fmt.Sprintf("./%v/index.html", config.DocumentRoot), true)
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-c
		log.Info().Msg("Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	if config.Mode == "develop" || !config.Tls.Use {
		if err := app.Listen(":" + config.Port); err != nil {
			log.Error().Msg(err.Error())
		}
	} else {
		cer, err := tls.LoadX509KeyPair(config.Tls.Cert, config.Tls.Key)
		if err != nil {
			log.Error().Msg("TLS error")
			log.Error().Msg(err.Error())
			return
		}

		cert := &tls.Config{Certificates: []tls.Certificate{cer}}

		ln, err := tls.Listen("tcp", ":"+config.Port, cert)
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		if err := app.Listener(ln); err != nil {
			log.Error().Msg(err.Error())
		}
	}
}

package commands

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/urfave/cli/v2"
)

// Server serve fyne wasm application over http
type Server struct {
	*appData
	debug       bool
	port        int
	srcDir, dir string
}

// Serve return the cli command for serving fyne wasm application over http
func Serve() *cli.Command {
	s := &Server{appData: &appData{}}

	return &cli.Command{
		Name:        "serve",
		Aliases:     []string{"s"},
		Usage:       "Packages an application using WebAssembly and exposes it via a web server",
		Description: `The serve command packages an application using WebAssembly and expose it via a web server which port can be overridden with port.`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "debug",
				Usage:       "enable compiling in debug mode",
				Destination: &s.debug,
			},
			stringFlags["src"](&s.srcDir),
			stringFlags["icon"](&s.icon),
			intFlags["http-port"](&s.port),
		},
		Action: s.Server,
	}
}

func (s *Server) requestPackage() error {
	p := &Packager{
		os:      "wasm",
		srcDir:  s.srcDir,
		appData: s.appData,
		release: !s.debug,
	}

	if s.debug {
		p.tags = "debug"
	}

	err := p.Package()
	s.dir = p.dir
	return err
}

func (s *Server) serve() error {
	err := s.validate()
	if err != nil {
		return err
	}

	err = s.requestPackage()
	if err != nil {
		return err
	}

	webDir := util.EnsureSubDir(s.dir, "wasm")
	fileServer := http.FileServer(http.Dir(webDir))

	http.Handle("/", fileServer)

	fmt.Printf("Serving %s at: http://localhost:%v\n", s.appData.AppID, s.port)
	err = http.ListenAndServe(":"+strconv.Itoa(s.port), nil)

	return err
}

// Server serve the wasm application over http
func (s *Server) Server(ctx *cli.Context) error {
	if ctx.Args().Len() != 0 {
		return errors.New("unexpected parameter after flags")
	}

	return s.serve()
}

func (s *Server) validate() error {
	if s.port == 0 {
		s.port = 8080
	}
	if s.port < 0 || s.port > 65535 {
		return fmt.Errorf("the port must be a strictly positive number and be strictly smaller than 65536 (Got %v)", s.port)
	}
	return nil
}

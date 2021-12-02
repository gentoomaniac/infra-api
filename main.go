package main

import (
	"encoding/json"
	"net/http"

	"github.com/alecthomas/kong"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/gentoomaniac/infra-api/pkg/gocli"
	"github.com/gentoomaniac/infra-api/pkg/logging"
)

var (
	version = "unknown"
	commit  = "unknown"
	binName = "unknown"
	builtBy = "unknown"
	date    = "unknown"
)

var (
	ListenAddr = ":10000"
	Names      = make(map[string]bool)
)

var cli struct {
	logging.LoggingConfig

	Version gocli.VersionFlag `short:"V" help:"Display version."`
}

func main() {
	ctx := kong.Parse(&cli, kong.UsageOnError(), kong.Vars{
		"version": version,
		"commit":  commit,
		"binName": binName,
		"builtBy": builtBy,
		"date":    date,
	})
	logging.Setup(&cli.LoggingConfig)

	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", root).Methods("GET")
	myRouter.HandleFunc("/hello/{name}", hello).Methods("GET")
	myRouter.HandleFunc("/names/{name}", updateName).Methods("PUT")
	myRouter.HandleFunc("/names/{name}", addName).Methods("POST")

	log.Info().Str("listenAddr", ListenAddr).Msg("starting server")
	if err := http.ListenAndServe(ListenAddr, myRouter); err != nil {
		log.Error().Err(err).Msg("")
	}

	ctx.Exit(0)
}

type Message struct {
	Msg string `json:"msg"`
}

func root(w http.ResponseWriter, r *http.Request) {
	data := Message{
		Msg: "hello, world!",
	}

	json.NewEncoder(w).Encode(data)
}

func hello(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	data := Message{
		Msg: "hello, " + name + "!",
	}

	json.NewEncoder(w).Encode(data)
}

func addName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	if _, ok := Names[name]; ok {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(Message{Msg: "already exists"})
		log.Debug().Str("name", name).Msg("name already exists")
		return
	}

	Names[name] = true
	log.Debug().Str("name", name).Msg("added new name")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Message{Msg: "ok"})
}

func updateName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	Names[name] = true
	log.Debug().Str("name", name).Msg("updated name")
	json.NewEncoder(w).Encode(Message{Msg: "ok"})
}

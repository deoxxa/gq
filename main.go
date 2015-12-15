package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app       = kingpin.New("gq", "Perform a GraphQL query")
	url       = app.Flag("url", "GraphQL Endpoint URL").Default("http://127.0.0.1:3000/graphql").String()
	queryFile = app.Flag("query_file", "File containing a GraphQL query").String()
	query     = app.Flag("query", "GraphQL Query (overrides query_file)").String()
	varsFile  = app.Flag("vars_file", "JSON file containing variables for the query").String()
	vars      = app.Flag("vars", "JSON-encoded variables for the query (overrides vars_file)").String()
	headers   = app.Flag("headers", "Headers to send with the request").StringMap()
	verbose   = app.Flag("verbose", "Show response headers and other noisy information on stderr").Bool()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	var q string
	var v map[string]interface{}

	switch {
	case query != nil && *query != "":
		q = *query
	case queryFile != nil && *queryFile != "":
		qd, err := ioutil.ReadFile(*queryFile)
		if err != nil {
			panic(err)
		}
		q = string(qd)
	default:
		panic(fmt.Errorf("a query must be provided"))
	}

	switch {
	case vars != nil && *vars != "":
		if err := json.Unmarshal([]byte(*vars), &v); err != nil {
			panic(err)
		}
	case varsFile != nil && *varsFile != "":
		vd, err := ioutil.ReadFile(*varsFile)
		if err != nil {
			panic(err)
		}
		if err := json.Unmarshal(vd, &v); err != nil {
			panic(err)
		}
	}

	d := struct {
		Q string                 `json:"query"`
		V map[string]interface{} `json:"variables"`
	}{q, v}

	j, err := json.Marshal(d)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", *url, bytes.NewReader(j))
	if err != nil {
		panic(err)
	}
	for k, v := range *headers {
		req.Header.Set(k, v)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "%s %s\n", res.Proto, res.Status)
		for k, l := range res.Header {
			for _, v := range l {
				fmt.Fprintf(os.Stderr, "%s: %s\n", k, v)
			}
		}
		fmt.Fprint(os.Stderr, "\n")
	}

	if _, err := io.Copy(os.Stdout, res.Body); err != nil {
		panic(err)
	}
}

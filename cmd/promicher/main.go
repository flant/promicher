package main

import (
	"github.com/flant/promicher/pkg/kube"
	"github.com/flant/promicher/pkg/promicher"
	"github.com/flant/promicher/pkg/server"
	"github.com/romana/rlog"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var (
	App = kingpin.New(filepath.Base(os.Args[0]), "The Promicher: Prometheus alerts enricher")

	Labels = App.
		Flag("labels", `Pattern to select labels from kubernetes resources by keys.
May be passed several times, all labels of kubernetes resource will be checked agains
each of the labels patterns. Pattern is a regular expression.
If regular expression groups are specified in the pattern,
then the last matched group will be used as result label key.
For example label {"monitoring.somehost.io/tier": "some-value"}
match agains pattern "monitoring.somehost.io/(tier)" will result
in prometheus label named {"tier": "some-value"}.

Examples:

{"hello/world": "value"} pattern "hello/(.*)" => {"world": "value"}

{"some.host.io/path": "value"} pattern "(.*)/(path)" => {"path": "value"} (using the last matched group)

{"some.host.io/path": "value"} pattern "some.host.io" -> NO MATCH

{"some.host.io/path": "value"} pattern "some.host.io/.*" -> {"some.host.io/path": "value"}`).
		Strings()

	Annotations = App.
			Flag("annotations", `Pattern to select annotations from kubernetes resources by keys.
May be passed several times, all annotations of kubernetes resource will be checked agains
each of the annotations patterns. The format is the same as labels.`).
			Strings()

	EvaluationInterval = App.
				Flag("evaluation-interval", "Prometheus evaluation interval.").
				Default("30s").
				String()

	Listen = App.
		Flag("listen", "Listen on the specified address for incoming requests.").
		Default("0.0.0.0:80").
		String()

	DestinationUrl = App.
			Flag("destination-url", "Proxy enriched requests to the specified address.").
			Default("http://localhost:8000/api/v1/alerts").
			String()
)

func WaitForExitCode() int {
	interruptCh := make(chan os.Signal, 1)
	signal.Notify(interruptCh, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case sig := <-interruptCh:
			rlog.Infof("Grace shutdown with %s signal", sig.String())
			return 0
		}
	}
}

func main() {
	App.HelpFlag.Short('h')

	App.Version("0.1.0")

	kingpin.MustParse(App.Parse(os.Args[1:]))

	kube, err := kube.NewKube()
	if err != nil {
		rlog.Criticalf("Cannot initialize kube: %s", err)
		os.Exit(1)
	}

	promicher := promicher.NewPromicher(kube, *Labels, *Annotations)
	srv := server.NewServer(*Listen, *DestinationUrl, promicher)

	go func() {
		err := srv.Run()
		if err != nil {
			rlog.Critical("Cannot start http server: %s", err)
			os.Exit(1)
		}
	}()

	os.Exit(WaitForExitCode())
}

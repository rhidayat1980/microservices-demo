// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/namsral/flag"
	"github.com/sirupsen/logrus"

	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/ochttp"
	_ "go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"

	"contrib.go.opencensus.io/exporter/jaeger"
	"contrib.go.opencensus.io/exporter/prometheus"
	"contrib.go.opencensus.io/exporter/zipkin"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
)

var (
	version        = "no version set"
	displayVersion = flag.Bool("version", false, "Show version and quit")
	logLevel       = flag.String("logLevel", "warn", "log level from debug, info, warning, error. When debug, genetate 100% Tracing")
	srvURL         = flag.String("srvURL", ":8080", "IP and port to bind, localhost:8080 or :8080")
	adFile         = flag.String("adFile", "ads.json", "path to the Ads json file")

	jaegerSvcAddr = flag.String("JAEGER_SERVICE_ADDR", "", "URL to Jaeger Tracing agent")
	zipkinSvcAddr = flag.String("ZIPKIN_SERVICE_ADDR", "", "URL to Zipkin Tracing agent (ex: zipkin:9411)")
)

func main() {
	// parse flags
	flag.Parse()
	if *displayVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	// setup logs
	log := logrus.New()
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}
	log.Out = os.Stdout
	currLogLevel, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("error parsing Log Level %s", err)
	}
	log.Level = currLogLevel

	a := &adserviceServer{
		adFile:   *adFile,
		adsIndex: make(map[string][]int),
	}

	err = a.loadAdsFile()
	if err != nil {
		log.Fatalf("error parsing Ads json file %s", err)
	}

	randomAd := a.getRandomAds()
	log.Infof("got Ad %v", randomAd)
	log.Infof("index %v", a.adsIndex)
	catAds := a.getAdsByCategory("photography")
	log.Infof("photography %v", catAds)
	// r := mux.NewRouter()
	// r.HandleFunc("/", svc.homeHandler).Methods(http.MethodGet, http.MethodHead)
	// r.HandleFunc("/product/{id}", svc.productHandler).Methods(http.MethodGet, http.MethodHead)
	// r.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "ok") })

	// also init the prometheus handler
	// go initTracing(log)
	// initPrometheusStats(log, r)

	// var handler http.Handler = r
	// handler = &logHandler{log: log, next: handler} // add logging
	// handler = ensureSessionID(handler)             // add session ID
	// handler = &ochttp.Handler{                     // add opencensus instrumentation
	// 	Handler:     handler,
	// 	Propagation: &b3.HTTPFormat{}
	// }

	log.Infof("starting server on " + *srvURL)
	// log.Fatal(http.ListenAndServe(*srvURL, handler))
}

func initTracing(log *logrus.Logger) {
	// This is a demo app with low QPS. trace.AlwaysSample() is used here
	// to make sure traces are available for observation and analysis.
	// In a production environment or high QPS setup please use
	// trace.ProbabilitySampler set at the desired probability.
	if log.Level == logrus.DebugLevel {
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	}

	initJaegerTracing(log)
	initZipkinTracing(log)
}

func initJaegerTracing(log *logrus.Logger) {

	if *jaegerSvcAddr == "" {
		log.Info("jaeger initialization disabled.")
		return
	}

	// Register the Jaeger exporter to be able to retrieve
	// the collected spans.
	exporter, err := jaeger.NewExporter(jaeger.Options{
		Endpoint: fmt.Sprintf("http://%s", *jaegerSvcAddr),
		Process: jaeger.Process{
			ServiceName: "frontend",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	trace.RegisterExporter(exporter)
	log.Info("jaeger initialization completed.")
}

func initZipkinTracing(log *logrus.Logger) {
	// start zipkin exporter
	// URL to zipkin is like http://zipkin.tcc:9411/api/v2/spans
	if *zipkinSvcAddr == "" {
		log.Info("zipkin initialization disabled.")
		return
	}

	reporter := zipkinhttp.NewReporter(fmt.Sprintf("http://%s/api/v2/spans", *zipkinSvcAddr))
	exporter := zipkin.NewExporter(reporter, nil)
	trace.RegisterExporter(exporter)

	log.Info("zipkin initialization completed.")
}

func initPrometheusStats(log logrus.FieldLogger, r *mux.Router) {
	// init the prometheus /metrics endpoint
	exporter, err := prometheus.NewExporter(prometheus.Options{})
	if err != nil {
		log.Fatal(err)
	}

	// register basic views
	initStats(log, exporter)

	// init the prometheus /metrics endpoint
	r.Handle("/metrics", exporter).Methods(http.MethodGet, http.MethodHead)
	log.Info("prometheus metrics initialization completed.")
}

func initStats(log logrus.FieldLogger, exporter *prometheus.Exporter) {
	view.SetReportingPeriod(60 * time.Second)
	view.RegisterExporter(exporter)
	if err := view.Register(ochttp.DefaultServerViews...); err != nil {
		log.Warn("Error registering http default server views")
	} else {
		log.Info("Registered http default server views")
	}
	if err := view.Register(ocgrpc.DefaultClientViews...); err != nil {
		log.Warn("Error registering grpc default client views")
	} else {
		log.Info("Registered grpc default client views")
	}
}

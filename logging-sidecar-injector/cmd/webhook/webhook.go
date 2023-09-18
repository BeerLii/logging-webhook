package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/bwei/logging-sidecar-injector/pkg/util"
	"github.com/sirupsen/logrus"
	kwhhttp "github.com/slok/kubewebhook/v2/pkg/http"
	kwhlogrus "github.com/slok/kubewebhook/v2/pkg/log/logrus"
	kwhmodel "github.com/slok/kubewebhook/v2/pkg/model"
	kwhmutating "github.com/slok/kubewebhook/v2/pkg/webhook/mutating"
	appsv1 "k8s.io/api/apps/v1"
	"net/http"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type config struct {
	certFile  string
	keyFile   string
	imagename string
}

func initFlags() *config {
	cfg := &config{}

	fl := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fl.StringVar(&cfg.certFile, "tls-cert-file", "", "TLS certificate file")
	fl.StringVar(&cfg.keyFile, "tls-key-file", "", "TLS key file")
	fl.StringVar(&cfg.imagename, "image-name", "", "sidecar image name")
	_ = fl.Parse(os.Args[1:])
	return cfg
}

func run() error {
	logrusLogEntry := logrus.NewEntry(logrus.New())
	logrusLogEntry.Logger.SetLevel(logrus.DebugLevel)
	logger := kwhlogrus.NewLogrus(logrusLogEntry)
	cfg := initFlags()
	// Create mutator.
	mt := kwhmutating.MutatorFunc(func(_ context.Context, _ *kwhmodel.AdmissionReview, obj metav1.Object) (*kwhmutating.MutatorResult, error) {
		depoy, ok := obj.(*appsv1.Deployment)
		d := depoy.DeepCopy()
		fmt.Println("ffff")
		fmt.Println(d)
		if !ok {
			return &kwhmutating.MutatorResult{}, nil
		}
		name, ok := d.Annotations["loginfo-inject-name"]
		fmt.Println("myname")
		fmt.Println(name)
		if !ok {
			return &kwhmutating.MutatorResult{}, nil
		}

		client := util.RestClient()
		fmt.Println("client")
		fmt.Println(client)
		if client == nil {
			return &kwhmutating.MutatorResult{}, nil
		}

		result := util.Fetch(client, depoy.ObjectMeta.Namespace, name)
		fmt.Println("result")
		fmt.Println(result)
		LogInfoList := result.Spec.Log
		llist := []util.Log{}
		for i, log := range LogInfoList {
			path := log.LogPath
			container := log.ConName
			flist := []util.LogDetails{}
			for _, detail := range log.LogDetail {
				flist = append(flist, util.LogDetails{detail.FileName, detail.Name})

			}
			v := util.CreateVolume(fmt.Sprintf("varlog%d", i))
			llist = append(llist, util.Log{flist, v, path, container, fmt.Sprintf("varlog%d", i)})
		}
		for _, l := range llist {
			util.PatchDeploy(l, d, cfg.imagename)
		}

		return &kwhmutating.MutatorResult{MutatedObject: d}, nil
	})

	// Create webhook.
	mcfg := kwhmutating.WebhookConfig{
		ID:      "log-annotate",
		Mutator: mt,
		Logger:  logger,
	}
	wh, err := kwhmutating.NewWebhook(mcfg)
	if err != nil {
		return fmt.Errorf("error creating webhook: %w", err)
	}

	// Get HTTP handler from webhook.
	whHandler, err := kwhhttp.HandlerFor(kwhhttp.HandlerConfig{Webhook: wh, Logger: logger})
	if err != nil {
		return fmt.Errorf("error creating webhook handler: %w", err)
	}

	// Serve.
	logger.Infof("Listening on :8080")
	err = http.ListenAndServeTLS(":8080", cfg.certFile, cfg.keyFile, whHandler)
	return nil
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running app: %s", err)
		os.Exit(1)
	}
}

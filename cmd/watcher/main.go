package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog"
	"go-tsv-watcher/config"
	"go-tsv-watcher/internal/handler"
	"go-tsv-watcher/internal/storage"
	"go-tsv-watcher/internal/storage/queries"
	"go-tsv-watcher/internal/usecase"
	"go-tsv-watcher/pkg/logger"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	loggerInstance := httplog.NewLogger("watcher", httplog.Options{
		Concise: true,
	})

	st, err := storage.New(cfg.DBConfig, logger.New(loggerInstance))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logic := usecase.New(st, cfg.DirectoryOut, logger.New(loggerInstance))

	go func() {
		err := logic.Process(ctx, cfg.Refresh, cfg.Directory)
		if err != nil {
			log.Fatal(err)
		}
	}()

	h := handler.New(logic)

	router := chi.NewRouter()
	router.Group(h.PublicRoutes)

	// http server
	go func() {
		if cfg.HTTPS != "" {
			log.Println("HTTPS server started on ", cfg.HTTPS)
			ca := &x509.Certificate{
				SerialNumber: big.NewInt(2019),
				Subject: pkix.Name{
					Organization: []string{"GasaySecure, INC."},
					Country:      []string{"US"},
					Locality:     []string{"New York"},
				},
				NotBefore:             time.Now(),
				NotAfter:              time.Now().AddDate(10, 0, 0),
				IsCA:                  true,
				ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
				KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
				BasicConstraintsValid: true,
			}
			caPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
			if err != nil {
				log.Fatalf("Failed to generate private key: %s", err.Error())
			}

			caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivateKey.PublicKey, caPrivateKey)
			if err != nil {
				log.Fatalf("Failed to create certificate: %s", err.Error())
			}

			caPEM := new(bytes.Buffer)
			pem.Encode(caPEM, &pem.Block{
				Type:  "CERTIFICATE",
				Bytes: caBytes,
			})

			caPrivKeyPEM := new(bytes.Buffer)
			pem.Encode(caPrivKeyPEM, &pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(caPrivateKey),
			})

			caFile, err := os.Create("ca.crt")
			if err != nil {
				log.Fatalf("Failed to create file: %s", err.Error())
			}
			_, err = caFile.Write(caPEM.Bytes())
			if err != nil {
				log.Fatalf("Failed to write file: %s", err.Error())
			}

			caFile.Close()

			caFile, err = os.Create("ca.key")
			if err != nil {
				log.Fatalf("Failed to create file: %s", err.Error())
			}

			_, err = caFile.Write(caPrivKeyPEM.Bytes())
			if err != nil {
				log.Fatalf("Failed to write file: %s", err.Error())
			}

			caFile.Close()

			http.ListenAndServeTLS(cfg.HTTPS, "ca.crt", "ca.key", router)
		} else {
			log.Println("HTTP server started on ", cfg.HTTP)
			if err := http.ListenAndServe(cfg.HTTP, router); err != nil {
				log.Fatal(err)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	cancel()

	err = queries.Close()
	if err != nil {
		log.Println(err)
	}

	log.Println("Done!")
}

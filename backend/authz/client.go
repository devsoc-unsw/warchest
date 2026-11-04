package authz

import (
	"context"
	"crypto/tls"
	"log"
	"os"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func NewClient(addr, presharedKey string) (*authzed.Client, error) {
	// Secure connection rpc connection with TLS
	creds := credentials.NewTLS(&tls.Config{})
	return authzed.NewClient(
		addr,
		grpcutil.WithBearerToken(presharedKey),
		grpc.WithTransportCredentials(creds),
	)
}

// NewInsecureClient is for local development and CI only
func NewInsecureClient(addr, presharedKey string) (*authzed.Client, error) {
	return authzed.NewClient(
		addr,
		grpcutil.WithInsecureBearerToken(presharedKey),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

func BuildClient(addr, presharedKey string) (*authzed.Client, error) {
	if os.Getenv("SPICEDB_INSECURE_MODE") == "true" {
		return NewInsecureClient(addr, presharedKey)
	}
	return NewClient(addr, presharedKey)
}

func WriteSchemaFromFile(
	ctx context.Context,
	client *authzed.Client,
	path string) error {
	schemaBytes, err := os.ReadFile(path)
	if err != nil {
		log.Println("path not found")
		return err
	}

	_, err = client.WriteSchema(
		ctx,
		&v1.WriteSchemaRequest{Schema: string(schemaBytes)},
	)
	if err != nil {
		log.Println("schema upload failed: ", err)
		return err
	}

	log.Println("schema uploaded successfully")
	return nil
}

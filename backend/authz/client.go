package authz

import (
	"context"
	"log"
	"os"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewClient(addr, presharedKey string) (*authzed.Client, error) {
	return authzed.NewClient(
		addr,
		grpcutil.WithInsecureBearerToken(presharedKey),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

func WriteSchemaFromFile(client *authzed.Client, path string) error {
	schemaBytes, err := os.ReadFile(path)
	if err != nil {
		log.Println("path not found")
		return err
	}

	_, err = client.WriteSchema(context.Background(), &v1.WriteSchemaRequest{Schema: string(schemaBytes)})
	if err != nil {
		log.Println("schema upload failed: ", err)
		return err
	}

	log.Println("schema uploaded successfully")
	return nil
}

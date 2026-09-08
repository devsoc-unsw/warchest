package authz

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
)

func createRelation(
	ctx context.Context,
	client *authzed.Client,
	resourceType, resourceID, relation, subjectType, subjectID string,
) {
	_, err := WriteRelationship(
		ctx, client, resourceType, resourceID, relation,
		subjectType, subjectID, "",
	)
	if err != nil {
		log.Printf("failed to write relationship: %s", err)
	}
}

func checkBobCanCreateEvents(ctx context.Context, client *authzed.Client) {
	result, err := CheckPermission(
		ctx, client, "organisation", "example_org", "event_create",
		"user", "Bob",
	)
	if err != nil {
		log.Printf("Permission check failed: %s", err)
		return
	}

	if result.Permissionship ==
		v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION {
		log.Println("Bob can create events")
	} else {
		log.Println("Bob cannot create events")
	}
}

func Test(t *testing.T) {
	spicedbAddr := os.Getenv("SPICEDB_ADDRESS")
	spicedbKey := os.Getenv("SPICEDB_PRESHARED_KEY")

	authzClient, err := BuildClient(spicedbAddr, spicedbKey)
	if err != nil {
		log.Fatalf("failed to connect to spicedb: %s", err)
	}

	// Context to prevent blocking calls
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// Releases resources when the test finishes to avoid waiting on timer
	defer cancel()

	err = WriteSchemaFromFile(ctx, authzClient, "spicedb_schema.zed")
	if err != nil {
		log.Printf("failed to write schema: %s", err)
	}

	createRelation(
		ctx, authzClient, "organisation", "example_org",
		"treasurer", "user", "John_Doe",
	)
	createRelation(
		ctx, authzClient, "organisation", "example_org",
		"subcom", "user", "Bob",
	)

	createRelation(
		ctx, authzClient, "event", "example_event",
		"member", "user", "Bob1",
	)
	createRelation(
		ctx, authzClient, "event", "example_event2",
		"member", "user", "Bob1",
	)
	createRelation(
		ctx, authzClient, "event", "example_event",
		"member", "user", "Bob2",
	)
	createRelation(
		ctx, authzClient, "event", "example_event",
		"member", "user", "Bob3",
	)

	relations, err := ReadRelationships(
		ctx, authzClient, "event", "example_event", "",
	)
	if err != nil {
		log.Printf("failed to read relationships: %s", err)
	}
	log.Printf("%+v %v", relations, err)

	resources, err := LookupResources(
		ctx, authzClient, "event", "user", "Bob1", "view",
	)
	if err != nil {
		log.Printf("failed to lookupResources: %s", err)
	}
	log.Printf("resources: %v", resources)

	subjects, err := LookupSubjects(
		ctx, authzClient, "user", "event", "example_event", "view",
	)
	if err != nil {
		log.Printf("failed to lookupSubjects: %s", err)
	}
	log.Printf("subjects: %v", subjects)

	_, err = DeleteRelationship(
		ctx, authzClient, "event", "example_event",
		"member", "user", "Bob3",
	)
	if err != nil {
		log.Printf("failed to delete relationship: %s", err)
	}

	subjects, err = LookupSubjects(
		ctx, authzClient, "user", "event", "example_event", "view",
	)
	if err != nil {
		log.Printf("failed to lookupSubjects: %s", err)
	}
	log.Printf("subjects: %v", subjects)

	checkBobCanCreateEvents(ctx, authzClient)

	log.Printf("End of test")
}

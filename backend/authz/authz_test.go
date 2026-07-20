package authz

import (
	"log"
	"os"
	"testing"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
)

func Test(t *testing.T) {
	spicedbAddr := os.Getenv("SPICEDB_ADDRESS")
	spicedbKey := os.Getenv("SPICEDB_PRESHARED_KEY")

	authzClient, err := NewClient(spicedbAddr, spicedbKey)
	if err != nil {
		log.Fatalf("failed to connect to spicedb: %s", err)
	}

	//Authzed route testing
	err = WriteSchemaFromFile(authzClient, "spicedb_schema.zed")
	if err != nil {
		log.Printf("failed to write schema: %s", err)
	}

	_, err = WriteRelationship(authzClient, "organisation", "example_org", "treasurer", "user", "John_Doe")
	if err != nil {
		log.Printf("failed to write relationship: %s", err)
	}

	_, err = WriteRelationship(authzClient, "organisation", "example_org", "subcom", "user", "Bob")
	if err != nil {
		log.Printf("failed to write relationship: %s", err)
	}

	//Create an event with certain members
	_, err = WriteRelationship(authzClient, "event", "example_event", "member", "user", "Bob1")
	if err != nil {
		log.Printf("failed to write relationship: %s", err)
	}

	_, err = WriteRelationship(authzClient, "event", "example_event2", "member", "user", "Bob1")
	if err != nil {
		log.Printf("failed to write relationship: %s", err)
	}

	_, err = WriteRelationship(authzClient, "event", "example_event", "member", "user", "Bob2")
	if err != nil {
		log.Printf("failed to write relationship: %s", err)
	}

	_, err = WriteRelationship(authzClient, "event", "example_event", "member", "user", "Bob3")
	if err != nil {
		log.Printf("failed to write relationship: %s", err)
	}

	relations, err := ReadRelationships(authzClient, "event", "example_event", "")
	if err != nil {
		log.Printf("failed to read relationships: %s", err)
	}

	log.Printf("%+v %v", relations, err)

	resources, err := LookupResources(authzClient, "event", "user", "Bob1", "view")
	if err != nil {
		log.Printf("failed to lookupResources: %s", err)
	}

	log.Printf("resources: %v", resources)

	subjects, err := LookupSubjects(authzClient, "user", "event", "example_event", "view")
	if err != nil {
		log.Printf("failed to lookupSubjects: %s", err)
	}

	log.Printf("subjects: %v", subjects)

	_, err = DeleteRelationship(authzClient, "event", "example_event", "member", "user", "Bob3")
	if err != nil {
		log.Printf("failed to delete relationship: %s", err)
	}

	subjects, err = LookupSubjects(authzClient, "user", "event", "example_event", "view")
	if err != nil {
		log.Printf("failed to lookupSubjects: %s", err)
	}

	log.Printf("subjects: %v", subjects)

	result, err := CheckPermission(authzClient, "organisation", "example_org", "event_create", "user", "Bob")
	if err != nil {
		log.Printf("Permission check failed: %s", err)
	}

	if result.Permissionship == v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION {
		log.Println("Bob can create events")
	} else {
		log.Println("Bob cannot create events")
	}

	_, err = CheckPermission(authzClient, "organisation", "example_org", "event_create", "user", "John_Doe")
	if err != nil {
		log.Printf("Permission check failed: %s", err)
	}

	log.Printf("Success")
}

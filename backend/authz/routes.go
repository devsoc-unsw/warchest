package authz

import (
	"context"
	"fmt"
	"io"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
)

/*
resourceType - The type of 'thing' being accessed, eg. document
resourceID - The ID of a specific 'thing'
relation - How the subject is connected to the resource, eg. viewer / editor / owner
subjectType - The kind of 'thing' that has access, eg. user / director / subcom
subjectID - The ID for that specific 'thing'
*/

// Pass in an array of relationships to make sure it happens in one go.
// Allow it to take an array of relationships.
// Debezium will make this redundant -> Whenever psql changes it will
// put a request to kafak and rewrite it into psql.

func WriteRelationship(
	ctx context.Context,
	client *authzed.Client,
	resourceType, resourceID, relation, subjectType, subjectID string,
	subjectRelation string,
) (*v1.WriteRelationshipsResponse, error) {
	if resourceID == "" || relation == "" || subjectID == "" {
		return nil, fmt.Errorf("resourceID, relation, and subjectID cannot be empty")
	}
	ref := &v1.SubjectReference{
		Object: &v1.ObjectReference{
			ObjectType: subjectType,
			ObjectId:   subjectID,
		},
	}

	// Allows multiple individuals in a relation to be written to
	if subjectRelation != "" {
		ref.OptionalRelation = subjectRelation
	}

	req := &v1.WriteRelationshipsRequest{
		Updates: []*v1.RelationshipUpdate{
			{
				Operation: v1.RelationshipUpdate_OPERATION_TOUCH,
				Relationship: &v1.Relationship{
					Resource: &v1.ObjectReference{
						ObjectType: resourceType,
						ObjectId:   resourceID,
					},
					Relation: relation,
					Subject:  ref,
				},
			},
		},
	}

	result, err := client.WriteRelationships(ctx, req)
	return result, err
}

func CheckPermission(
	ctx context.Context,
	client *authzed.Client,
	resourceType ObjectType, resourceID string,
	permission Permission,
	subjectType ObjectType, subjectID string,
) (bool, error) {
	req := &v1.CheckPermissionRequest{
		Resource: &v1.ObjectReference{
			ObjectType: string(resourceType),
			ObjectId:   resourceID,
		},
		Permission: string(permission),
		Subject: &v1.SubjectReference{
			Object: &v1.ObjectReference{
				ObjectType: string(subjectType),
				ObjectId:   subjectID,
			},
		},
		Consistency: &v1.Consistency{
			Requirement: &v1.Consistency_FullyConsistent{
				FullyConsistent: true,
			},
		},
	}

	result, err := client.CheckPermission(ctx, req)
	if err != nil {
		return false, err
	}

	return result.Permissionship ==
		v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION, nil
}

func DeleteRelationship(
	ctx context.Context,
	client *authzed.Client,
	resourceType ObjectType, resourceID string,
	relation Relation,
	subjectType ObjectType, subjectID string,
) (*v1.DeleteRelationshipsResponse, error) {
	if resourceID == "" || relation == "" || subjectID == "" {
		return nil, fmt.Errorf("resource / subject ID and relation cannot be empty")
	}
	req := &v1.DeleteRelationshipsRequest{
		RelationshipFilter: &v1.RelationshipFilter{
			ResourceType:       string(resourceType),
			OptionalResourceId: resourceID,
			OptionalRelation:   string(relation),
			OptionalSubjectFilter: &v1.SubjectFilter{
				SubjectType:       string(subjectType),
				OptionalSubjectId: subjectID,
			},
		},
	}

	result, err := client.DeleteRelationships(ctx, req)
	return result, err
}

// Returns a list of relationships which match a certain filter
// Example: Who are the treasurers? What relationships exist for
// a certain resource?
func ReadRelationships(
	ctx context.Context,
	client *authzed.Client,
	resourceType ObjectType, resourceID string, relation Relation,
) ([]*v1.Relationship, error) {
	req := &v1.ReadRelationshipsRequest{
		RelationshipFilter: &v1.RelationshipFilter{
			ResourceType:       string(resourceType),
			OptionalResourceId: resourceID,
			OptionalRelation:   string(relation),
		},
		Consistency: &v1.Consistency{
			Requirement: &v1.Consistency_FullyConsistent{
				FullyConsistent: true,
			},
		},
	}

	stream, err := client.ReadRelationships(ctx, req)
	if err != nil {
		return nil, err
	}
	var relationships []*v1.Relationship
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		relationships = append(relationships, resp.Relationship)
	}
	return relationships, nil
}

// Returns a list of resources where resources of type X can subject
// Y access via permission Z.
// Example: What events can the treasurer bob view?
func LookupResources(
	ctx context.Context,
	client *authzed.Client,
	subjectType ObjectType, subjectID string,
	permission Permission,
	resourceType ObjectType,
) ([]string, error) {
	req := &v1.LookupResourcesRequest{
		ResourceObjectType: string(resourceType),
		Permission:         string(permission),
		Subject: &v1.SubjectReference{
			Object: &v1.ObjectReference{
				ObjectType: string(subjectType),
				ObjectId:   subjectID,
			},
		},
		Consistency: &v1.Consistency{
			Requirement: &v1.Consistency_FullyConsistent{
				FullyConsistent: true,
			},
		},
	}

	stream, err := client.LookupResources(ctx, req)
	if err != nil {
		return nil, err
	}

	var resources []string
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		resources = append(resources, resp.ResourceObjectId)
	}

	return resources, nil
}

// Returns a list of subjects where subject X can access resource Y
// via permission Z.
// Example: Who can currently view an event?
func LookupSubjects(
	ctx context.Context,
	client *authzed.Client,
	resourceType ObjectType, resourceID string,
	permission Permission,
	subjectType ObjectType,
) ([]string, error) {
	req := &v1.LookupSubjectsRequest{
		Resource: &v1.ObjectReference{
			ObjectType: string(resourceType),
			ObjectId:   resourceID,
		},
		Permission:        string(permission),
		SubjectObjectType: string(subjectType),
		Consistency: &v1.Consistency{
			Requirement: &v1.Consistency_FullyConsistent{
				FullyConsistent: true,
			},
		},
	}

	stream, err := client.LookupSubjects(ctx, req)
	if err != nil {
		return nil, err
	}

	var subjects []string
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		subjects = append(subjects, resp.Subject.SubjectObjectId)
	}
	return subjects, nil
}

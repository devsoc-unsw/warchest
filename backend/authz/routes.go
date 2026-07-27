package authz

import (
	"context"
	"io"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
)

func WriteRelationship(
	client *authzed.Client,
	resourceType, resourceID, relation, subjectType, subjectID string,
) (*v1.WriteRelationshipsResponse, error) {
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
					Subject: &v1.SubjectReference{
						Object: &v1.ObjectReference{
							ObjectType: subjectType,
							ObjectId:   subjectID,
						},
					},
				},
			},
		},
	}

	result, err := client.WriteRelationships(context.Background(), req)
	return result, err
}

func CheckPermission(
	client *authzed.Client,
	resourceType, resourceID, permission, subjectType, subjectID string,
) (*v1.CheckPermissionResponse, error) {
	req := &v1.CheckPermissionRequest{
		Resource: &v1.ObjectReference{
			ObjectType: resourceType,
			ObjectId:   resourceID,
		},
		Permission: permission,
		Subject: &v1.SubjectReference{
			Object: &v1.ObjectReference{
				ObjectType: subjectType,
				ObjectId:   subjectID,
			},
		},
		Consistency: &v1.Consistency{
			Requirement: &v1.Consistency_FullyConsistent{
				FullyConsistent: true,
			},
		},
	}

	result, err := client.CheckPermission(context.Background(), req)
	return result, err
}

func DeleteRelationship(
	client *authzed.Client,
	resourceType, resourceID, relation, subjectType, subjectID string,
) (*v1.DeleteRelationshipsResponse, error) {
	req := &v1.DeleteRelationshipsRequest{
		RelationshipFilter: &v1.RelationshipFilter{
			ResourceType:       resourceType,
			OptionalResourceId: resourceID,
			OptionalRelation:   relation,
			OptionalSubjectFilter: &v1.SubjectFilter{
				SubjectType:       subjectType,
				OptionalSubjectId: subjectID,
			},
		},
	}

	result, err := client.DeleteRelationships(context.Background(), req)
	return result, err
}

// Returns a list of relationships which match a certain filter
// Example: Who are the treasurers? What relationships exist for
// a certain resource?
func ReadRelationships(
	client *authzed.Client,
	resourceType, resourceID, relation string,
) ([]*v1.Relationship, error) {
	req := &v1.ReadRelationshipsRequest{
		RelationshipFilter: &v1.RelationshipFilter{
			ResourceType:       resourceType,
			OptionalResourceId: resourceID,
			OptionalRelation:   relation,
		},
		Consistency: &v1.Consistency{
			Requirement: &v1.Consistency_FullyConsistent{
				FullyConsistent: true,
			},
		},
	}

	stream, err := client.ReadRelationships(context.Background(), req)
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
	client *authzed.Client,
	resourceType, subjectType, subjectID, permission string,
) ([]string, error) {
	req := &v1.LookupResourcesRequest{
		ResourceObjectType: resourceType,
		Permission:         permission,
		Subject: &v1.SubjectReference{
			Object: &v1.ObjectReference{
				ObjectType: subjectType,
				ObjectId:   subjectID,
			},
		},
		Consistency: &v1.Consistency{
			Requirement: &v1.Consistency_FullyConsistent{
				FullyConsistent: true,
			},
		},
	}

	stream, err := client.LookupResources(context.Background(), req)
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
	client *authzed.Client,
	subjectType, resourceType, resourceID, permission string,
) ([]string, error) {
	req := &v1.LookupSubjectsRequest{
		Resource: &v1.ObjectReference{
			ObjectType: resourceType,
			ObjectId:   resourceID,
		},
		Permission:        permission,
		SubjectObjectType: subjectType,
		Consistency: &v1.Consistency{
			Requirement: &v1.Consistency_FullyConsistent{
				FullyConsistent: true,
			},
		},
	}

	stream, err := client.LookupSubjects(context.Background(), req)
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
		subjects = append(subjects, resp.SubjectObjectId)
	}
	return subjects, nil
}

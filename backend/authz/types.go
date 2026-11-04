package authz

type ObjectType string
type Relation string
type Permission string

const (
	RelationViewer Relation = "viewer"
	RelationEditor Relation = "editor"
	RelationOwner  Relation = "owner"
)

const (
	PermissionView Permission = "view"
	PermissionEdit Permission = "edit"
)

func (r Relation) Valid() bool {
	switch r {
	case RelationViewer, RelationEditor, RelationOwner:
		return true
	}
	return false
}

func (p Permission) Valid() bool {
	switch p {
	case PermissionView, PermissionEdit:
		return true
	}
	return false
}

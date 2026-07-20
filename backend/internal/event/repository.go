// talks directly to the database, forwarded to db.Queries
// this layer separate data access from business logic

package event

import(
	"context"
	"backend/db"
)

// what methods do we have for event
type EventRepository interface {
	CreateEvent(ctx context.Context, params db.CreateEventParams) (db.Event, error)
	// add more methods for event...
}

// define eventRepositoryImpl struct
type eventRepositoryImpl struct {
	queries *db.Queries // go use pointer
}

// constructor function to create eventRepositoryImpl
func NewEventRepository(queries *db.Queries) EventRepository {
	return &eventRepositoryImpl{queries: queries} // stores the incoming queries into its queries field
	// & read pointer

}

// write func for eventRepositoryImpl to CreateEvent
func (r *eventRepositoryImpl) CreateEvent(ctx context.Context, params db.CreateEventParams) (db.Event, error){
	// (r *eventRepositoryImpl) is the receiver, this method belongs to the object r
	return r.queries.CreateEvent(ctx, params) // simply call sqlc-generated method and return the result
}
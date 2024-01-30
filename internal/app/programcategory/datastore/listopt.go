package datastore

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (impl ProgramCategoryStorerImpl) ListAsSelectOptionByFilter(ctx context.Context, f *ProgramCategoryPaginationListFilter) ([]*ProgramCategoryAsSelectOption, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	// Get a reference to the collection
	collection := impl.Collection

	startAfter := "" // The ID to start after, initially empty for the first page

	// Pagination query
	query := bson.M{}
	options := options.Find().
		SetLimit(int64(f.PageSize)).
		SetSort(bson.D{{f.SortField, f.SortOrder}})

	// Add filter conditions to the query
	if !f.TenantID.IsZero() {
		query["tenant_id"] = f.TenantID
	}

	if startAfter != "" {
		// Find the document with the given startAfter ID
		cursor, err := collection.FindOne(ctx, bson.M{"_id": startAfter}).DecodeBytes()
		if err != nil {
			log.Fatal(err)
		}
		options.SetSkip(1)
		query["_id"] = bson.M{"$gt": cursor.Lookup("_id").ObjectID()}
	}

	if f.Status != 0 {
		query["status"] = f.Status
	}

	// Full-name search
	if f.SearchName != "" {
		query["$name"] = bson.M{"$search": f.SearchName}
		options.SetProjection(bson.M{"score": bson.M{"$meta": "nameScore"}})
		options.SetSort(bson.D{{"score", bson.M{"$meta": "nameScore"}}})
	}

	options.SetSort(bson.D{{f.SortField, 1}}) // Sort in ascending order based on the specified field

	// Retrieve the list of items from the collection
	cursor, err := collection.Find(ctx, query, options)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	var results = []*ProgramCategoryAsSelectOption{}
	if err = cursor.All(ctx, &results); err != nil {
		panic(err)
	}

	return results, nil
}

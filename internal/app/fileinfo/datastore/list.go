package datastore

import (
	"context"
	"log"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (impl FileInfoStorerImpl) ListByFilter(ctx context.Context, f *FileInfoPaginationListFilter) (*FileInfoPaginationListResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	filter, err := impl.newPaginationFilter(f)
	if err != nil {
		return nil, err
	}

	// Add filter conditions to the filter
	if !f.TenantID.IsZero() {
		filter["tenant_id"] = f.TenantID
	}
	if f.Name != "" {
		filter["name"] = bson.M{"$regex": primitive.Regex{Pattern: f.Name, Options: "i"}}
	}
	if f.Description != "" {
		filter["description"] = bson.M{"$regex": primitive.Regex{Pattern: f.Description, Options: "i"}}
	}

	// Create a slice to store conditions
	var conditions []bson.M

	// Add filter conditions to the slice
	if !f.CreatedAtGTE.IsZero() {
		conditions = append(conditions, bson.M{"created_at": bson.M{"$gte": f.CreatedAtGTE}})
	}
	if !f.CreatedAtGT.IsZero() {
		conditions = append(conditions, bson.M{"created_at": bson.M{"$gt": f.CreatedAtGT}})
	}
	if !f.CreatedAtLTE.IsZero() {
		conditions = append(conditions, bson.M{"created_at": bson.M{"$lte": f.CreatedAtLTE}})
	}
	if !f.CreatedAtLT.IsZero() {
		conditions = append(conditions, bson.M{"created_at": bson.M{"$lt": f.CreatedAtLT}})
	}

	// Combine conditions with $and operator
	if len(conditions) > 0 {
		filter["$and"] = conditions
	}

	impl.Logger.Debug("fetching assistant file list",
		slog.String("Cursor", f.Cursor),
		slog.Int64("PageSize", f.PageSize),
		slog.String("SortField", f.SortField),
		slog.Any("SortOrder", f.SortOrder),
		slog.Any("TenantID", f.TenantID),
	)

	// Include additional filters for our cursor-based pagination pertaining to sorting and limit.
	options, err := impl.newPaginationOptions(f)
	if err != nil {
		return nil, err
	}

	// Include Full-text search
	if f.SearchText != "" {
		filter["$text"] = bson.M{"$search": f.SearchText}
		options.SetProjection(bson.M{"score": bson.M{"$meta": "textScore"}})
		options.SetSort(bson.D{{"score", bson.M{"$meta": "textScore"}}})
	}

	// Execute the query
	cursor, err := impl.Collection.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// var results = []*FileInfo{}
	// if err = cursor.All(ctx, &results); err != nil {
	// 	panic(err)
	// }

	// Retrieve the documents and check if there is a next page
	results := []*FileInfo{}
	hasNextPage := false
	for cursor.Next(ctx) {
		document := &FileInfo{}
		if err := cursor.Decode(document); err != nil {
			return nil, err
		}
		results = append(results, document)
		// Stop fetching documents if we have reached the desired page size
		if int64(len(results)) >= f.PageSize {
			hasNextPage = true
			break
		}
	}

	// Get the next cursor and encode it
	var nextCursor string
	if hasNextPage {
		nextCursor, err = impl.newPaginatorNextCursorForFull(f, results)
		if err != nil {
			return nil, err
		}
	}

	return &FileInfoPaginationListResult{
		Results:     results,
		NextCursor:  nextCursor,
		HasNextPage: hasNextPage,
	}, nil
}

func (impl FileInfoStorerImpl) ListAsSelectOptionByFilter(ctx context.Context, f *FileInfoPaginationListFilter) ([]*FileInfoAsSelectOption, error) {
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

	options.SetSort(bson.D{{f.SortField, 1}}) // Sort in ascending order based on the specified field

	// Retrieve the list of items from the collection
	cursor, err := collection.Find(ctx, query, options)
	if err != nil {
		return nil, nil
	}
	defer cursor.Close(ctx)

	var results = []*FileInfoAsSelectOption{}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, nil
	}

	return results, nil
}

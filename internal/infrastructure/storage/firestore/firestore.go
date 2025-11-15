package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/auth-service/internal/domain/model"
	sharedQuery "github.com/histopathai/auth-service/internal/shared/query"
	"google.golang.org/api/iterator"
)

type FirestoreUserRepositoryImpl struct {
	client     *firestore.Client
	collection string
}

func NewFirestoreUserRepository(client *firestore.Client, collection string) *FirestoreUserRepositoryImpl {
	return &FirestoreUserRepositoryImpl{
		client:     client,
		collection: collection,
	}
}

func (fur *FirestoreUserRepositoryImpl) Create(ctx context.Context, entity *model.User) error {
	data := UserToFirestoreMap(entity)
	_, err := fur.client.Collection(fur.collection).Doc(entity.UserID).Set(ctx, data)
	if err != nil {
		return MapFirestoreError(err)
	}

	return nil
}

func (fur *FirestoreUserRepositoryImpl) GetByUserID(ctx context.Context, userID string) (*model.User, error) {
	doc, err := fur.client.Collection(fur.collection).Doc(userID).Get(ctx)
	if err != nil {
		return nil, MapFirestoreError(err)
	}

	user, err := UserFromFirestoreDoc(doc)
	if err != nil {
		return nil, MapFirestoreError(err)
	}
	user.UserID = userID
	return user, nil
}

func (fur *FirestoreUserRepositoryImpl) Update(ctx context.Context, userID string, updates *model.UpdateUser) error {
	updateData := UpdateUserToFirestoreUpdates(updates)

	_, err := fur.client.Collection(fur.collection).Doc(userID).Update(ctx, updateData)
	if err != nil {
		return MapFirestoreError(err)
	}
	return nil
}

func (fur *FirestoreUserRepositoryImpl) Delete(ctx context.Context, userID string) error {
	_, err := fur.client.Collection(fur.collection).Doc(userID).Delete(ctx)
	if err != nil {
		return MapFirestoreError(err)
	}

	return nil
}

func (fur *FirestoreUserRepositoryImpl) List(ctx context.Context, pagination *sharedQuery.Pagination) (*sharedQuery.Result[*model.User], error) {

	query := fur.client.Collection(fur.collection).Query

	isLimited := pagination.Limit > 0
	if isLimited {
		query = query.Limit(pagination.Limit + 1).Offset(pagination.Offset)
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	results := make([]*model.User, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}

		entity, err := UserFromFirestoreDoc(doc)
		if err != nil {
			return nil, MapFirestoreError(err)
		}

		results = append(results, entity)
	}

	hasMore := false
	if isLimited && len(results) > pagination.Limit {
		hasMore = true
		results = results[:pagination.Limit]
	}

	return &sharedQuery.Result[*model.User]{
		Data:    results,
		Limit:   pagination.Limit,
		Offset:  pagination.Offset,
		HasMore: hasMore,
	}, nil

}

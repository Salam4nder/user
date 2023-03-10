package user

import (
	"context"
	"errors"
	"time"

	"github.com/Salam4nder/user/pkg/util"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Storage defines the user storage interface.
type Storage interface {
	InsertOne(ctx context.Context, param InsertOneParam) (string, error)
	FindOneByID(ctx context.Context, id string) (User, error)
	FindOneByEmail(ctx context.Context, email string) (User, error)
	FindByFilter(ctx context.Context, filter Filter) ([]User, error)
	UpdateOne(ctx context.Context, param UpdateParam) (User, error)
	DeleteOne(ctx context.Context, id string) error
}

// storage implements the Storage interface.
type storage struct {
	collection *mongo.Collection
}

// NewStorage creates a new user service.
func NewStorage(c *mongo.Collection) Storage {
	return &storage{
		collection: c,
	}
}

// InsertOne creates a new user. Returns the created ID as a string.
// An empty string and an error is returned if the user could not be created.
func (s *storage) InsertOne(
	ctx context.Context, param InsertOneParam) (string, error) {
	if err := param.Validate(); err != nil {
		return "", err
	}

	hasedPassword, err := util.HashPassword(param.Password)
	if err != nil {
		return "", err
	}

	param.Password = hasedPassword
	param.CreatedAt = time.Now().Format(time.DateOnly)

	createdUser, err := s.collection.InsertOne(ctx, param)
	if err != nil {
		return "", err
	}

	id := createdUser.InsertedID.(primitive.ObjectID).Hex()

	return id, nil
}

// FindOneByID returns a user by its ID.
// An empty user and an error is returned if the user could not be found.
func (s *storage) FindOneByID(ctx context.Context, id string) (User, error) {
	var user User

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return User{}, err
	}

	query := bson.D{{Key: "_id", Value: objID}}

	if err := s.collection.FindOne(ctx, query).Decode(&user); err != nil {
		return User{}, err
	}

	return user, nil
}

// FindOneByEmail returns a user by its email.
// An empty user and an error is returned if the user could not be found.
func (s *storage) FindOneByEmail(ctx context.Context, email string) (User, error) {
	var user User

	if email == "" {
		return User{}, errors.New("email is required")
	}

	query := bson.D{{Key: "email", Value: email}}

	if err := s.collection.FindOne(ctx, query).Decode(&user); err != nil {
		return User{}, err
	}

	return user, nil
}

// FindByFilter returns a list of users by a filter.
// An empty list and an error is returned if the users could not be found.
func (s *storage) FindByFilter(ctx context.Context, filter Filter) ([]User, error) {
	var users []User

	if err := filter.Validate(); err != nil {
		return []User{}, err
	}

	query := bson.D{
		{Key: "full_name", Value: filter.FullName},
		{Key: "email", Value: filter.Email},
		{Key: "created_at", Value: filter.CreatedAt},
	}

	cursor, err := s.collection.Find(ctx, query)
	if err != nil {
		return []User{}, err
	}

	if err = cursor.All(ctx, &users); err != nil {
		return []User{}, err
	}

	return users, nil
}

// UpdateOne updates a user by its ID.
// An empty user and an error is returned if the user could not be updated.
func (s *storage) UpdateOne(ctx context.Context, param UpdateParam) (User, error) {
	var user User

	if err := param.Validate(); err != nil {
		return User{}, err
	}

	objID, err := primitive.ObjectIDFromHex(param.ID)
	if err != nil {
		return user, err
	}

	query := bson.D{{Key: "_id", Value: objID}}

	update := bson.D{{Key: "$set", Value: param}}

	err = s.collection.FindOneAndUpdate(ctx, query, update).Decode(&user)
	if err != nil {
		return user, err
	}

	return user, nil
}

// DeleteOne deletes a user by its ID.
// An error is returned if the user could not be deleted.
func (s *storage) DeleteOne(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	query := bson.D{{Key: "_id", Value: objID}}

	res, err := s.collection.DeleteOne(ctx, query)
	if err != nil || res.DeletedCount < 1 {
		return err
	}

	return nil
}

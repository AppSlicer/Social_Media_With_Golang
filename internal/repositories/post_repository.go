package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/anonto42/nano-midea/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PostRepository defines the interface for post data operations
type PostRepository interface {
	CreatePost(ctx context.Context, post *models.Post) error
	GetPostByID(ctx context.Context, id string) (*models.Post, error)
	GetPostsByUserID(ctx context.Context, userID string, skip, limit int64) ([]models.Post, error)
	GetAllPosts(ctx context.Context, skip, limit int64) ([]models.Post, error)
	UpdatePost(ctx context.Context, id string, post *models.Post) error
	DeletePost(ctx context.Context, id string) error
	IncrementLikesCount(ctx context.Context, postID string) error
	DecrementLikesCount(ctx context.Context, postID string) error
	IncrementCommentsCount(ctx context.Context, postID string) error
	DecrementCommentsCount(ctx context.Context, postID string) error
}

// MongoPostRepository implements PostRepository for MongoDB
type MongoPostRepository struct {
	collection *mongo.Collection
}

// NewMongoPostRepository creates a new MongoPostRepository
func NewMongoPostRepository(db *mongo.Database) *MongoPostRepository {
	return &MongoPostRepository{collection: db.Collection("posts")}
}

// CreatePost creates a new post in MongoDB
func (r *MongoPostRepository) CreatePost(ctx context.Context, post *models.Post) error {
	post.ID = primitive.NewObjectID()
	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, post)
	return err
}

// GetPostByID retrieves a post by ID from MongoDB
func (r *MongoPostRepository) GetPostByID(ctx context.Context, id string) (*models.Post, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid post ID format: %w", err)
	}

	var post models.Post
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&post)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("post not found")
		}
		return nil, err
	}
	return &post, nil
}

// GetPostsByUserID retrieves posts by a specific user from MongoDB
func (r *MongoPostRepository) GetPostsByUserID(ctx context.Context, userID string, skip, limit int64) ([]models.Post, error) {
	var posts []models.Post
	findOptions := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &posts); err != nil {
		return nil, err
	}
	return posts, nil
}

// GetAllPosts retrieves all posts from MongoDB with pagination
func (r *MongoPostRepository) GetAllPosts(ctx context.Context, skip, limit int64) ([]models.Post, error) {
	var posts []models.Post
	findOptions := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &posts); err != nil {
		return nil, err
	}
	return posts, nil
}

// UpdatePost updates an existing post in MongoDB
func (r *MongoPostRepository) UpdatePost(ctx context.Context, id string, post *models.Post) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid post ID format: %w", err)
	}

	post.UpdatedAt = time.Now()
	update := bson.M{
		"$set": bson.M{
			"content":    post.Content,
			"image_urls": post.ImageURLs,
			"video_urls": post.VideoURLs,
			"updated_at": post.UpdatedAt,
		},
	}
	res, err := r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return fmt.Errorf("post not found or not modified")
	}
	return nil
}

// DeletePost deletes a post by ID from MongoDB
func (r *MongoPostRepository) DeletePost(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid post ID format: %w", err)
	}

	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("post not found")
	}
	return nil
}

// IncrementLikesCount increments the likes count of a post
func (r *MongoPostRepository) IncrementLikesCount(ctx context.Context, postID string) error {
	objID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return fmt.Errorf("invalid post ID format: %w", err)
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$inc": bson.M{"likes_count": 1}})
	return err
}

// DecrementLikesCount decrements the likes count of a post
func (r *MongoPostRepository) DecrementLikesCount(ctx context.Context, postID string) error {
	objID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return fmt.Errorf("invalid post ID format: %w", err)
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$inc": bson.M{"likes_count": -1}})
	return err
}

// IncrementCommentsCount increments the comments count of a post
func (r *MongoPostRepository) IncrementCommentsCount(ctx context.Context, postID string) error {
	objID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return fmt.Errorf("invalid post ID format: %w", err)
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$inc": bson.M{"comments_count": 1}})
	return err
}

// DecrementCommentsCount decrements the comments count of a post
func (r *MongoPostRepository) DecrementCommentsCount(ctx context.Context, postID string) error {
	objID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return fmt.Errorf("invalid post ID format: %w", err)
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$inc": bson.M{"comments_count": -1}})
	return err
}

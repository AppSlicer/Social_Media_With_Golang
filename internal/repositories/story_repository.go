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
	"gorm.io/gorm"
)

// StoryRepository defines the interface for story operations
type StoryRepository interface {
	CreateStory(ctx context.Context, story *models.Story) error
	GetStoryByID(ctx context.Context, id string) (*models.Story, error)
	GetStoriesByUserIDs(ctx context.Context, userIDs []string) ([]models.Story, error)
	GetActiveStories(ctx context.Context) ([]models.Story, error)
	DeleteExpiredStories(ctx context.Context) error
	MarkSeen(storySeen *models.StorySeen) error
	HasSeen(storyID string, userID uint) (bool, error)
	GetSeenStoryIDs(userID uint, storyIDs []string) (map[string]bool, error)
	AddReaction(reaction *models.StoryReaction) error
}

type storyRepository struct {
	mongoCollection *mongo.Collection
	pgDB            *gorm.DB
}

func NewStoryRepository(mongoDB *mongo.Database, pgDB *gorm.DB) StoryRepository {
	return &storyRepository{
		mongoCollection: mongoDB.Collection("stories"),
		pgDB:            pgDB,
	}
}

func (r *storyRepository) CreateStory(ctx context.Context, story *models.Story) error {
	story.ID = primitive.NewObjectID()
	story.CreatedAt = time.Now()
	story.ExpiresAt = time.Now().Add(24 * time.Hour)
	_, err := r.mongoCollection.InsertOne(ctx, story)
	return err
}

func (r *storyRepository) GetStoryByID(ctx context.Context, id string) (*models.Story, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid story ID format")
	}
	var story models.Story
	err = r.mongoCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&story)
	if err != nil {
		return nil, err
	}
	return &story, nil
}

func (r *storyRepository) GetStoriesByUserIDs(ctx context.Context, userIDs []string) ([]models.Story, error) {
	filter := bson.M{
		"user_id":    bson.M{"$in": userIDs},
		"expires_at": bson.M{"$gt": time.Now()},
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.mongoCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stories []models.Story
	if err = cursor.All(ctx, &stories); err != nil {
		return nil, err
	}
	return stories, nil
}

func (r *storyRepository) GetActiveStories(ctx context.Context) ([]models.Story, error) {
	filter := bson.M{"expires_at": bson.M{"$gt": time.Now()}}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.mongoCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stories []models.Story
	if err = cursor.All(ctx, &stories); err != nil {
		return nil, err
	}
	return stories, nil
}

func (r *storyRepository) DeleteExpiredStories(ctx context.Context) error {
	_, err := r.mongoCollection.DeleteMany(ctx, bson.M{"expires_at": bson.M{"$lte": time.Now()}})
	return err
}

func (r *storyRepository) MarkSeen(storySeen *models.StorySeen) error {
	storySeen.SeenAt = time.Now()
	return r.pgDB.Create(storySeen).Error
}

func (r *storyRepository) HasSeen(storyID string, userID uint) (bool, error) {
	var count int64
	err := r.pgDB.Model(&models.StorySeen{}).Where("story_id = ? AND user_id = ?", storyID, userID).Count(&count).Error
	return count > 0, err
}

func (r *storyRepository) GetSeenStoryIDs(userID uint, storyIDs []string) (map[string]bool, error) {
	result := make(map[string]bool)
	if len(storyIDs) == 0 {
		return result, nil
	}
	var seen []models.StorySeen
	err := r.pgDB.Where("user_id = ? AND story_id IN ?", userID, storyIDs).Find(&seen).Error
	if err != nil {
		return nil, err
	}
	for _, s := range seen {
		result[s.StoryID] = true
	}
	return result, nil
}

func (r *storyRepository) AddReaction(reaction *models.StoryReaction) error {
	reaction.CreatedAt = time.Now()
	return r.pgDB.Create(reaction).Error
}

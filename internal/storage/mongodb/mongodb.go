// Пакет для работы с базой данных MongoDB в сервисе комментариев.
package mongodb

import (
	"GoExamComments/internal/config"
	"GoExamComments/internal/storage"
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Название базы и коллекции в БД. Используются переменные,
// а не константы, так как в тестах им присваиваются другие
// значения.
var (
	dbName  string = "goExam"
	colName string = "comments"
)

// tmConn - таймаут на создание пула подключений.
const tmConn time.Duration = time.Second * 20

// Storage - пул подключений к БД.
type Storage struct {
	db *mongo.Client
}

// New - обертка для конструктора пула подключений new.
func New(cfg *config.Config) *Storage {
	opts := setOpts(cfg.StoragePath, cfg.StorageUser, cfg.StoragePasswd)
	storage, err := new(opts)
	if err != nil {
		log.Fatalf("failed to init storage: %s", err.Error())
	}
	return storage
}

// setOpts настраивает опции нового подключения к БД.
// Функция вынесена отдельно для подмены ее в пакете
// с тестами.
func setOpts(path, user, password string) *options.ClientOptions {
	credential := options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		AuthSource:    "admin",
		Username:      user,
		Password:      password,
	}
	opts := options.Client().ApplyURI(path).SetAuth(credential)
	return opts
}

// setTestOpts возвращает опции нового подключения без авторизации.
// func setTestOpts(path string) *options.ClientOptions {
// 	return options.Client().ApplyURI(path)
// }

// new - конструктор пула подключений к БД.
func new(opts *options.ClientOptions) (*Storage, error) {
	const operation = "storage.mongodb.new"

	tm, cancel := context.WithTimeout(context.Background(), tmConn)
	defer cancel()

	db, err := mongo.Connect(tm, opts)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	err = db.Ping(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	// Создаем индекс по полю postId, чтобы ускорить выдачу всех комментариев
	// по переданному ID поста.
	collection := db.Database(dbName).Collection(colName)
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "postId", Value: -1}},
	}
	_, err = collection.Indexes().CreateOne(tm, indexModel)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return &Storage{db: db}, nil
}

// Close - обертка для закрытия пула подключений.
func (s *Storage) Close() error {
	return s.db.Disconnect(context.Background())
}

// AddComment записывает переданный комментарий в БД.
func (s *Storage) AddComment(ctx context.Context, com storage.Comment) (string, error) {
	const operation = "storage.mongodb.AddComment"

	if _, err := primitive.ObjectIDFromHex(com.PostID); err != nil {
		return "", fmt.Errorf("%s: %w", operation, storage.ErrIncorrectPostID)
	}

	collection := s.db.Database(dbName).Collection(colName)

	// Проверим, что родительский комментарий существует, чтобы
	// избежать вставки комментария с некорректной связью.
	if com.ParentID != "" {
		id, err := primitive.ObjectIDFromHex(com.ParentID)
		if err != nil {
			return "", fmt.Errorf("%s: %w", operation, storage.ErrIncorrectParentID)
		}
		filter := bson.D{{Key: "_id", Value: id}}
		res := collection.FindOne(ctx, filter)
		if res.Err() != nil {
			if res.Err() == mongo.ErrNoDocuments {
				return "", fmt.Errorf("%s: %w", operation, storage.ErrParentNotFound)
			}
			return "", fmt.Errorf("%s: %w", operation, res.Err())
		}
	}

	id := primitive.NewObjectID()
	bsn := bson.D{
		{Key: "_id", Value: id},
		{Key: "parentId", Value: com.ParentID},
		{Key: "postId", Value: com.PostID},
		{Key: "pubTime", Value: primitive.NewDateTimeFromTime(time.Now())},
		{Key: "content", Value: com.Content},
	}

	_, err := collection.InsertOne(ctx, bsn)
	if err != nil {
		return "", fmt.Errorf("%s: %w", operation, err)
	}

	return id.Hex(), nil
}

// Comments возвращает все деревья комментариев по переданному ID поста,
// отсортированные по дате создания.
func (s *Storage) Comments(ctx context.Context, post string) ([]storage.Comment, error) {
	const operation = "storage.mongodb.AddComment"

	if post == "" {
		return nil, storage.ErrIncorrectPostID
	}
	if _, err := primitive.ObjectIDFromHex(post); err != nil {
		return nil, fmt.Errorf("%s: %w", operation, storage.ErrIncorrectPostID)
	}

	var comments []storage.Comment
	collection := s.db.Database(dbName).Collection(colName)
	opts := options.Find().SetSort(bson.D{{Key: "pubTime", Value: -1}})
	filter := bson.D{{Key: "postId", Value: post}}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	err = cursor.All(ctx, &comments)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	if len(comments) == 0 {
		return nil, fmt.Errorf("%s: %w", operation, storage.ErrNoComments)
	}
	return comments, nil
}

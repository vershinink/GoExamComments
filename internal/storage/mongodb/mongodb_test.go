// Пакет для работы с базой данных MongoDB в сервисе комментариев.

package mongodb

import (
	"GoExamComments/internal/storage"
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var path string = "mongodb://192.168.0.102:27017/"

// addOne добавляет один комментарий в БД и возвращает его ObjectID.
// Функция для использования в тестах.
func (s *Storage) addOne(com storage.Comment) (string, error) {
	bsn := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "parentId", Value: com.ParentID},
		{Key: "postId", Value: com.PostID},
		{Key: "pubTime", Value: primitive.NewDateTimeFromTime(time.Now())},
		{Key: "allowed", Value: true},
		{Key: "content", Value: com.Content},
		{Key: "childs", Value: bson.A{}},
	}
	collection := s.db.Database(dbName).Collection(colName)
	res, err := collection.InsertOne(context.Background(), bsn)
	if err != nil {
		return "", err
	}
	hex := res.InsertedID.(primitive.ObjectID)
	return hex.Hex(), nil
}

func Test_new(t *testing.T) {

	// Для тестирования авторизации.
	// opts := setOpts(path, "admin", os.Getenv("DB_PASSWD"))

	opts := setTestOpts(path)
	st, err := new(opts)
	if err != nil {
		t.Fatal(err.Error())
	}
	st.Close()
}

func TestStorage_AddComment_WithoutParents(t *testing.T) {
	dbName = "testDB"
	colName = "testComments"

	st, err := new(setTestOpts(path))
	if err != nil {
		t.Fatal(err.Error())
	}
	defer st.Close()

	tests := []struct {
		name    string
		s       *Storage
		comment storage.Comment
		wantErr bool
	}{
		{
			name:    "Comment_1_OK",
			s:       st,
			comment: storage.Comment{PostID: "news1", Content: "First comment on news 1"},
			wantErr: false,
		},
		{
			name:    "Comment_2_OK",
			s:       st,
			comment: storage.Comment{PostID: "news1", Content: "Second comment on news 1"},
			wantErr: false,
		},
		{
			name:    "Comment_3_OK",
			s:       st,
			comment: storage.Comment{PostID: "news2", Content: "Comment on news 2"},
			wantErr: false,
		},
		{
			name:    "Empty_Post_ID",
			s:       st,
			comment: storage.Comment{Content: "Empty Post ID"},
			wantErr: true,
		},
		{
			name:    "Empty_Content",
			s:       st,
			comment: storage.Comment{PostID: "news1"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.AddComment(context.Background(), tt.comment); (err != nil) != tt.wantErr {
				t.Errorf("Storage.AddComment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStorage_AddComment_Parents(t *testing.T) {
	dbName = "testDB"
	colName = "testComments"

	st, err := new(setTestOpts(path))
	if err != nil {
		t.Fatal(err.Error())
	}
	defer st.Close()

	parent, err := st.addOne(storage.Comment{PostID: "news1", Content: "Comment with childs"})
	if err != nil {
		t.Fatalf("addOne error = %v", err)
	}
	t.Log(parent)

	tests := []struct {
		name    string
		s       *Storage
		comment storage.Comment
		wantErr bool
	}{
		{
			name:    "Child_Comment_1_OK",
			s:       st,
			comment: storage.Comment{ParentID: parent, PostID: "news1", Content: "First child comment"},
			wantErr: false,
		},
		{
			name:    "Child_Comment_2_OK",
			s:       st,
			comment: storage.Comment{ParentID: parent, PostID: "news1", Content: "Second child comment"},
			wantErr: false,
		},
		{
			name:    "Incorrect_Parent_ID",
			s:       st,
			comment: storage.Comment{ParentID: "asdf", PostID: "news1", Content: "Incorrect Parent ID"},
			wantErr: true,
		},
		{
			name:    "Incorrect_Post_ID",
			s:       st,
			comment: storage.Comment{ParentID: parent, PostID: "asdf", Content: "Second child comment"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.AddComment(context.Background(), tt.comment); (err != nil) != tt.wantErr {
				t.Errorf("Storage.AddComment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStorage_Comments(t *testing.T) {
	dbName = "testDB"
	colName = "testComments"

	st, err := new(setTestOpts(path))
	if err != nil {
		t.Fatal(err.Error())
	}
	defer st.Close()

	var count int = 3
	var id string = strconv.Itoa(rand.Int())
	for i := 1; i <= count; i++ {
		_, err := st.addOne(storage.Comment{PostID: id, Content: fmt.Sprintf("Comment %d on news %s", i, id)})
		if err != nil {
			t.Fatalf("addOne error = %v", err)
		}
	}

	tests := []struct {
		name    string
		s       *Storage
		post    string
		want    int
		wantErr bool
	}{
		{
			name:    "Comments_OK",
			s:       st,
			post:    id,
			want:    count,
			wantErr: false,
		},
		{
			name:    "Empty_Post_ID",
			s:       st,
			post:    "",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Incorrect_Post_ID",
			s:       st,
			post:    "asdf",
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Comments(context.Background(), tt.post)
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.Comments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("Storage.Comments() error = len %v, want %v", got, tt.want)
				return
			}
		})
	}
}

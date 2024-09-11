// Пакет tree содержит инструменты для создания и сортировки дерева комментариев.

package tree

import (
	"GoExamComments/internal/storage"
	"testing"
)

var comments = []storage.Comment{
	{ID: "com-10", ParentID: "com-3"},
	{ID: "com-9", ParentID: "com-7"},
	{ID: "com-8", ParentID: ""},
	{ID: "com-7", ParentID: "com-2"},
	{ID: "com-6", ParentID: "com-4"},
	{ID: "com-5", ParentID: "com-1"},
	{ID: "com-4", ParentID: "com-3"},
	{ID: "com-3", ParentID: "com-1"},
	{ID: "com-2", ParentID: ""},
	{ID: "com-1", ParentID: ""},
}

func TestBuild(t *testing.T) {
	want := len(comments)
	root, _ := Build(comments)
	count := traverseRoot(root)
	if count != want {
		t.Errorf("Build() error, len = %d, want %d", count, want)
	}
}

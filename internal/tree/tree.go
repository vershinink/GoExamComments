// Пакет tree содержит инструменты для создания и сортировки дерева комментариев.
package tree

import (
	"GoExamComments/internal/storage"
	"errors"
	"fmt"
)

// Root - структура дерева комментариев к посту.
type Root struct {
	Comments []*Node
}

// Node - структура узла дерева комментариев.
type Node struct {
	Comment storage.Comment `json:"comment"`
	Childs  []*Node         `json:"childs"`
}

var (
	// ErrEmptySlice - пустой слайс комментариев.
	ErrEmptySlice = errors.New("empty comment array")
)

// Build строит полное дерево комментариев из переданного слайса.
func Build(comments []storage.Comment) (Root, error) {
	const operation = "tree.Build"

	m := make(map[string]*Node)
	root := Root{}

	if len(comments) == 0 {
		return root, fmt.Errorf("%s: %w", operation, ErrEmptySlice)
	}

	for _, comment := range comments {
		node := &Node{Comment: comment}

		n, ok := m[comment.ID]
		if ok {
			node.Childs = n.Childs
		}
		m[comment.ID] = node

		if comment.ParentID == "" {
			root.Comments = append(root.Comments, node)
			continue
		}

		_, ok = m[comment.ParentID]
		if !ok {
			parent := &Node{}
			parent.Childs = append(parent.Childs, node)
			m[comment.ParentID] = parent
			continue
		}
		m[comment.ParentID].Childs = append(m[comment.ParentID].Childs, node)
	}

	return root, nil
}

// traverseRoot обходит все дерево комментариев в глубину. Выводит содержимое
// каждого комментария в stdout. Возвращает общее число посещенных узлов.
// Функция для отладки и тестирования.
func traverseRoot(root Root) int {
	count := traverse(root.Comments)
	return count
}

// traverse обходит переданный слайс комментариев и печатает содержимое в stdout.
// Рекурсивно вызывает саму себя на вложенных слайсах. Возвращает количество
// посещенных узлов. Функция для отладки и тестирования.
func traverse(arr []*Node) int {
	var count int
	for _, node := range arr {
		var c int
		fmt.Printf("%+v", node.Comment)
		count++
		if len(node.Childs) > 0 {
			c = traverse(node.Childs)
		}
		count += c
	}
	return count
}

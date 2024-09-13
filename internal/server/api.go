package server

import (
	"GoExamComments/internal/logger"
	"GoExamComments/internal/middleware"
	"GoExamComments/internal/storage"
	"GoExamComments/internal/tree"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

// AddComment записывает переданный в запросе комментарий в БД. В заголовках
// должен быть "Content-Type" со значением "application/json" в начале. Размер
// тела запроса ограничен 1 Мбайтом. Размер комментария не более 1000 символов.
func AddComment(ln int, st storage.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const operation = "server.AddComment"

		log := slog.Default().With(
			slog.String("op", operation),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Info("request to add comment")

		ct := r.Header.Get("Content-Type")
		media := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if media != "application/json" {
			log.Error("content-Type header is not application/json")
			http.Error(w, "Content-Type header is not application/json", http.StatusUnsupportedMediaType)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1048576)

		var comm storage.Comment
		err := json.NewDecoder(r.Body).Decode(&comm)
		if err != nil {
			log.Error("cannot decode request", logger.Err(err))
			http.Error(w, "cannot decode request", http.StatusBadRequest)
			return
		}
		log.Debug("request body decoded")

		if comm.Content == "" {
			log.Error("comment has empty content field")
			http.Error(w, "empty comment", http.StatusBadRequest)
			return
		}
		if len([]rune(comm.Content)) > ln {
			log.Error("comment content field has more than 1000 characters")
			http.Error(w, "the length of the comment must not exceed 1000 characters", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		id, err := st.AddComment(ctx, comm)
		if err != nil {
			log.Error("cannot add comment to DB", logger.Err(err))
			if errors.Is(err, storage.ErrIncorrectParentID) || errors.Is(err, storage.ErrIncorrectPostID) {
				http.Error(w, "incorrect data", http.StatusBadRequest)
				return
			}
			http.Error(w, "cannot add the comment", http.StatusInternalServerError)
			return
		}
		log.Debug("comment added to DB successfully", slog.String("id", id))

		w.WriteHeader(http.StatusCreated)
		log.Info("request served successfuly")
		log = nil
	}
}

// Comments записывает в ResponseWriter полное дерево комментариев по
// принятому ID поста.
func Comments(st storage.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const operation = "server.Comments"

		log := slog.Default().With(
			slog.String("op", operation),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Info("request to receive comments")

		id := r.PathValue("id")
		if id == "" {
			log.Error("empty post id")
			http.Error(w, "empty post id", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		comms, err := st.Comments(ctx, id)
		if err != nil {
			log.Error("cannot receive comments", logger.Err(err))
			if errors.Is(err, storage.ErrNoComments) {
				http.Error(w, "post id not found", http.StatusNotFound)
				return
			}
			if errors.Is(err, storage.ErrIncorrectPostID) {
				http.Error(w, "incorrect post id", http.StatusBadRequest)
				return
			}
			http.Error(w, "cannot receive comments", http.StatusInternalServerError)
			return
		}
		log.Debug("comments received successfully")

		root, err := tree.Build(comms)
		if err != nil {
			log.Error("cannot build comments tree", logger.Err(err))
			http.Error(w, "cannot receive comments", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		err = enc.Encode(root.Comments)
		if err != nil {
			log.Error("cannot encode comments", logger.Err(err))
			http.Error(w, "cannot encode comments", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		log.Info("request served successfuly")
		log = nil
	}
}

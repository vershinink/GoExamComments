package server

import (
	"GoExamComments/internal/storage"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

func AddComment(ln int, st storage.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const operation = "server.AddComment"

		slog.Info("new request to add comment")

		ct := r.Header.Get("Content-Type")
		media := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if media != "application/json" {
			slog.Error("content-Type header is not application/json", slog.String("op", operation))
			http.Error(w, "Content-Type header is not application/json", http.StatusUnsupportedMediaType)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1048576)

		var comm storage.Comment
		err := json.NewDecoder(r.Body).Decode(&comm)
		if err != nil {
			slog.Error("cannot decode request", slog.String("err", err.Error()), slog.String("op", operation))
			http.Error(w, "cannot decode request", http.StatusBadRequest)
			return
		}
		slog.Debug("request body decoded")

		if len([]rune(comm.Content)) > ln {
			slog.Error("comment content field has more than 1000 characters", slog.String("op", operation))
			http.Error(w, "the length of the comment must not exceed 1000 characters", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		err = st.AddComment(ctx, comm)
		if err != nil {
			slog.Error("cannot add comment to DB", slog.String("err", err.Error()), slog.String("op", operation))
			http.Error(w, "cannot add the comment", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func Comments(st storage.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const operation = "server.Comments"

		slog.Info("new request to receive comments")

		id := r.PathValue("id")
		if id == "" {
			slog.Error("empty post id", slog.String("op", operation))
			http.Error(w, "empty post id", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		comms, err := st.Comments(ctx, id)
		if err != nil {
			slog.Error("cannot receive comments", slog.String("err", err.Error()), slog.String("op", operation))
			http.Error(w, "cannot receive comments", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		err = enc.Encode(comms)
		if err != nil {
			slog.Error("cannot encode comments", slog.String("err", err.Error()), slog.String("op", operation))
			http.Error(w, "cannot encode comments", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
	}
}

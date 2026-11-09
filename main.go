package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Моковая БД: map, где ключ — ID задачи.
var (
	tasks = map[int]Task{
		1: {ID: 1, Title: "Посмотреть записи курса", Description: "Сделать заметки"},
		2: {ID: 2, Title: "Сделать домашку", Description: "Используя полученные знания"},
		3: {ID: 3, Title: "Выложить проект на GitHub", Description: "Передать ссылку с доступом"},
	}

	mu sync.RWMutex
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id должен быть числом"})
			return
		}

		mu.RLock()
		t, ok := tasks[id] // ok == false, если такого ключа нет
		mu.RUnlock()

		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "задача не найдена"})
			return
		}
		writeJSON(w, http.StatusOK, t)
	})

	mux.HandleFunc("POST /tasks", func(w http.ResponseWriter, r *http.Request) {
		var t Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"Некорректный JSON": err.Error()})
			return
		}

		mu.Lock()
		defer mu.Unlock()

		if _, exists := tasks[t.ID]; exists {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "Задача с таким ID уже существует"})
			return
		}

		tasks[t.ID] = t
		writeJSON(w, http.StatusCreated, t)
	})

	// Все маршруты регистрируем ДО запуска: ListenAndServe блокирует выполнение.
	log.Println("Сервер запущен: http://localhost:8080/tasks")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

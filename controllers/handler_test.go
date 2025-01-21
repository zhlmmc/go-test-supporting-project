package controllers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"go-test-supporting-project/controllers"
	"go-test-supporting-project/models"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockStore struct {
	createPostFunc func(post models.Post) (models.Post, error)
}

func (m *MockStore) CreatePost(post models.Post) (models.Post, error) {
	return m.createPostFunc(post)
}

func TestHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockStoreFunc  func(post models.Post) (models.Post, error)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:        "Success",
			requestBody: `{"title": "Valid Title", "body": "Valid Body"}`,
			mockStoreFunc: func(post models.Post) (models.Post, error) {
				post.Id = 1
				return post, nil
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   map[string]interface{}{"id": float64(1)},
		},
		{
			name:           "Invalid Title",
			requestBody:    `{"title": "Invalid@Title", "body": "Body"}`,
			mockStoreFunc:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "Invalid post title: title is required and only alpha-numeric characters and underscore are permitted in title"},
		},
		{
			name:        "Store Error",
			requestBody: `{"title": "Valid Title", "body": "Valid Body"}`,
			mockStoreFunc: func(post models.Post) (models.Post, error) {
				return models.Post{}, errors.New("database error")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "post create failed: database error"},
		},
		{
			name:           "Empty Request Body",
			requestBody:    `{}`,
			mockStoreFunc:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "Invalid post title: title is required and only alpha-numeric characters and underscore are permitted in title"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &MockStore{
				createPostFunc: tt.mockStoreFunc,
			}

			handler := controllers.Handler{
				Store: mockStore,
			}

			req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewBufferString(tt.requestBody))
			w := httptest.NewRecorder()

			handler.Create(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			body, _ := ioutil.ReadAll(resp.Body)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			var actualBody map[string]interface{}
			json.Unmarshal(body, &actualBody)

			if !compareJSONBodies(actualBody, tt.expectedBody) {
				t.Errorf("expected body %v, got %v", tt.expectedBody, actualBody)
			}
		})
	}
}

func TestHandler_Status(t *testing.T) {
	handler := controllers.Handler{}

	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	w := httptest.NewRecorder()

	handler.Status(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	expectedBody := "Status OK"
	if string(body) != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, string(body))
	}
}

func compareJSONBodies(actual, expected map[string]interface{}) bool {
	if len(actual) != len(expected) {
		return false
	}

	for key, expectedValue := range expected {
		actualValue, exists := actual[key]
		if !exists || actualValue != expectedValue {
			return false
		}
	}

	return true
}

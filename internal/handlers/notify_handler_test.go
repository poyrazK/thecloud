package httphandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	topicsPath      = "/notify/topics"
	subsPath        = "/notify/subscriptions"
	testTopicName   = "topic-1"
	testExampleURL2 = "http://example.com"
	subSuffix       = "/subscriptions"
	publishSuffix   = "/publish"
	notifyPathInvalid     = "/invalid"
)

type mockNotifyService struct {
	mock.Mock
}

func (m *mockNotifyService) CreateTopic(ctx context.Context, name string) (*domain.Topic, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Topic), args.Error(1)
}

func (m *mockNotifyService) ListTopics(ctx context.Context) ([]*domain.Topic, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Topic), args.Error(1)
}

func (m *mockNotifyService) DeleteTopic(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockNotifyService) Subscribe(ctx context.Context, topicID uuid.UUID, protocol domain.SubscriptionProtocol, endpoint string) (*domain.Subscription, error) {
	args := m.Called(ctx, topicID, protocol, endpoint)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Subscription), args.Error(1)
}

func (m *mockNotifyService) ListSubscriptions(ctx context.Context, topicID uuid.UUID) ([]*domain.Subscription, error) {
	args := m.Called(ctx, topicID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Subscription), args.Error(1)
}

func (m *mockNotifyService) Unsubscribe(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockNotifyService) Publish(ctx context.Context, topicID uuid.UUID, body string) error {
	args := m.Called(ctx, topicID, body)
	return args.Error(0)
}

func setupNotifyHandlerTest(_ *testing.T) (*mockNotifyService, *NotifyHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := new(mockNotifyService)
	handler := NewNotifyHandler(svc)
	r := gin.New()
	return svc, handler, r
}

func TestNotifyHandlerCreateTopic(t *testing.T) {
	svc, handler, r := setupNotifyHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST(topicsPath, handler.CreateTopic)

	topic := &domain.Topic{ID: uuid.New(), Name: testTopicName}
	svc.On("CreateTopic", mock.Anything, testTopicName).Return(topic, nil)

	body, err := json.Marshal(map[string]interface{}{"name": testTopicName})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", topicsPath, bytes.NewBuffer(body))
	assert.NoError(t, err)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestNotifyHandlerListTopics(t *testing.T) {
	svc, handler, r := setupNotifyHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET(topicsPath, handler.ListTopics)

	topics := []*domain.Topic{{ID: uuid.New(), Name: testTopicName}}
	svc.On("ListTopics", mock.Anything).Return(topics, nil)

	req := httptest.NewRequest(http.MethodGet, topicsPath, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNotifyHandlerDeleteTopic(t *testing.T) {
	svc, handler, r := setupNotifyHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.DELETE(topicsPath+"/:id", handler.DeleteTopic)

	id := uuid.New()
	svc.On("DeleteTopic", mock.Anything, id).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, topicsPath+"/"+id.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNotifyHandlerSubscribe(t *testing.T) {
	svc, handler, r := setupNotifyHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST(topicsPath+"/:id"+subSuffix, handler.Subscribe)

	id := uuid.New()
	sub := &domain.Subscription{ID: uuid.New(), TopicID: id, Endpoint: testExampleURL2}
	svc.On("Subscribe", mock.Anything, id, domain.SubscriptionProtocol("http"), testExampleURL2).Return(sub, nil)

	body, err := json.Marshal(map[string]interface{}{
		"protocol": "http",
		"endpoint": testExampleURL2,
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", topicsPath+"/"+id.String()+subSuffix, bytes.NewBuffer(body))
	assert.NoError(t, err)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestNotifyHandlerListSubscriptions(t *testing.T) {
	svc, handler, r := setupNotifyHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET(topicsPath+"/:id"+subSuffix, handler.ListSubscriptions)

	id := uuid.New()
	subs := []*domain.Subscription{{ID: uuid.New(), TopicID: id}}
	svc.On("ListSubscriptions", mock.Anything, id).Return(subs, nil)

	req := httptest.NewRequest(http.MethodGet, topicsPath+"/"+id.String()+subSuffix, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNotifyHandlerUnsubscribe(t *testing.T) {
	svc, handler, r := setupNotifyHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.DELETE(subsPath+"/:id", handler.Unsubscribe)

	id := uuid.New()
	svc.On("Unsubscribe", mock.Anything, id).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, subsPath+"/"+id.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNotifyHandlerPublish(t *testing.T) {
	svc, handler, r := setupNotifyHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST(topicsPath+"/:id"+publishSuffix, handler.Publish)

	id := uuid.New()
	svc.On("Publish", mock.Anything, id, "hello").Return(nil)

	body, err := json.Marshal(map[string]interface{}{"message": "hello"})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", topicsPath+"/"+id.String()+publishSuffix, bytes.NewBuffer(body))
	assert.NoError(t, err)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
func TestNotifyHandlerTopicErrors(t *testing.T) {
	t.Run("CreateInvalidJSON", func(t *testing.T) {
		_, handler, r := setupNotifyHandlerTest(t)
		r.POST(topicsPath, handler.CreateTopic)
		req, _ := http.NewRequest("POST", topicsPath, bytes.NewBufferString("invalid"))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateServiceError", func(t *testing.T) {
		svc, handler, r := setupNotifyHandlerTest(t)
		r.POST(topicsPath, handler.CreateTopic)
		svc.On("CreateTopic", mock.Anything, mock.Anything).Return(nil, errors.New(errors.Internal, "error"))
		body, _ := json.Marshal(map[string]interface{}{"name": "n"})
		req, _ := http.NewRequest("POST", topicsPath, bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("ListServiceError", func(t *testing.T) {
		svc, handler, r := setupNotifyHandlerTest(t)
		r.GET(topicsPath, handler.ListTopics)
		svc.On("ListTopics", mock.Anything).Return(nil, errors.New(errors.Internal, "error"))
		req, _ := http.NewRequest("GET", topicsPath, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("DeleteInvalidID", func(t *testing.T) {
		_, handler, r := setupNotifyHandlerTest(t)
		r.DELETE(topicsPath+"/:id", handler.DeleteTopic)
		req, _ := http.NewRequest("DELETE", topicsPath+notifyPathInvalid, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("DeleteServiceError", func(t *testing.T) {
		svc, handler, r := setupNotifyHandlerTest(t)
		r.DELETE(topicsPath+"/:id", handler.DeleteTopic)
		id := uuid.New()
		svc.On("DeleteTopic", mock.Anything, id).Return(errors.New(errors.Internal, "error"))
		req, _ := http.NewRequest("DELETE", topicsPath+"/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestNotifyHandlerSubscriptionErrors(t *testing.T) {
	t.Run("SubscribeInvalidID", func(t *testing.T) {
		_, handler, r := setupNotifyHandlerTest(t)
		r.POST(topicsPath+"/:id"+subSuffix, handler.Subscribe)
		req, _ := http.NewRequest("POST", topicsPath+notifyPathInvalid+subSuffix, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("SubscribeInvalidJSON", func(t *testing.T) {
		_, handler, r := setupNotifyHandlerTest(t)
		r.POST(topicsPath+"/:id"+subSuffix, handler.Subscribe)
		id := uuid.New()
		req, _ := http.NewRequest("POST", topicsPath+"/"+id.String()+subSuffix, bytes.NewBufferString("invalid"))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("SubscribeServiceError", func(t *testing.T) {
		svc, handler, r := setupNotifyHandlerTest(t)
		r.POST(topicsPath+"/:id"+subSuffix, handler.Subscribe)
		id := uuid.New()
		svc.On("Subscribe", mock.Anything, id, mock.Anything, mock.Anything).Return(nil, errors.New(errors.Internal, "error"))
		body, _ := json.Marshal(map[string]interface{}{"protocol": "http", "endpoint": "e"})
		req, _ := http.NewRequest("POST", topicsPath+"/"+id.String()+subSuffix, bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("ListSubsInvalidID", func(t *testing.T) {
		_, handler, r := setupNotifyHandlerTest(t)
		r.GET(topicsPath+"/:id"+subSuffix, handler.ListSubscriptions)
		req, _ := http.NewRequest("GET", topicsPath+notifyPathInvalid+subSuffix, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ListSubsServiceError", func(t *testing.T) {
		svc, handler, r := setupNotifyHandlerTest(t)
		r.GET(topicsPath+"/:id"+subSuffix, handler.ListSubscriptions)
		id := uuid.New()
		svc.On("ListSubscriptions", mock.Anything, id).Return(nil, errors.New(errors.Internal, "error"))
		req, _ := http.NewRequest("GET", topicsPath+"/"+id.String()+subSuffix, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("UnsubscribeInvalidID", func(t *testing.T) {
		_, handler, r := setupNotifyHandlerTest(t)
		r.DELETE(subsPath+"/:id", handler.Unsubscribe)
		req, _ := http.NewRequest("DELETE", subsPath+notifyPathInvalid, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UnsubscribeServiceError", func(t *testing.T) {
		svc, handler, r := setupNotifyHandlerTest(t)
		r.DELETE(subsPath+"/:id", handler.Unsubscribe)
		id := uuid.New()
		svc.On("Unsubscribe", mock.Anything, id).Return(errors.New(errors.Internal, "error"))
		req, _ := http.NewRequest("DELETE", subsPath+"/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestNotifyHandlerPublishErrors(t *testing.T) {
	t.Run("PublishInvalidID", func(t *testing.T) {
		_, handler, r := setupNotifyHandlerTest(t)
		r.POST(topicsPath+"/:id"+publishSuffix, handler.Publish)
		req, _ := http.NewRequest("POST", topicsPath+notifyPathInvalid+publishSuffix, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("PublishInvalidJSON", func(t *testing.T) {
		_, handler, r := setupNotifyHandlerTest(t)
		r.POST(topicsPath+"/:id"+publishSuffix, handler.Publish)
		id := uuid.New()
		req, _ := http.NewRequest("POST", topicsPath+"/"+id.String()+publishSuffix, bytes.NewBufferString("invalid"))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("PublishServiceError", func(t *testing.T) {
		svc, handler, r := setupNotifyHandlerTest(t)
		r.POST(topicsPath+"/:id"+publishSuffix, handler.Publish)
		id := uuid.New()
		svc.On("Publish", mock.Anything, id, mock.Anything).Return(errors.New(errors.Internal, "error"))
		body, _ := json.Marshal(map[string]interface{}{"message": "m"})
		req, _ := http.NewRequest("POST", topicsPath+"/"+id.String()+publishSuffix, bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

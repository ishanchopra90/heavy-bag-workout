package generator

import (
	"bytes"
	"errors"
	"heavybagworkout/internal/mocks"
	"io"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
)

func newHTTPResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}

func TestOpenAIClientGenerateWorkoutRequest_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockHTTPClient(ctrl)
	mockClient.EXPECT().
		Do(gomock.Any()).
		Return(newHTTPResponse(http.StatusOK, `{"choices":[{"message":{"content":"{\"rounds\":[]}"}}]}`), nil)

	client := NewOpenAIClientWithHTTPClient("test-key", mockClient)
	resp, err := client.GenerateWorkoutRequest("prompt")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp != "{\"rounds\":[]}" {
		t.Fatalf("unexpected response: %s", resp)
	}
}

func TestOpenAIClientGenerateWorkoutRequest_HTTPError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockHTTPClient(ctrl)
	mockClient.EXPECT().
		Do(gomock.Any()).
		Return(nil, errors.New("network failure"))

	client := NewOpenAIClientWithHTTPClient("test-key", mockClient)
	_, err := client.GenerateWorkoutRequest("prompt")
	if err == nil || err.Error() != "failed to send request: network failure" {
		t.Fatalf("expected network failure error, got %v", err)
	}
}

func TestOpenAIClientGenerateWorkoutRequest_StatusError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockHTTPClient(ctrl)
	mockClient.EXPECT().
		Do(gomock.Any()).
		Return(newHTTPResponse(http.StatusInternalServerError, "boom"), nil)

	client := NewOpenAIClientWithHTTPClient("test-key", mockClient)
	_, err := client.GenerateWorkoutRequest("prompt")
	if err == nil || err.Error() != "API error (status 500): boom" {
		t.Fatalf("expected status error, got %v", err)
	}
}

func TestOpenAIClientGenerateWorkoutRequest_APIErrorPayload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockHTTPClient(ctrl)
	mockClient.EXPECT().
		Do(gomock.Any()).
		Return(newHTTPResponse(http.StatusOK, `{"choices":[],"error":{"message":"oops"}}`), nil)

	client := NewOpenAIClientWithHTTPClient("test-key", mockClient)
	_, err := client.GenerateWorkoutRequest("prompt")
	if err == nil || err.Error() != "OpenAI API error: oops" {
		t.Fatalf("expected OpenAI API error, got %v", err)
	}
}

func TestOpenAIClientGenerateWorkoutRequest_NoChoices(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockHTTPClient(ctrl)
	mockClient.EXPECT().
		Do(gomock.Any()).
		Return(newHTTPResponse(http.StatusOK, `{"choices":[]}`), nil)

	client := NewOpenAIClientWithHTTPClient("test-key", mockClient)
	_, err := client.GenerateWorkoutRequest("prompt")
	if err == nil || err.Error() != "no choices in response" {
		t.Fatalf("expected no choices error, got %v", err)
	}
}

func TestOpenAIClientGenerateWorkoutRequest_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockHTTPClient(ctrl)
	mockClient.EXPECT().
		Do(gomock.Any()).
		Return(newHTTPResponse(http.StatusOK, `invalid`), nil)

	client := NewOpenAIClientWithHTTPClient("test-key", mockClient)
	_, err := client.GenerateWorkoutRequest("prompt")
	if err == nil || err.Error() != "failed to unmarshal response: invalid character 'i' looking for beginning of value" {
		t.Fatalf("expected json error, got %v", err)
	}
}

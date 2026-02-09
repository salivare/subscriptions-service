package subscription_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/salivare/subscriptions-service/tests/suite"
)

func TestUpdateSubscription_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	subID := st.CreateSubscription(t)

	updateReq := fmt.Sprintf(
		`{
            "service_name": "%s",
            "price": %d,
            "user_id": "%s",
            "start_date": "%s"
        }`,
		gofakeit.AppName(),
		int64(gofakeit.Number(100, 1000)),
		uuid.New().String(),
		suite.RandomMonth(),
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPatch,
		st.URL("/api/v1/subscription/"+subID),
		bytes.NewBufferString(updateReq),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := st.Client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestUpdateSubscription_InvalidUUIDInPath(t *testing.T) {
	ctx, st := suite.New(t)

	updateReq := fmt.Sprintf(
		`{
            "service_name": "%s",
            "price": %d,
            "user_id": "%s",
            "start_date": "%s"
        }`,
		gofakeit.AppName(),
		int64(gofakeit.Number(100, 1000)),
		uuid.New().String(),
		suite.RandomMonth(),
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPatch,
		st.URL("/api/v1/subscription/not-a-uuid"),
		bytes.NewBufferString(updateReq),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := st.Client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUpdateSubscription_InvalidUserID(t *testing.T) {
	ctx, st := suite.New(t)

	subID := st.CreateSubscription(t)

	updateReq := fmt.Sprintf(
		`{
            "service_name": "%s",
            "price": %d,
            "user_id": "broken-uuid",
            "start_date": "%s"
        }`,
		gofakeit.AppName(),
		int64(gofakeit.Number(100, 1000)),
		suite.RandomMonth(),
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPatch,
		st.URL("/api/v1/subscription/"+subID),
		bytes.NewBufferString(updateReq),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := st.Client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUpdateSubscription_InvalidDate(t *testing.T) {
	ctx, st := suite.New(t)

	subID := st.CreateSubscription(t)

	updateReq := fmt.Sprintf(
		`{
            "service_name": "%s",
            "price": %d,
            "user_id": "%s",
            "start_date": "2024-01"
        }`,
		gofakeit.AppName(),
		int64(gofakeit.Number(100, 1000)),
		uuid.New().String(),
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPatch,
		st.URL("/api/v1/subscription/"+subID),
		bytes.NewBufferString(updateReq),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := st.Client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUpdateSubscription_EmptyServiceName(t *testing.T) {
	ctx, st := suite.New(t)

	subID := st.CreateSubscription(t)

	updateReq := fmt.Sprintf(
		`{
            "service_name": "",
            "price": %d,
            "user_id": "%s",
            "start_date": "%s"
        }`,
		int64(gofakeit.Number(100, 1000)),
		uuid.New().String(),
		suite.RandomMonth(),
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPatch,
		st.URL("/api/v1/subscription/"+subID),
		bytes.NewBufferString(updateReq),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := st.Client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

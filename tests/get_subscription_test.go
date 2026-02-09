package subscription_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/salivare/subscriptions-service/tests/suite"
)

type GetResponse struct {
	Status string `json:"status"`
	Data   struct {
		ID          string  `json:"id"`
		ServiceName string  `json:"service_name"`
		Price       int64   `json:"price"`
		UserID      string  `json:"user_id"`
		StartDate   string  `json:"start_date"`
		EndDate     *string `json:"end_date"`
	} `json:"data"`
}

func TestGetSubscription_HappyPath(t *testing.T) {
	_, st := suite.New(t)

	subID := st.CreateSubscription(t)

	resp, err := st.Client.Get(
		st.URL("/api/v1/subscription/" + subID),
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var data GetResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&data))
	resp.Body.Close()

	assert.Equal(t, subID, data.Data.ID)
	assert.NotEmpty(t, data.Data.ServiceName)
	assert.NotEmpty(t, data.Data.UserID)
	assert.NotEmpty(t, data.Data.StartDate)
}

func TestGetSubscription_InvalidUUID(t *testing.T) {
	_, st := suite.New(t)

	resp, err := st.Client.Get(
		st.URL("/api/v1/subscription/not-a-uuid"),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetSubscription_NotFound(t *testing.T) {
	_, st := suite.New(t)

	randomID := uuid.New().String()

	resp, err := st.Client.Get(
		st.URL("/api/v1/subscription/" + randomID),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

package subscription_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/salivare/subscriptions-service/tests/suite"
)

type CreateResponse struct {
	Status string `json:"status"`
	Data   struct {
		ID string `json:"id"`
	} `json:"data"`
}

func TestSaveSubscription_HappyPath(t *testing.T) {
	_, st := suite.New(t)

	body := fmt.Sprintf(
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

	resp, err := st.Client.Post(
		st.URL("/api/v1/subscription"),
		"application/json",
		bytes.NewBufferString(body),
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var created CreateResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()

	assert.NotEmpty(t, created.Data.ID)
}

func TestSaveSubscription_InvalidUUID(t *testing.T) {
	_, st := suite.New(t)

	body := `{
        "service_name": "Netflix",
        "price": 500,
        "user_id": "not-a-uuid",
        "start_date": "01-2024"
    }`

	resp, err := st.Client.Post(
		st.URL("/api/v1/subscription"),
		"application/json",
		bytes.NewBufferString(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSaveSubscription_NegativePrice(t *testing.T) {
	_, st := suite.New(t)

	body := `{
        "service_name": "Netflix",
        "price": -10,
        "user_id": "d8a7c4f2-5c8b-4c6b-9b7e-2b9c1e5a1f33",
        "start_date": "01-2024"
    }`

	resp, err := st.Client.Post(
		st.URL("/api/v1/subscription"),
		"application/json",
		bytes.NewBufferString(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSaveSubscription_InvalidDate(t *testing.T) {
	_, st := suite.New(t)

	body := `{
        "service_name": "Netflix",
        "price": 500,
        "user_id": "d8a7c4f2-5c8b-4c6b-9b7e-2b9c1e5a1f33",
        "start_date": "2024-01"
    }`

	resp, err := st.Client.Post(
		st.URL("/api/v1/subscription"),
		"application/json",
		bytes.NewBufferString(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSaveSubscription_EmptyServiceName(t *testing.T) {
	_, st := suite.New(t)

	body := `{
        "service_name": "",
        "price": 500,
        "user_id": "d8a7c4f2-5c8b-4c6b-9b7e-2b9c1e5a1f33",
        "start_date": "01-2024"
    }`

	resp, err := st.Client.Post(
		st.URL("/api/v1/subscription"),
		"application/json",
		bytes.NewBufferString(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSaveSubscription_MissingFields(t *testing.T) {
	_, st := suite.New(t)

	body := `{}`

	resp, err := st.Client.Post(
		st.URL("/api/v1/subscription"),
		"application/json",
		bytes.NewBufferString(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

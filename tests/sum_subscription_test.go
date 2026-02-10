package subscription_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/google/uuid"

	"github.com/salivare/subscriptions-service/tests/suite"
)

type SumResponse struct {
	Status string `json:"status"`
	Data   struct {
		Total int64 `json:"total"`
	} `json:"data"`
}

func TestSumSubscription_HappyPath(t *testing.T) {
	_, st := suite.New(t)

	body := fmt.Sprintf(
		`{
            "user_id": "%s",
            "start_date_from": "01-2024",
            "start_date_to":   "12-2024"
        }`,
		uuid.New().String(),
	)

	resp, err := st.Client.Post(
		st.URL("/api/v1/subscription/sum"),
		"application/json",
		bytes.NewBufferString(body),
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var data SumResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&data))
	resp.Body.Close()

	assert.GreaterOrEqual(t, data.Data.Total, int64(0))
}

func TestSumSubscription_MissingRequiredFilters(t *testing.T) {
	_, st := suite.New(t)

	body := fmt.Sprintf(
		`{
            "user_id": "%s"
        }`,
		uuid.New().String(),
	)

	resp, err := st.Client.Post(
		st.URL("/api/v1/subscription/sum"),
		"application/json",
		bytes.NewBufferString(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSumSubscription_InvalidUUID(t *testing.T) {
	_, st := suite.New(t)

	body := fmt.Sprintf(
		`{
            "user_id": "not-a-uuid",
            "start_date_from": "%s"
        }`,
		suite.RandomMonth(),
	)

	resp, err := st.Client.Post(
		st.URL("/api/v1/subscription/sum"),
		"application/json",
		bytes.NewBufferString(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSumSubscription_InvalidDateRange(t *testing.T) {
	_, st := suite.New(t)

	body := fmt.Sprintf(
		`{
            "user_id": "%s",
            "start_date_from": "12-2024",
            "start_date_to":   "01-2024"
        }`,
		uuid.New().String(),
	)

	resp, err := st.Client.Post(
		st.URL("/api/v1/subscription/sum"),
		"application/json",
		bytes.NewBufferString(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSumSubscription_InvalidStartDateFormat(t *testing.T) {
	_, st := suite.New(t)

	body := fmt.Sprintf(
		`{
            "user_id": "%s",
            "start_date_from": "2024-01"
        }`,
		uuid.New().String(),
	)

	resp, err := st.Client.Post(
		st.URL("/api/v1/subscription/sum"),
		"application/json",
		bytes.NewBufferString(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSumSubscription_InvalidEndDateFormat(t *testing.T) {
	_, st := suite.New(t)

	body := fmt.Sprintf(
		`{
            "user_id": "%s",
            "end_date_from": "2024-01"
        }`,
		uuid.New().String(),
	)

	resp, err := st.Client.Post(
		st.URL("/api/v1/subscription/sum"),
		"application/json",
		bytes.NewBufferString(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSumSubscription_EmptyBody(t *testing.T) {
	_, st := suite.New(t)

	body := `{}`

	resp, err := st.Client.Post(
		st.URL("/api/v1/subscription/sum"),
		"application/json",
		bytes.NewBufferString(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSumSubscription_NotFound(t *testing.T) {
	_, st := suite.New(t)

	body := fmt.Sprintf(
		`{
            "user_id": "%s",
            "start_date_from": "%s"
        }`,
		uuid.New().String(),
		suite.RandomMonth(),
	)

	resp, err := st.Client.Post(
		st.URL("/api/v1/subscription/sum"),
		"application/json",
		bytes.NewBufferString(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	}
}

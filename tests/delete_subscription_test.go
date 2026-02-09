package subscription_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/salivare/subscriptions-service/tests/suite"
)

func TestDeleteSubscription_HappyPath(t *testing.T) {
	_, st := suite.New(t)

	subID := st.CreateSubscription(t)

	req, err := http.NewRequest(
		http.MethodDelete,
		st.URL("/api/v1/subscription/"+subID),
		nil,
	)
	require.NoError(t, err)

	resp, err := st.Client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDeleteSubscription_InvalidUUID(t *testing.T) {
	_, st := suite.New(t)

	req, err := http.NewRequest(
		http.MethodDelete,
		st.URL("/api/v1/subscription/not-a-uuid"),
		nil,
	)
	require.NoError(t, err)

	resp, err := st.Client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestDeleteSubscription_NotFound(t *testing.T) {
	_, st := suite.New(t)

	randomID := uuid.New().String()

	req, err := http.NewRequest(
		http.MethodDelete,
		st.URL("/api/v1/subscription/"+randomID),
		nil,
	)
	require.NoError(t, err)

	resp, err := st.Client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

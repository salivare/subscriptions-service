package suite

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type CreateResponse struct {
	Status string `json:"status"`
	Data   struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (s *Suite) CreateSubscription(t require.TestingT) string {
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
		RandomMonth(),
	)

	resp, err := s.Client.Post(
		s.URL("/api/v1/subscription"),
		"application/json",
		bytes.NewBufferString(body),
	)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var created CreateResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()

	require.NotEmpty(t, created.Data.ID)
	return created.Data.ID
}

func RandomMonth() string {
	month := gofakeit.Number(1, 12)
	year := 2024
	return fmt.Sprintf("%02d-%d", month, year)
}

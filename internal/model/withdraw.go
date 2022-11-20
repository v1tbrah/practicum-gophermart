package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type Withdraw struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

func (w Withdraw) MarshalJSON() ([]byte, error) {
	log.Debug().Msg("w.MarshalJSON START")
	defer log.Debug().Msg("w.MarshalJSON END")

	var strJSONWithdraw strings.Builder
	strJSONWithdraw.WriteString("{\"order\": \"" + w.Order + "\", \"sum\": " + fmt.Sprint(w.Sum) + ", \"processed_at\": \"" + w.ProcessedAt.Format(time.RFC3339) + "\"}")
	return []byte(strJSONWithdraw.String()), nil
}

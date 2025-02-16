package pkg

import "github.com/Te8va/MerchStore/internal/domain"

func FilterCoinHistory(history domain.CoinHistory) domain.CoinHistory {
	filteredHistory := domain.CoinHistory{}
	if history.Sent != nil {
		filteredSent := make([]domain.SentTransaction, 0, len(history.Sent))
		for _, tx := range history.Sent {
			if tx.ToUser != "" && tx.Amount > 0 {
				filteredSent = append(filteredSent, domain.SentTransaction{
					ToUser: tx.ToUser,
					Amount: tx.Amount,
				})
			}
		}
		if len(filteredSent) > 0 {
			filteredHistory.Sent = filteredSent
		}
	}

	if history.Received != nil {
		filteredReceived := make([]domain.ReceivedTransaction, 0, len(history.Received))
		for _, tx := range history.Received {
			if tx.FromUser != "" && tx.Amount > 0 {
				filteredReceived = append(filteredReceived, domain.ReceivedTransaction{
					FromUser: tx.FromUser,
					Amount:   tx.Amount,
				})
			}
		}
		if len(filteredReceived) > 0 {
			filteredHistory.Received = filteredReceived
		}
	}

	return filteredHistory
}

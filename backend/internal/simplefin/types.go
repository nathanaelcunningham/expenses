package simplefin

import (
	"time"
)

type AccountsResponse struct {
	Errors   []string  `json:"errors"`
	Accounts []Account `json:"accounts"`
}

type AccountResponse struct {
	Errors  []string `json:"errors"`
	Account Account  `json:"accounts"`
}

type AccountTransactionsRequest struct {
	AccountID string
	StartDate time.Time
	EndDate   time.Time
}

type AccountTransactionsResponse struct {
	Errors       []string       `json:"errors"`
	Transactions []Transactions `json:"transactions"`
}

type Account struct {
	Org              Org            `json:"org"`
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	Currency         string         `json:"currency"`
	Balance          string         `json:"balance"`
	AvailableBalance string         `json:"available-balance"`
	BalanceDate      int            `json:"balance-date"`
	Transactions     []Transactions `json:"transactions"`
	Extra            Extra          `json:"extra"`
}
type Org struct {
	Domain  string `json:"domain"`
	SfinURL string `json:"sfin-url"`
	Name    string `json:"name"`
}

type Transactions struct {
	ID          string `json:"id"`
	Posted      int    `json:"posted"`
	Amount      string `json:"amount"`
	Description string `json:"description"`
	Payee       string `json:"payee"`
}

type Extra struct {
	AccountOpenDate int `json:"account-open-date"`
}

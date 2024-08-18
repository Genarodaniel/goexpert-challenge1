package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type EconomiaResponse struct {
	Cotacao Cotacao `json:"USDBRL"`
}

type Cotacao struct {
	ID         int64  `gorm:"primaryKey" json:"-"`
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

func main() {
	db, err := gorm.Open(sqlite.Open("cotacoes.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&Cotacao{})

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cotacao" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		cotacao, err := CotaDolar(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		err = InsertCotacao(db, &cotacao.Cotacao)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			log.Println("Error while inserting result")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(cotacao)

	})

	http.ListenAndServe(":8080", nil)
}

func CotaDolar(w http.ResponseWriter) (*EconomiaResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	url := "https://economia.awesomeapi.com.br/json/last/USD-BRL"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Println("Error while creating request to dolar api")
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("Timeout while calling dolar api: %v", err)
			w.WriteHeader(http.StatusGatewayTimeout)
		} else {
			log.Printf("Internal Server Error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error while reading body in dolar api")
		return nil, err
	}

	var response EconomiaResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println("Request  dolaar Finalizada")
		return nil, err
	}

	return &response, err
}

func InsertCotacao(db *gorm.DB, cotacao *Cotacao) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()

	return db.WithContext(ctx).Create(cotacao).Error
}

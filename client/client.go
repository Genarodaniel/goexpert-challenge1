package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	USDBR struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {
	ctx := context.Background()
	dolarValue, err := cotaDolar(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(dolarValue)
	file, err := os.Create("cotacao.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao criar arquivo: %v\n", err)
	}
	defer file.Close()
	_, err = file.WriteString(fmt.Sprintf("Dólar: {%s}", *dolarValue))
	if err != nil {
		panic(err)
	}
	fmt.Println("Arquivo criado com sucesso!")

}

func cotaDolar(ctx context.Context) (*string, error) {
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()

	url := "http://localhost:8080/cotacao"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Println("Error while creating request to dolar api")
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("Timeout while calling dolar api: %v", err)
		} else {
			log.Printf("Internal Server Error: %v", err)
		}

		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error while reading body in dolar api")
		return nil, err
	}

	var response Cotacao
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println("Request dolar Finalizada")
		return nil, err
	}

	return &response.USDBR.Bid, err
}

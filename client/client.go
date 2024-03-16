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
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, error := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if error != nil {
		panic(error)
	}

	res, error := http.DefaultClient.Do(req)
	if error != nil {
		panic(error)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		resBody, error := io.ReadAll(res.Body)
		if error != nil {
			panic(error)
		}
		panic(string(resBody))
	}

	var cotacao Cotacao
	error = json.NewDecoder(res.Body).Decode(&cotacao)
	if error != nil {
		panic(error)
	}

	f, error := os.Create("cotacao.txt")
	if error != nil {
		log.Println("Criação do arquivo:", error)
		panic(error)
	}
	defer f.Close()

	f.Write([]byte(fmt.Sprintf("Dólar:{%v}", cotacao.Bid)))
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Cotacao struct {
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

type CotacaoData struct {
	Cotacao Cotacao `json:"USDBRL"`
}

type CotacaoResponse struct {
	Bid string `json:"bid"`
}

func main() {
	db, error := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if error != nil {
		panic(error)
	}
	db.AutoMigrate(&Cotacao{})

	cotacaoHandler := NewCotacaoHandler(db)

	http.HandleFunc("/cotacao", cotacaoHandler.getCotacao)

	error = http.ListenAndServe(":8080", nil)
	if error != nil {
		log.Fatal(error)
	}
}

type CotacaoHandler struct {
	db *gorm.DB
}

func NewCotacaoHandler(db *gorm.DB) *CotacaoHandler {
	return &CotacaoHandler{db: db}
}

func (c *CotacaoHandler) getCotacao(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := context.Background()
	ctxWeb, cancelWeb := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancelWeb()

	req, error := http.NewRequestWithContext(ctxWeb, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if error != nil {
		err := fmt.Errorf("erro ao criar requisição: %v", error.Error())
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, error := http.DefaultClient.Do(req)
	if error != nil {
		err := fmt.Errorf("erro ao buscar cotação: %v", error.Error())
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	var cotacao CotacaoData
	error = json.NewDecoder(res.Body).Decode(&cotacao)
	if error != nil {
		err := fmt.Errorf("erro ao parsear cotação: %v", error.Error())
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	error = c.saveCotacao(ctx, &cotacao.Cotacao)

	if error != nil {
		err := fmt.Errorf("erro ao salvar cotação: %v", error.Error())
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&CotacaoResponse{Bid: cotacao.Cotacao.Bid})
}

func (c *CotacaoHandler) saveCotacao(ctx context.Context, cotacao *Cotacao) error {
	ctxDb, cancelDb := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancelDb()

	error := c.db.WithContext(ctxDb).Create(cotacao).Error
	if error != nil {
		return error
	}

	return nil
}

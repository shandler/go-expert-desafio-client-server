package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"log"
	"net/http"

	"time"
)

func main() {
	client := &http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequest("GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Println("Erro ao criar a requisição:", err)
		return
	}

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		log.Println("Erro ao fazer a requisição:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Erro ao receber a resposta do servidor:", resp.Status)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Erro ao ler o corpo da resposta:", err)
		return
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		return
	}

	bid, ok := result["bid"].(string)
	if !ok {
		fmt.Println("Campo 'bid' não encontrado no JSON")
		return
	}

	err = os.WriteFile("cotacao.txt", []byte(fmt.Sprintf("Dólar: %s", bid)), 0644)
	if err != nil {
		fmt.Println("Erro ao salvar a cotação em 'cotacao.txt':", err)
		return
	}

	fmt.Println("Cotação do dólar salva em 'cotacao.txt':", bid)

}

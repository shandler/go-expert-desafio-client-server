package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "cotacoes.db")
	if err != nil {
		fmt.Println("Erro ao abrir o banco de dados:", err)
		return
	}
	defer db.Close()

	// Cria a tabela de cotações se ela não existir
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		data DATETIME,
		valor REAL
	)`)
	if err != nil {
		fmt.Println("Erro ao criar a tabela de cotações:", err)
		return
	}

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
		defer cancel()

		select {
		case <-ctx.Done():
			http.Error(w, "Timeout ao buscar cotação do dólar", http.StatusInternalServerError)
			fmt.Println("Timeout ao buscar cotação do dólar")
			return
		default:
			resp, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
			if err != nil {
				http.Error(w, "Erro ao buscar cotação do dólar", http.StatusInternalServerError)
				fmt.Println("Erro ao buscar cotação do dólar:", err)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				http.Error(w, "Erro ao ler resposta da API", http.StatusInternalServerError)
				fmt.Println("Erro ao ler resposta da API:", err)
				return
			}

			var result map[string]map[string]interface{}
			err = json.Unmarshal(body, &result)
			if err != nil {
				http.Error(w, "Erro ao decodificar JSON da API", http.StatusInternalServerError)
				fmt.Println("Erro ao decodificar JSON da API:", err)
				return
			}

			bid, ok := result["USDBRL"]["bid"].(string)
			if !ok {
				http.Error(w, "Campo 'bid' não encontrado no JSON da API", http.StatusInternalServerError)
				fmt.Println("Campo 'bid' não encontrado no JSON da API")
				return
			}

			// Registra a cotação no banco de dados
			ctx_insert, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			insertQuery := "INSERT INTO cotacoes (data, valor) VALUES (?, ?)"
			_, err = db.ExecContext(ctx_insert, insertQuery, time.Now(), bid)
			if err != nil {
				http.Error(w, "Erro ao inserir no banco de dados", http.StatusInternalServerError)
				log.Println("Erro ao inserir no banco de dados", http.StatusInternalServerError)
				return
			}

			// Retorna a cotação ao cliente
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"bid": bid})
		}
	})

	http.ListenAndServe(":8080", nil)
}

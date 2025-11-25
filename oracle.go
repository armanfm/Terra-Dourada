package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
)

// Estrutura do pacote vindo da sua UI
type OracleRequest struct {
	Event   string `json:"event"`
	Payload string `json:"payload"`
	Device  string `json:"device"`
}

// Estrutura da resposta do backend privado
type OracleResponse struct {
	Hash       string `json:"hash"`
	Timestamp  string `json:"timestamp"`
	LedgerLine string `json:"ledger_line"`
	Proof      string `json:"proof"`
	Status     string `json:"status"`
}

func main() {
	// URL do backend privado (NÃO aparece no GitHub)
	backendURL := os.Getenv("PRIVATE_ORACLE_URL")
	if backendURL == "" {
		log.Fatal("Erro: a variável PRIVATE_ORACLE_URL não está definida")
	}

	mux := http.NewServeMux()

	// Endpoint chamado pela UI
	mux.HandleFunc("/oracle", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método inválido", http.StatusMethodNotAllowed)
			return
		}

		// Receber JSON da UI
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Erro ao ler body", http.StatusBadRequest)
			return
		}

		// Encaminhar para o backend privado
		resp, err := http.Post(
			backendURL+"/oracle",
			"application/json",
			bytes.NewReader(body),
		)
		if err != nil {
			http.Error(w, "Erro ao contactar backend privado", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Ler resposta do backend privado
		privateResp, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Erro ao ler resposta backend privado", http.StatusBadGateway)
			return
		}

		// Replicar resposta para a UI
		w.Header().Set("Content-Type", "application/json")
		w.Write(privateResp)
	})

	// Libera CORS para frontend local ou remoto
	handler := corsMiddleware(mux)

	log.Println("🌐 Gateway Go rodando em http://localhost:7070")
	http.ListenAndServe(":7070", handler)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

		if r.Method == http.MethodOptions {
			return
		}

		next.ServeHTTP(w, r)
	})
}

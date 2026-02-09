package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

// Das ist der Code für den Scanner, der in die EXE kommt
const scannerCode = `
package main
import (
	"fmt"
	"os"
	"strings"
	"time"
)
var Detections string 
func main() {
	fmt.Println("--- CUSTOM SCANNER ---")
	searchList := strings.Split(Detections, ",")
	fmt.Printf("Scanne nach: %v\n", searchList)
	files, _ := os.ReadDir("./")
	found := false
	for _, file := range files {
		for _, detect := range searchList {
			if strings.Contains(strings.ToLower(file.Name()), strings.ToLower(detect)) {
				fmt.Printf("[TREFFER] Verdächtige Datei: %s\n", file.Name())
				found = true
			}
		}
	}
	if !found { fmt.Println("PC scheint sauber zu sein.") }
	fmt.Println("\nFenster schließt sich in 10 Sekunden...")
	time.Sleep(10 * time.Second)
}
`

func main() {
	// Erstelle die Scanner-Vorlage einmal beim Start
	ioutil.WriteFile("scanner_template.go", []byte(scannerCode), 0644)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Das HTML Formular
		fmt.Fprintf(w, `
			<html>
				<body style="font-family: sans-serif; text-align: center; padding: 50px;">
					<h1>Anti-Cheat Generator</h1>
					<form action="/download" method="POST">
						<input type="text" name="strings" placeholder="z.B. vape,reach,autoclicker" style="padding: 10px; width: 300px;">
						<br><br>
						<button type="submit" style="padding: 10px 20px; cursor: pointer;">EXE Erstellen & Downloaden</button>
					</form>
				</body>
			</html>
		`)
	})

	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" { http.Redirect(w, r, "/", 303); return }
		
		detects := r.FormValue("strings")
		
		// Hier passiert die Magie: Wir kompilieren für Windows (GOOS=windows)
		// Wir spritzen den String direkt in die Variable "Detections"
		cmd := exec.Command("go", "build", "-ldflags", "-X main.Detections="+detects, "-o", "scan.exe", "scanner_template.go")
		
		// WICHTIG: Damit es auf dem Linux-Server eine Windows-EXE baut
		cmd.Env = append(os.Environ(), "GOOS=windows", "GOARCH=amd64")
		
		output, err := cmd.CombinedOutput()
		if err != nil {
			http.Error(w, "Fehler beim Kompilieren: "+string(output), 500)
			return
		}

		// Datei zum Download anbieten
		w.Header().Set("Content-Disposition", "attachment; filename=scan.exe")
		http.ServeFile(w, r, "scan.exe")
	})

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	fmt.Println("Server startet auf Port " + port)
	http.ListenAndServe(":"+port, nil)
}

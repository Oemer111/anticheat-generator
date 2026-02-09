package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// --- DATENSTRUKTUREN ---

// Eine Detection-Regel, die du selbst erstellst
type CustomRule struct {
	Name      string `json:"name"`      // z.B. "SSTB Bypass"
	Detection string `json:"detection"` // z.B. "autoclicker_v2"
	Type      string `json:"type"`      // z.B. "String"
}

type ScanResult struct {
	ID         string
	Time       string
	Username   string
	RiskLevel  string
	Detections []string
}

// Globaler Speicher (RAM)
var (
	GlobalRules = []CustomRule{} // Hier landen deine Custom Detections
	GlobalScans = []ScanResult{}
	DataMutex   sync.Mutex
)

// --- DER SCANNER CODE (Der in die EXE kommt) ---
const scannerCode = `
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"strings"
	"time"
	"net/http"
	"io/ioutil"
)

// Diese Variablen werden vom Server beim Download gefüllt
var ServerURL string = "ERSETZE_MICH" 
var RulesJSON string = "[]" // Hier stecken deine Custom Rules drin

// Design
const (
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorWhite  = "\033[97m"
	ColorGray   = "\033[90m"
	Reset       = "\033[0m"
)

type CustomRule struct {
	Name      string ` + "`json:\"name\"`" + `
	Detection string ` + "`json:\"detection\"`" + `
	Type      string ` + "`json:\"type\"`" + `
}

type ReportPayload struct {
	Username   string   ` + "`json:\"username\"`" + `
	Detections []string ` + "`json:\"detections\"`" + `
}

func main() {
	// 1. UI Starten
	fmt.Print("\033[H\033[2J")
	fmt.Println(ColorRed + ` + "`" + `
  _____  ______  _____  ________  _   _  ______ 
 |  __ \|  ____||  __ \|___  /  \| \ | ||  ____|
 | |__) | |__   | |  | |  / /| |  \| | || |__   
 |  _  /|  __|  | |  | | / / | | .   | ||  __|  
 | | \ \| |____ | |__| |/ /__| | |\  | || |____ 
 |_|  \_\______||_____//_____|_|_| \_|_||______|
` + "`" + ` + Reset)
	
	fmt.Println(ColorGray + "\nLoading Custom Definitions..." + Reset)
	
	// 2. Deine Regeln laden
	var rules []CustomRule
	json.Unmarshal([]byte(RulesJSON), &rules)
	fmt.Printf("Loaded Rules: " + ColorWhite + "%d\n" + Reset, len(rules))

	currentUser, _ := user.Current()
	username := "Unknown"
	if currentUser != nil { username = currentUser.Username }

	fmt.Printf("Scanning Target: " + ColorWhite + "%s\n" + Reset, username)
	fmt.Println(ColorGray + "--------------------------------------------------" + Reset)

	foundThreats := []string{}

	// 3. Der Scan-Loop
	files, _ := os.ReadDir("./")
	
	for _, f := range files {
		// Wir prüfen JEDE deiner Regeln gegen JEDE Datei
		for _, rule := range rules {
			// Check 1: Dateiname
			if strings.Contains(strings.ToLower(f.Name()), strings.ToLower(rule.Detection)) {
				fmt.Printf(ColorRed + "[DETECTED] " + ColorWhite + "%s " + ColorGray + "-> Rule: [%s]\n" + Reset, f.Name(), rule.Name)
				foundThreats = append(foundThreats, rule.Name + " (File: " + f.Name() + ")")
			}

			// Check 2: Datei Inhalt (Einfacher String Scan)
			if !f.IsDir() {
				content, err := ioutil.ReadFile(f.Name())
				if err == nil {
					if strings.Contains(string(content), rule.Detection) {
						fmt.Printf(ColorRed + "[MEMORY/STRING] " + ColorWhite + "Found '%s' in %s " + ColorGray + "-> Rule: [%s]\n" + Reset, rule.Detection, f.Name(), rule.Name)
						foundThreats = append(foundThreats, rule.Name + " (In File: " + f.Name() + ")")
					}
				}
			}
		}
	}

	// 4. Report Senden
	if len(foundThreats) > 0 {
		fmt.Println("\n" + ColorRed + "THREATS IDENTIFIED. UPLOADING REPORT..." + Reset)
	} else {
		fmt.Println("\n" + ColorGreen + "SYSTEM CLEAN. UPLOADING LOGS..." + Reset)
	}
	
	payload := ReportPayload{ Username: username, Detections: foundThreats }
	jsonData, _ := json.Marshal(payload)
	http.Post(ServerURL + "/api/report", "application/json", bytes.NewBuffer(jsonData))
	
	time.Sleep(5 * time.Second)
}
`

// --- WEBSITE CODE ---
const htmlTemplate = `
<!DOCTYPE html>
<html lang="de" class="h-full bg-zinc-950">
<head>
    <meta charset="UTF-8">
    <title>REDZONE | Panel</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script>
        tailwind.config = { theme: { extend: { colors: { zinc: { 900: '#18181b', 950: '#09090b' }, red: { 600: '#dc2626' } } } } }
    </script>
    <style>
        .active-tab { background-color: rgba(220, 38, 38, 0.1); color: #dc2626; border-right: 2px solid #dc2626; }
        input, select { background: #09090b; border: 1px solid #27272a; color: white; padding: 8px; border-radius: 4px; }
        input:focus { outline: none; border-color: #dc2626; }
    </style>
</head>
<body class="h-full flex font-sans text-zinc-400 overflow-hidden">

    <div class="w-64 bg-zinc-900 flex flex-col border-r border-zinc-800 flex-shrink-0">
        <div class="h-16 flex items-center px-6 border-b border-zinc-800">
            <h1 class="text-2xl font-black italic tracking-widest text-white">RED<span class="text-red-600">ZONE</span></h1>
        </div>
        <nav class="flex-1 py-6 space-y-1">
            <button onclick="showTab('scans')" id="btn-scans" class="w-full flex items-center px-6 py-3 hover:bg-zinc-800 transition text-left active-tab">SCANS</button>
            <button onclick="showTab('custom')" id="btn-custom" class="w-full flex items-center px-6 py-3 hover:bg-zinc-800 transition text-left">CUSTOM DETECTIONS</button>
        </nav>
        <div class="p-4 border-t border-zinc-800">
            <form action="/download" method="POST">
                <button class="w-full bg-red-600 hover:bg-red-700 text-white font-bold py-3 rounded shadow-lg shadow-red-900/20 transition">EXE ERSTELLEN</button>
            </form>
        </div>
    </div>

    <div class="flex-1 overflow-y-auto bg-zinc-950 p-8">
        
        <div id="tab-scans" class="space-y-6">
            <h2 class="text-xl font-bold text-white mb-4">Live Ergebnisse</h2>
            <div class="rounded-lg border border-zinc-800 overflow-hidden">
                <table class="min-w-full bg-zinc-900">
                    <thead class="bg-zinc-800/50 text-zinc-500 uppercase text-xs">
                        <tr><th class="px-6 py-3 text-left">User</th><th class="px-6 py-3 text-left">Status</th><th class="px-6 py-3 text-left">Details</th></tr>
                    </thead>
                    <tbody class="divide-y divide-zinc-800">
                        {{range .Scans}}
                        <tr>
                            <td class="px-6 py-4 font-bold text-white">{{.Username}}</td>
                            <td class="px-6 py-4">
                                {{if eq .RiskLevel "DETECTED"}}<span class="text-red-500 bg-red-900/20 px-2 py-1 rounded border border-red-900">DETECTED</span>
                                {{else}}<span class="text-green-500 bg-green-900/20 px-2 py-1 rounded border border-green-900">CLEAN</span>{{end}}
                            </td>
                            <td class="px-6 py-4 text-sm font-mono text-zinc-400">{{.Detections}}</td>
                        </tr>
                        {{else}}
                        <tr><td colspan="3" class="px-6 py-10 text-center text-zinc-600">Keine Scans vorhanden.</td></tr>
                        {{end}}
                    </tbody>
                </table>
            </div>
        </div>

        <div id="tab-custom" class="hidden space-y-6">
             <div class="flex justify-between items-center mb-6">
                <h2 class="text-xl font-bold text-white">Deine Module</h2>
             </div>

             <div class="bg-zinc-900 p-6 rounded-xl border border-zinc-800 mb-8">
                <h3 class="text-white font-bold mb-4">Neue Detection hinzufügen</h3>
                <form action="/add_rule" method="POST" class="flex gap-4 items-end">
                    <div class="flex-1">
                        <label class="block text-xs mb-1">Name des Cheats (UI)</label>
                        <input type="text" name="name" placeholder="z.B. Vape V4" class="w-full">
                    </div>
                    <div class="flex-1">
                        <label class="block text-xs mb-1">Erkennungs-String</label>
                        <input type="text" name="detection" placeholder="z.B. vape_wrapper.dll" class="w-full">
                    </div>
                    <button class="bg-zinc-800 hover:bg-zinc-700 text-white font-bold py-2 px-6 rounded border border-zinc-700">+ Add</button>
                </form>
             </div>

             <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                {{range .Rules}}
                <div class="bg-zinc-900 rounded-xl border border-zinc-800 overflow-hidden group hover:border-red-600/50 transition relative">
                    <div class="h-24 bg-zinc-950 flex items-center justify-center border-b border-zinc-800">
                        <h3 class="text-lg font-bold text-white">{{.Name}}</h3>
                    </div>
                    <div class="p-4">
                        <div class="flex justify-between text-xs mb-2">
                            <span class="text-zinc-500">TYPE</span>
                            <span class="text-red-500">Active</span>
                        </div>
                        <code class="block bg-zinc-950 p-2 rounded text-xs text-zinc-400 break-all">{{.Detection}}</code>
                    </div>
                </div>
                {{else}}
                <div class="col-span-3 text-center py-12 border-2 border-dashed border-zinc-800 rounded-xl">
                    <p class="text-zinc-500">Noch keine Custom Detections erstellt.</p>
                </div>
                {{end}}
             </div>
        </div>

    </div>

    <script>
        function showTab(name) {
            ['scans', 'custom'].forEach(t => {
                document.getElementById('tab-' + t).classList.add('hidden');
                document.getElementById('btn-' + t).classList.remove('active-tab');
            });
            document.getElementById('tab-' + name).classList.remove('hidden');
            document.getElementById('btn-' + name).classList.add('active-tab');
        }
    </script>
</body>
</html>
`

func main() {
	ioutil.WriteFile("scanner_template.go", []byte(scannerCode), 0644)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		DataMutex.Lock()
		data := struct {
			Rules []CustomRule
			Scans []ScanResult
		}{Rules: GlobalRules, Scans: GlobalScans}
		DataMutex.Unlock()
		template.Must(template.New("x").Parse(htmlTemplate)).Execute(w, data)
	})

	// NEUE REGEL HINZUFÜGEN
	http.HandleFunc("/add_rule", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			name := r.FormValue("name")
			detect := r.FormValue("detection")
			if name != "" && detect != "" {
				DataMutex.Lock()
				GlobalRules = append(GlobalRules, CustomRule{Name: name, Detection: detect, Type: "String"})
				DataMutex.Unlock()
			}
		}
		http.Redirect(w, r, "/#custom", 303)
	})

	http.HandleFunc("/api/report", func(w http.ResponseWriter, r *http.Request) {
		var p struct { Username string; Detections []string }
		json.NewDecoder(r.Body).Decode(&p)
		
		risk := "CLEAN"
		if len(p.Detections) > 0 { risk = "DETECTED" }

		DataMutex.Lock()
		GlobalScans = append([]ScanResult{{
			Username: p.Username, 
			RiskLevel: risk, 
			Detections: p.Detections, 
			Time: time.Now().Format("15:04"),
		}}, GlobalScans...)
		DataMutex.Unlock()
	})

	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		host := "https://" + r.Host
		if strings.Contains(r.Host, "localhost") { host = "http://" + r.Host }

		// WICHTIG: Wir wandeln die Regeln in JSON um und brennen sie in die EXE
		DataMutex.Lock()
		jsonBytes, _ := json.Marshal(GlobalRules)
		DataMutex.Unlock()
		
		// Escape Anführungszeichen für den Command Line Befehl
		jsonString := strings.ReplaceAll(string(jsonBytes), "\"", "\\\"")

		cmd := exec.Command("go", "build", 
			"-ldflags", fmt.Sprintf("-X main.ServerURL=%s -X \"main.RulesJSON=%s\"", host, jsonString), 
			"-o", "redzone_client.exe", "scanner_template.go")
		
		cmd.Env = append(os.Environ(), "GOOS=windows", "GOARCH=amd64")
		cmd.Run()

		w.Header().Set("Content-Disposition", "attachment; filename=redzone_client.exe")
		http.ServeFile(w, r, "redzone_client.exe")
	})

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	fmt.Println("Server läuft auf Port " + port)
	http.ListenAndServe(":"+port, nil)
}

package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// --- NEUES EXE DESIGN (Konsolen-basiert) ---
// Wir nutzen ANSI-Farbcodes und ASCII-Art für den Look.
const scannerCode = `
package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Farbcodes für die Konsole
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorWhite  = "\033[97m"
	ColorGray   = "\033[90m"
	bgBlack     = "\033[40m"
)

var Detections string

func main() {
	// Konsole "säubern" und Header anzeigen
	fmt.Print("\033[H\033[2J") // Clear Screen
	fmt.Println(bgBlack + ColorRed + ` + "`" + `
██████╗ ███████╗██████╗ ███████╗ ██████╗ ███╗   ██╗███████╗
██╔══██╗██╔════╝██╔══██╗╚══███╔╝██╔═══██╗████╗  ██║██╔════╝
██████╔╝█████╗  ██║  ██║  ███╔╝ ██║   ██║██╔██╗ ██║█████╗  
██╔══██╗██╔════╝██║  ██║ ███╔╝  ██║   ██║██║╚██╗██║██╔════╝
██║  ██║███████╗██████╔╝███████╗╚██████╔╝██║ ╚████║███████╗
╚═╝  ╚═╝╚══════╝╚═════╝ ╚══════╝ ╚═════╝ ╚═╝  ╚═══╝╚══════╝
` + "`" + ` + ColorReset)
	fmt.Println(ColorGray + "---------------------------------------------------------" + ColorReset)
	fmt.Println(ColorWhite + "                SYSTEM SCAN INITIATED" + ColorReset)
	fmt.Println(ColorGray + "---------------------------------------------------------" + ColorReset)
	fmt.Println("")

	searchList := strings.Split(Detections, ",")
	if len(Detections) == 0 {
		fmt.Println(ColorRed + "[!] ERROR: No detection strings loaded." + ColorReset)
		time.Sleep(5 * time.Second)
		return
	}

	fmt.Printf(ColorGray+"Loaded Signatures: "+ColorWhite+"%d\n"+ColorReset, len(searchList))
	fmt.Println(ColorGray + "Scanning current directory..." + ColorReset)
	fmt.Println("")

	files, err := os.ReadDir("./")
	if err != nil {
		fmt.Println(ColorRed + "[!] Error reading directory." + ColorReset)
		return
	}

	foundCount := 0
	for _, file := range files {
		for _, detect := range searchList {
			if detect != "" && strings.Contains(strings.ToLower(file.Name()), strings.ToLower(detect)) {
				fmt.Printf(" "+ColorRed+"[DETECTED] "+ColorWhite+"%s "+ColorGray+"(Signature: %s)\n"+ColorReset, file.Name(), detect)
				foundCount++
				break // Ein Treffer pro Datei reicht
			}
		}
	}

	fmt.Println("")
	fmt.Println(ColorGray + "---------------------------------------------------------" + ColorReset)
	if foundCount == 0 {
		fmt.Println("            " + ColorWhite + "STATUS: " + ColorRed + "CLEAN" + ColorReset)
	} else {
		fmt.Printf("            "+ColorWhite+"STATUS: "+ColorRed+"THREATS FOUND (%d)\n"+ColorReset, foundCount)
	}
	fmt.Println(ColorGray + "---------------------------------------------------------" + ColorReset)

	fmt.Println(ColorGray + "\nClosing in 10 seconds..." + ColorReset)
	time.Sleep(10 * time.Second)
}
`

func main() {
	ioutil.WriteFile("scanner_template.go", []byte(scannerCode), 0644)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// --- NEUES WEBSITE DESIGN (HTML + Tailwind CSS) ---
		// Wir definieren die Farben direkt in der Config, um sie überall zu nutzen.
		// bg-zinc-950 = Fast Schwarz (Hintergrund)
		// bg-zinc-900 = Sehr dunkles Grau (Sidebar/Karten)
		// red-600 = Unser Akzent-Rot
		html := `
<!DOCTYPE html>
<html lang="de" class="h-full bg-zinc-950">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>REDZONE - Panel</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script>
        tailwind.config = {
            theme: {
                extend: {
                    colors: {
                        zinc: { 900: '#18181b', 950: '#09090b' },
                        red: { 600: '#dc2626', 700: '#b91c1c' }
                    }
                }
            }
        }
    </script>
    <style>
        /* Verstecke Scrollbars für cleaneren Look */
        ::-webkit-scrollbar { width: 8px; }
        ::-webkit-scrollbar-track { background: #09090b; }
        ::-webkit-scrollbar-thumb { background: #dc2626; border-radius: 4px; }
    </style>
</head>
<body class="h-full flex font-sans text-zinc-300">

    <div class="w-64 bg-zinc-900 flex flex-col border-r border-zinc-800">
        <div class="h-16 flex items-center px-6 border-b border-zinc-800">
            <h1 class="text-2xl font-black tracking-wider text-red-600">REDZONE</h1>
        </div>
        <nav class="flex-1 py-6 px-4 space-y-2 overflow-y-auto">
            <a href="#" class="flex items-center px-4 py-3 text-sm font-medium rounded-lg bg-red-600/10 text-red-600">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" /></svg>
                String Ersteller
            </a>
            <a href="#" class="flex items-center px-4 py-3 text-sm font-medium rounded-lg hover:bg-zinc-800 transition-colors">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-3 text-zinc-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                Benutzerdefinierte Erk.
            </a>
            </nav>
        <div class="p-4 border-t border-zinc-800">
            <div class="flex items-center">
                 <div class="w-8 h-8 rounded-full bg-red-600 flex items-center justify-center font-bold text-white">A</div>
                 <div class="ml-3">
                     <p class="text-sm font-medium text-white">Admin</p>
                     <p class="text-xs text-zinc-500">Administrator</p>
                 </div>
            </div>
        </div>
    </div>

    <div class="flex-1 flex flex-col overflow-hidden">
        <header class="h-16 bg-zinc-900 border-b border-zinc-800 flex items-center justify-between px-8">
            <h2 class="text-lg font-semibold text-white">String Builder</h2>
        </header>

        <main class="flex-1 overflow-y-auto p-8">
            
            <div class="mb-12">
                <div class="flex justify-between items-center mb-6">
                    <div>
                         <span class="inline-block px-4 py-2 text-sm font-medium rounded-l-lg bg-red-600 text-white cursor-pointer">Your Custom Strings</span>
                         <span class="inline-block px-4 py-2 text-sm font-medium rounded-r-lg bg-zinc-800 text-zinc-400 cursor-not-allowed">Enterprise Strings</span>
                    </div>
                </div>

                <div class="bg-zinc-900 rounded-xl border border-zinc-800 p-8 flex flex-col items-center justify-center min-h-[400px]">
                    <div class="text-center max-w-md">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 mx-auto text-zinc-700 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 002-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                        </svg>
                        <h3 class="text-xl font-bold text-white mb-2">Erstelle deine erste Detection</h3>
                        <p class="text-zinc-500 mb-8">Füge Strings hinzu, nach denen der Scanner suchen soll. Die EXE wird live generiert.</p>
                        
                        <form action="/download" method="POST" class="flex gap-2">
                            <input type="text" name="strings" placeholder="z.B. vape,reach,autoclicker.exe" class="flex-1 bg-zinc-950 border border-zinc-700 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-red-600 transition-colors">
                            <button type="submit" class="bg-red-600 hover:bg-red-700 text-white font-bold py-3 px-6 rounded-lg transition-colors flex items-center">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" /></svg>
                                EXE Generieren
                            </button>
                        </form>
                    </div>
                </div>
            </div>

            <div>
                 <h3 class="text-lg font-semibold text-white mb-4 flex items-center">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                    Deine Detections (Vorschau)
                 </h3>
                 <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                    <div class="bg-zinc-900/50 rounded-xl border-2 border-dashed border-zinc-700 hover:border-red-600 flex items-center justify-center h-48 cursor-pointer transition-colors group">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 text-zinc-700 group-hover:text-red-600 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
                    </div>
                    <div class="bg-zinc-900 rounded-xl border border-zinc-800 overflow-hidden hover:border-red-600/50 transition-colors">
                        <div class="h-32 bg-zinc-950 flex items-center justify-center relative overflow-hidden">
                             <div class="absolute inset-0 bg-gradient-to-br from-red-600/20 to-transparent opacity-50"></div>
                             <h4 class="text-xl font-bold text-white relative z-10">MINECRAFT KILLAURA</h4>
                        </div>
                        <div class="p-4 flex justify-between items-center bg-zinc-900">
                            <div class="flex space-x-2">
                                <span class="text-xs px-2 py-1 bg-red-600/10 text-red-600 rounded-md">Memory</span>
                                <span class="text-xs px-2 py-1 bg-zinc-800 text-zinc-400 rounded-md">V1.2</span>
                            </div>
                             <div class="flex space-x-3 text-zinc-500">
                                 <button class="hover:text-red-600"><svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" /></svg></button>
                                 <button class="hover:text-red-600"><svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg></button>
                             </div>
                        </div>
                    </div>
                 </div>
             </div>

        </main>
    </div>
</body>
</html>
`
		fmt.Fprint(w, html)
	})

	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", 303)
			return
		}

		detects := r.FormValue("strings")
		// Verhindere leere Strings, die alles matchen würden
		if strings.TrimSpace(detects) == "" {
			http.Error(w, "Fehler: Bitte gib mindestens einen String ein.", 400)
			return
		}

		// Wir bauen die Windows EXE und injecten die Strings
		cmd := exec.Command("go", "build", "-ldflags", "-X main.Detections="+detects, "-o", "redzone_scanner.exe", "scanner_template.go")
		cmd.Env = append(os.Environ(), "GOOS=windows", "GOARCH=amd64")

		output, err := cmd.CombinedOutput()
		if err != nil {
			// Zeige Build-Fehler im Browser an (hilfreich fürs Debugging)
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(500)
			fmt.Fprintf(w, "Build Error:\n%s", output)
			return
		}

		w.Header().Set("Content-Disposition", "attachment; filename=redzone_scanner.exe")
		http.ServeFile(w, r, "redzone_scanner.exe")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("REDZONE Server startet auf Port " + port)
	http.ListenAndServe(":"+port, nil)
}

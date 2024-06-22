package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

var proxy *httputil.ReverseProxy

func vending_send(w http.ResponseWriter, r *http.Request) {
	//access flippo database
	data := loadData()
	department := r.URL.Path[len("/vending/"):]
	departments, ok := data["dirDatabase"].(map[string]interface{})["department"].(map[string]interface{})
	if !ok {
		http.Error(w, "Internal Server Error: department data format incorrect", http.StatusInternalServerError)
		return
	}
	departmentData, ok := departments[department].(map[string]interface{})
	if !ok {
		http.Error(w, fmt.Sprintf("No data found for department: %s", department), http.StatusNotFound)
		return
	}
	var responseItems = make(map[string]map[string]interface{})
	for key, value := range departmentData {
		itemDetails, ok := value.(map[string]interface{})
		if !ok {
			continue
		}
		price := int(itemDetails["price"].(float64))
		desc := itemDetails["desc"].(string)
		path := itemDetails["index"].(string)
		responseItems[key] = map[string]interface{}{
			"keyName": key,
			"price":   price,
			"desc":    desc,
			"path":    path,
		}
	}
	jsonResponse, err := json.Marshal(responseItems)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
func solicit_offer(w http.ResponseWriter, r *http.Request) {
	keyName := r.URL.Path[len("/solicit/"):]
	data := loadData()
	departments, ok := data["dirDatabase"].(map[string]interface{})["department"].(map[string]interface{})
	if !ok {
		http.Error(w, "Internal Server Error: department data format incorrect", http.StatusInternalServerError)
		return
	}
	for _, departmentData := range departments {
		item, ok := departmentData.(map[string]interface{})[keyName].(map[string]interface{})
		if ok {
			vendor := item["vendor"].(string)
			vendorData := data["vendorIndex"].(map[string]interface{})[vendor].(map[string]interface{})
			hdpub := vendorData["hdpub"].(string)

			address, err := generateAddress(hdpub)
			if err != nil {
				log.Printf("Error generating address: %v", err)
				http.Error(w, "Error generating address", http.StatusInternalServerError)
				return
			}

			catalog("vendor/"+vendor, "new_address: "+address)

			skellykey := data["dirDatabase"].(map[string]interface{})["skellykey"].(string)
			constructedString := fmt.Sprintf("%s %s %s", address, vendor, skellykey)
			hasher := sha256.New()
			hasher.Write([]byte(constructedString))
			fullHash := hasher.Sum(nil)
			hexhash := hex.EncodeToString(fullHash)
			hashed := hexhash[:32]

			response := map[string]interface{}{
				"address": address,
				"hash":    hashed,
			}

			jsonResponse, err := json.Marshal(response)
			if err != nil {
				http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonResponse)
			return
		}
	}
	http.Error(w, "Item not found", http.StatusNotFound)
}
func dvrmod_send(w http.ResponseWriter, r *http.Request) {
	// Load data from JSON file
	data := loadData() // Ensure loadData() is defined to read from 'data.json'
	dirDatabase, ok := data["dirDatabase"].(map[string]interface{})
	if !ok {
		http.Error(w, "Internal Server Error: dirDatabase format incorrect", http.StatusInternalServerError)
		return
	}
	// Extract the module path from the request URL
	modulePath := r.URL.Path[len("/dvrmod/"):]
	moduleFilePath := "module/" + modulePath
	log.Printf("%s request serving: %s", r.Method, moduleFilePath)

	// Check for permit conditions
	permits, ok := dirDatabase["permit"].(map[string]interface{})
	if !ok {
		log.Println("No permit conditions found")
	} else {
		if permitDetails, ok := permits[modulePath].(map[string]interface{}); ok {
			log.Printf("Permit details for %s: %v", modulePath, permitDetails)
			// Permit condition exists, check for POST and body content
			if r.Method != http.MethodPost {
				http.Error(w, "Forbidden: POST request required", http.StatusForbidden)
				return
			}

			// Read and check the request body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusInternalServerError)
				return
			}
			var requestData map[string]interface{}
			err = json.Unmarshal(body, &requestData)
			if err != nil {
				http.Error(w, "Error parsing request body", http.StatusInternalServerError)
				return
			}
			// pull out receipt variables
			txid, ok := requestData["tx"].(string)
			if !ok {
				http.Error(w, "Invalid client data", http.StatusBadRequest)
				return
			}
			// research tx
			log.Printf("transaction details: %v", txid)
			txDetails, err := fetchTX(txid)
			if err != nil {
				log.Printf("Error fetching transaction details: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			extractHash, address, err := extractTXdata(txDetails)
			if err != nil {
				log.Printf("Error extracting data: %v", err)
				http.Error(w, "Error extracting data", http.StatusInternalServerError)
				return
			}
			//identify vendor
			vendor, ok := permitDetails["vendor"].(string)
			if !ok {
				http.Error(w, "Invalid client data", http.StatusBadRequest)
				return
			}
			skellykey := dirDatabase["skellykey"].(string)
			// recreate the hash
			constructedString := fmt.Sprintf("%s %s %s", address, vendor, skellykey)
			hasher := sha256.New()
			hasher.Write([]byte(constructedString))
			fullHash := hasher.Sum(nil)
			hexhash := hex.EncodeToString(fullHash)
			hashed := hexhash[:32]

			log.Printf("Extracted vendor: %s", vendor)
			log.Printf("Extracted address: %s", address)
			log.Printf("Extracted data: %s", extractHash)
			log.Printf("reconstructed hash: %s", hashed)
			logEntry(fmt.Sprintf("valid receipt for vendor %s = %s ", vendor, txid))
			//compare to verify transaction
			if extractHash != hashed {
				http.Error(w, "Forbidden: Hash mismatch", http.StatusForbidden)
				return
			}
			//add logic for receipt use count via log
		}
	}
	// Serve the file based on the module path
	http.ServeFile(w, r, moduleFilePath)
}
func dvrbox_send(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	//unpack request json
	var requestData map[string]interface{}
	err = json.Unmarshal(body, &requestData)
	if err != nil {
		http.Error(w, "Error parsing request body", http.StatusInternalServerError)
		return
	}
	//unpack request data
	client, ok := requestData["client"].(map[string]interface{})
	if !ok {
		http.Error(w, "Invalid client data", http.StatusBadRequest)
		return
	}
	var auxs []string
	auxInterface, ok := client["aux"].([]interface{})
	if !ok {
		http.Error(w, "aux data is missing or not a list", http.StatusBadRequest)
		return
	}
	for _, aux := range auxInterface {
		auxStr, ok := aux.(string)
		if !ok {
			http.Error(w, "aux contains non-string elements", http.StatusBadRequest)
			return
		}
		auxs = append(auxs, auxStr)
	}
	hostName, ok := client["href"].(string)
	if !ok {
		http.Error(w, "Host data is missing or not a string", http.StatusBadRequest)
		return
	}
	// Process database auxs and hostnames
	data := loadData()
	dirDatabase, ok := data["dirDatabase"].(map[string]interface{})
	if !ok {
		http.Error(w, "Internal Server Error: dirDatabase format incorrect", http.StatusInternalServerError)
		return
	}
	auxMap, ok := dirDatabase["aux"].(map[string]interface{})
	if !ok {
		log.Fatal("Expected 'aux' to be a map[string]interface{}")
	}
	hostsMap, ok := dirDatabase["hosts"].(map[string]interface{})
	if !ok {
		log.Fatal("Expected 'hosts' to be a map[string]interface{}")
	}
	//collect urls
	var urls []string
	for _, aux := range auxs {
		if dirNames, ok := auxMap[aux].([]interface{}); ok {
			for _, dirName := range dirNames {
				dirNameStr, ok := dirName.(string)
				if !ok {
					log.Printf("Expected directory name to be a string, got: %v", dirName)
					continue
				}
				urls = append(urls, "/flippo/"+dirNameStr)
			}
		}
	}
	hostEntries, ok := hostsMap[hostName].(map[string]interface{})
	if ok {
		// Process host entries
		for _, aux := range auxs {
			if dirNames, ok := hostEntries["aux"].(map[string]interface{})[aux].([]interface{}); ok {
				for _, dirName := range dirNames {
					dirNameStr, ok := dirName.(string)
					if !ok {
						log.Printf("Expected directory name to be a string, got: %v", dirName)
						continue
					}
					urls = append(urls, "/flippo/dvrmod/"+dirNameStr)
				}
			}
		}
	}

	log.Printf("dvrbox urls: %s", urls)
	// Convert the list of URLs to JSON and send it back to the client
	jsonResponse, err := json.Marshal(urls)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func serve(w http.ResponseWriter, r *http.Request) {
	data := loadData()
	dirDatabase, ok := data["dirDatabase"].(map[string]interface{})
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	skellykey := dirDatabase["skellykey"].(string)

	pathSegments := strings.Split(r.URL.Path[1:], "/") // Split path and remove the leading slash
	// Navigate through the nested JSON structure
	currentData := data["index"].(map[string]interface{})
	for _, segment := range pathSegments {
		if nextData, ok := currentData["/"+segment].(map[string]interface{}); ok {
			currentData = nextData
		} else {
			http.Error(w, "Item not found", http.StatusNotFound)
			return
		}
	}

	// Check the fee at the final segment
	fee, ok := currentData["fee"].(float64)
	if !ok {
		http.Error(w, "Fee data missing or incorrect format", http.StatusBadRequest)
		return
	}

	if fee == 0 {
		filePath := r.URL.Path[1:]
		if strings.HasSuffix(filePath, ".php") {
			proxy.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filePath)
	} else {
		httxid := r.Header.Get("httx")
		if httxid == "" {
			queryParams := r.URL.Query()
			httxid = queryParams.Get("httxid")
			if httxid == "" {
				log.Printf("httx: no header no queryparam")
				http.Error(w, "HTTX is required", http.StatusBadRequest)
				return
			}
		}
		// Use the httxHeader for processing
		log.Printf("Handling non-zero fee for %s", r.URL.Path)
		txDetails, err := fetchTX(httxid)
		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		extractHash, address, err := extractTXdata(txDetails)
		if err != nil {
			log.Printf("Error extracting data: %v", err)
			http.Error(w, "Error extracting data", http.StatusInternalServerError)
			return
		}
		//identify vendor
		vendor, ok := currentData["vendor"].(string)
		if !ok {
			http.Error(w, "Forbidden: unidentified httx vendor", http.StatusBadRequest)
			return
		}
		// recreate the hash
		constructedString := fmt.Sprintf("%s %s %s", address, vendor, skellykey)
		hasher := sha256.New()
		hasher.Write([]byte(constructedString))
		fullHash := hasher.Sum(nil)
		hexhash := hex.EncodeToString(fullHash)
		hashed := hexhash[:32]
		//compare to verify transaction
		if extractHash != hashed {
			log.Printf("hash mismatch: %s vs %s", hashed, extractHash)
			http.Error(w, "Forbidden: httx hash mismatch", http.StatusForbidden)
			return
		}

		//survived check

		filePath := r.URL.Path[1:]
		if strings.HasSuffix(filePath, ".php") {
			proxy.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filePath)
	}
}

func main() {

	initLogger()
	logEntry(fmt.Sprintf("Flippo was activated! %s", time.Now()))

	apacheServerURL, err := url.Parse("http://localhost:3301")
	if err != nil {
		log.Fatal(err)
	}
	proxy = httputil.NewSingleHostReverseProxy(apacheServerURL)

	router := http.NewServeMux()
	router.HandleFunc("/command/", atc_com)
	//atc tower simulation for mud style clients
	router.HandleFunc("/dvrbox/", dvrbox_send)
	//a request is received for a list of imports
	router.HandleFunc("/dvrmod/", dvrmod_send)
	//a request is received for import of module
	router.HandleFunc("/vending/", vending_send)
	//a request for a list of contents from department
	router.HandleFunc("/solicit/", solicit_offer)
	//a request for an offer to be generated for a program

	router.HandleFunc("/", serve)
	//default to attempting to serve file via route

	fmt.Println("FLIPPO transponder is active")
	catalog("flippo", "transponder is active")
	log.Fatal(http.ListenAndServe(":8081", router))
	defer close(logEntries)
}

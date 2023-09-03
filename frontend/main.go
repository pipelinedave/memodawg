package main

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
)

func main() {
	http.HandleFunc("/", formHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Starting frontend server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a request.")
	if r.Method == "POST" {
		log.Println("POST request received.")

		// Parsing form
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			log.Println("Failed to parse form: ", err)
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}
		log.Println("Form parsed successfully.")

		var b bytes.Buffer
		wr := multipart.NewWriter(&b)

		// Processing form fields
		for key, val := range r.PostForm {
			for _, v := range val {
				fw, err := wr.CreateFormField(key)
				if err != nil {
					log.Println("Failed to create form field: ", err)
					http.Error(w, "Unable to parse form", http.StatusBadRequest)
					return
				}
				if _, err = fw.Write([]byte(v)); err != nil {
					log.Println("Failed to write to form field: ", err)
					http.Error(w, "Unable to parse form", http.StatusBadRequest)
					return
				}
			}
		}

		log.Println("Form fields processed.")

		// Processing file
		file, _, err := r.FormFile("file")
		if err != nil {
			log.Println("Failed to get file: ", err)
			http.Error(w, "Invalid file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		fw, err := wr.CreateFormFile("file", "file")
		if err != nil {
			log.Println("Failed to create form file: ", err)
			http.Error(w, "Unable to add file to form", http.StatusInternalServerError)
			return
		}
		io.Copy(fw, file)

		log.Println("File processed.")

		// Close writer
		if err := wr.Close(); err != nil {
			log.Println("Failed to close writer: ", err)
			http.Error(w, "Unable to add file to form", http.StatusInternalServerError)
			return
		}

		// Creating request
		client := &http.Client{}
		req, err := http.NewRequest("POST", "https://memodawg.stillon.top/transcribe", &b)
		if err != nil {
			log.Println("Failed to create request: ", err)
			http.Error(w, "Error creating request", http.StatusInternalServerError)
			return
		}

		// Adding headers
		req.Header.Set("Content-Type", wr.FormDataContentType())
		req.Header.Set("X-API-Key", r.FormValue("api_key"))

		log.Printf("Headers set: %v", req.Header)

		// Making request
		resp, err := client.Do(req)
		if err != nil {
			log.Println("Failed to make request: ", err)
			http.Error(w, "Error making request", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Reading and displaying response
		respBody, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Received response: %v", string(respBody))
		w.Write(respBody)

	} else {
		log.Println("Serving HTML form.")
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		if err := tmpl.Execute(w, nil); err != nil {
			log.Println("Failed to execute template: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

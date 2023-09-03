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
	if r.Method == "POST" {
		// // Check the API key first
		// suppliedKey := r.FormValue("api_key") // Assuming the API key is sent in the form as "api_key"

		// if suppliedKey != os.Getenv("MEMODAWG_KEY") {
		// 	http.Error(w, "Invalid API key", http.StatusUnauthorized)
		// 	return
		// }

		// Parse the form data to get the file
		err := r.ParseMultipartForm(10 << 20) // limit your maxMultipartMemory
		if err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		// Prepare a form that you will submit to that URL.
		var b bytes.Buffer
		wr := multipart.NewWriter(&b)
		for key, val := range r.PostForm {
			for _, v := range val {
				fw, err := wr.CreateFormField(key)
				if err != nil {
					http.Error(w, "Unable to parse form", http.StatusBadRequest)
					return
				}
				if _, err = fw.Write([]byte(v)); err != nil {
					http.Error(w, "Unable to parse form", http.StatusBadRequest)
					return
				}
			}
		}

		// Add the file to the form
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Invalid file", http.StatusBadRequest)
			return
		}
		defer file.Close()
		fw, err := wr.CreateFormFile("file", "file")
		if err != nil {
			http.Error(w, "Unable to add file to form", http.StatusInternalServerError)
			return
		}
		io.Copy(fw, file)

		// Close the multipart writer so that the form data io.Reader is complete.
		if err := wr.Close(); err != nil {
			http.Error(w, "Unable to add file to form", http.StatusInternalServerError)
			return
		}

		// Create client and request
		client := &http.Client{}
		req, err := http.NewRequest("POST", "https://memodawg.stillon.top/transcribe", &b)
		if err != nil {
			http.Error(w, "Error creating request", http.StatusInternalServerError)
			return
		}

		// Set headers
		req.Header.Set("Content-Type", wr.FormDataContentType())
		req.Header.Set("X-API-Key", r.FormValue("api_key"))

		// Make the request
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Error making request", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Read and display the response
		respBody, _ := ioutil.ReadAll(resp.Body)
		w.Write(respBody)

	} else {
		// Serve the HTML form
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

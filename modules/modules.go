package modules

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"net/http"

	"github.com/golang/gddo/httputil/header"
	"github.com/julienschmidt/httprouter"
)

const (
	// filename  = "data/CSV_data_in.csv"
	// shortForm = "2006-01-02" // shortform declares short form of standard time
	// table elements
	trs     = "<tr>"
	trsHref = "<tr data-href>"
	trc     = "</tr>"
	tds     = "<td>"
	tdc     = "</td>"
	tdsd    = "<td scope=\"row\" id=\"td-dark\">"
	thsd    = "<th scope=\"col\">"
	ths     = "<th>"
	thc     = "</th>"
	inps    = "<input type=\"text\" name="
	inpc    = "\" class=\"form-control\">"
	scs     = "<script>"
	scc     = "</script>"
)

// <input type="text" name="fname" id="fname" class="form-control">
// malformedRequest struct for handling JSON errors
type malformedRequest struct {
	status int
	msg    string
}

// Controller struct for template controller
type Controller struct {
	tpl *template.Template
}

// CSVData struct for csv data
type CSVData struct {
	RC [][]string
}

// jsRow struct for js row addition
type jsRow struct {
	Cell       template.JS
	CellNameJS template.JS
	CellName   string
	EndMarker  template.JS
}

// Data csv data for use
var (
	Data CSVData
	//Name stores the uplaod filename
	Name string
	//ExportName stores the download filename
	ExportName string
	//FullName stores the full download path
	FullName string
	// ColCount counts the columns
	ColCount int
	// FileError to indecate error condition on file read
	FileError string
)

// NewController provides new controller for template processing
func NewController(t *template.Template) *Controller {
	return &Controller{t}
}

// Main provides a GET ROUTE for the root page
func (c Controller) Main(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	//
	// populate the template struct with empty values
	templateData := struct {
		Empty    bool
		FileName string
	}{
		Empty:    !check(),
		FileName: Name,
	}
	c.tpl.ExecuteTemplate(w, "landing.gohtml", templateData)
}

// Index GET ROUTE dislays form for csv.
func (c Controller) Index(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	// Check if file was loaded
	if check() {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	fmt.Println("Find ROUTE activated")
	// Create the row header slice
	tempColI := ""
	tempInputI := ""
	ColHX := []template.HTML{}
	InpHX := []template.HTML{}
	rowCount := len(Data.RC)
	for i := 0; i < ColCount+1; i++ {
		// build the headers
		tempColI = thsd
		if i == 0 { //for the first line
			tempColI += "R"
			tempColI += thc
			ColHX = append(ColHX, template.HTML(tempColI))
			tempInputI = inps
			tempInputI += "\"rc"
			tempInputI += fmt.Sprint(i + 1)
			tempInputI += "\" id=\"rc"
			tempInputI += fmt.Sprint(i + 1)
			tempInputI += "\""
			tempInputI += " readonly>"
			InpHX = append(InpHX, template.HTML(tempInputI))
			tempInputI = ""
		} else { //for the rest
			tempColI += fmt.Sprint(i)
			tempColI += thc
			ColHX = append(ColHX, template.HTML(tempColI))
			tempInputI = inps
			tempInputI += "\"rc"
			tempInputI += fmt.Sprint(i + 1)
			tempInputI += "\" id=\"rc"
			tempInputI += fmt.Sprint(i + 1)
			tempInputI += "\""
			tempInputI += ">"
			InpHX = append(InpHX, template.HTML(tempInputI))
			tempInputI = ""
		}
	}
	// js template format
	// cell1 = newRow.insertCell(0) -  New cells inside new row
	// <input type="text" name="fname" id="fname" class="form-control">

	// build the rows of the table
	tempI := ""
	RowIX := []template.HTML{}
	for row := 0; row < rowCount; row++ { //for all table rows
		tempI = trs
		tempI += tdsd
		tempI += fmt.Sprint(row + 1)
		columnCount := len(Data.RC[row])
		// check if column was added
		if columnCount != ColCount { // column was added
			newColumnCount := columnCount + 1
			for col := 0; col < newColumnCount; col++ { //for all table columns
				//check if last column
				if col < columnCount { // Not last column
					tempI += tds
					tempI += Data.RC[row][col]
					tempI += tdc
				} else { //last column
					tempI += tds
					tempI += ""
					tempI += tdc
				}
			}
			tempI += trc
			RowIX = append(RowIX, template.HTML(tempI)) //add to the row slice
			tempI = ""
		} else { // no column added
			for col := 0; col < columnCount; col++ {
				tempI += tds
				tempI += Data.RC[row][col]
				tempI += tdc
			}
			tempI += trc
			RowIX = append(RowIX, template.HTML(tempI)) //add to the row slice
			tempI = ""
		}
	}
	// build the input row
	s := buildInputSelection(ColCount)
	// build the row add cells & row
	js, jsR := newRow(ColCount)

	// populate the template struct with empty values
	templateData := struct {
		ColHX    []template.HTML
		RowIX    []template.HTML
		InpHX    []template.HTML
		JSIX     []template.JS
		JSRX     []jsRow
		Script   []string
		ColCount int
		FileName string
	}{
		ColHX:    ColHX,
		RowIX:    RowIX,
		InpHX:    InpHX,
		JSIX:     js,
		JSRX:     jsR,
		Script:   s,
		ColCount: ColCount + 2,
		FileName: Name,
	}
	c.tpl.ExecuteTemplate(w, "main.gohtml", templateData)

}

// IndexV GET ROUTE displays view for csv file.
func (c Controller) IndexV(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	// Check if file was loaded
	if check() {
		// check if error on loading file
		if FileError != "" {
			templateData := struct {
				Empty     bool
				FileName  string
				FileError string
			}{
				Empty:     !check(),
				FileName:  Name,
				FileError: FileError,
			}
			c.tpl.ExecuteTemplate(w, "error.gohtml", templateData)
		}
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	fmt.Println("Find ROUTE activated")
	// Create the row header slice
	tempColI := ""
	tempInputI := ""
	ColHX := []template.HTML{}
	InpHX := []template.HTML{}
	ColCount = len(Data.RC[0])
	rowCount := len(Data.RC)
	for i := 0; i < ColCount+1; i++ {
		// build the headers
		tempColI = thsd
		if i == 0 {
			tempColI += "R"
			tempColI += thc
			ColHX = append(ColHX, template.HTML(tempColI))
			tempInputI = inps
			tempInputI += "\"rc"
			tempInputI += fmt.Sprint(i + 1)
			tempInputI += "\" id=\"rc"
			tempInputI += fmt.Sprint(i + 1)
			tempInputI += "\""
			tempInputI += " readonly>"
			InpHX = append(InpHX, template.HTML(tempInputI))
			tempInputI = ""
		} else {
			tempColI += fmt.Sprint(i)
			tempColI += thc
			ColHX = append(ColHX, template.HTML(tempColI))
			tempInputI = inps
			tempInputI += "\"rc"
			tempInputI += fmt.Sprint(i + 1)
			tempInputI += "\" id=\"rc"
			tempInputI += fmt.Sprint(i + 1)
			tempInputI += "\""
			tempInputI += ">"
			InpHX = append(InpHX, template.HTML(tempInputI))
			tempInputI = ""
		}
	}

	tempI := ""
	RowIX := []template.HTML{}
	for row := 0; row < rowCount; row++ {
		tempI = trs
		tempI += tdsd
		tempI += fmt.Sprint(row + 1)
		for col := 0; col < len(Data.RC[row]); col++ {
			tempI += tds
			tempI += Data.RC[row][col]
			tempI += tdc
		}
		tempI += trc
		RowIX = append(RowIX, template.HTML(tempI))
		tempI = ""
	}
	// build the input row
	s := buildInputSelection(ColCount)
	// build the row add cells & row
	js, jsR := newRow(ColCount)
	// populate the template struct with empty values
	templateData := struct {
		ColHX    []template.HTML
		RowIX    []template.HTML
		InpHX    []template.HTML
		JSIX     []template.JS
		JSRX     []jsRow
		Script   []string
		ColCount int
		FileName string
	}{
		ColHX:    ColHX,
		RowIX:    RowIX,
		InpHX:    InpHX,
		JSIX:     js,
		JSRX:     jsR,
		Script:   s,
		ColCount: ColCount + 2,
		FileName: Name,
	}
	c.tpl.ExecuteTemplate(w, "mainView.gohtml", templateData)

}

// AJAX requests
// *******************************************

// AddNewRecord adds a new record from AJAX request
func (c Controller) AddNewRecord(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	type dataIn struct {
		Row []string `json:"r"`
	}
	fmt.Println("AddNewRecord activated")
	var t dataIn
	err := json.NewDecoder(req.Body).Decode(&t)
	if err != nil {
		log.Println(err)
	}
	log.Println(t)
	tmpI := make([]string, len(t.Row)-1)  //make a new empty inner slice
	copy(tmpI, t.Row[1:])                 //copy new row without 1st item to inner slice
	tmp := make([][]string, len(Data.RC)) //make a new empty outer slice
	copy(tmp, Data.RC)                    //copy current rows outer slice
	tmp = append(tmp, tmpI)               //add inner slice to outer
	Data.RC = tmp                         //put temp slice back as Data
	fmt.Println(Data.RC)
	msg, jErr := json.Marshal("Added!")
	if jErr != nil {
		log.Printf("Cannot marshal msg: %v\n", jErr)
	}
	_, wErr := w.Write(msg)
	if wErr != nil {
		log.Println("Cannot write message", wErr)
	}
	return
}

// AddNewColumn adds a new column from AJAX request
func (c Controller) AddNewColumn(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	fmt.Println("AddNewColumn activated")
	ColCount++
	msg, jErr := json.Marshal("Column added!")
	if jErr != nil {
		log.Printf("Cannot marshal msg: %v\n", jErr)
	}
	_, wErr := w.Write(msg)
	if wErr != nil {
		log.Println("Cannot write message", wErr)
	}
	return
}

// DeleteRecord POST ROUTE deletes a row on AJAX request.
func (c Controller) DeleteRecord(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	type dataD struct {
		Row string `json:"index"`
	}
	fmt.Println("Remove Record activated")
	var t dataD
	err := json.NewDecoder(req.Body).Decode(&t)
	if err != nil {
		log.Println(err)
	}
	log.Println(t)
	// Find csv record
	index, cErr := strconv.Atoi(t.Row)
	if cErr != nil {
		log.Printf("Cannot make index %v", cErr)
	}
	//Delete csv record
	tmpXX := [][]string{}

	tmpXX = removeIndex(Data.RC, index)
	Data.RC = Data.RC[:0]
	Data.RC = tmpXX
	fmt.Println("Total rows remaining ", len(Data.RC))

	//marshal json for response to client
	msg, jErr := json.Marshal("Deleted!")
	if jErr != nil {
		log.Printf("Cannot marshal msg: %v\n", jErr)
	}
	_, wErr := w.Write(msg)
	if wErr != nil {
		log.Println("Cannot write message", wErr)
	}
	return
}

// EditRecord POST ROUTE edits a row on AJAX request.
func (c Controller) EditRecord(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	type dataIn struct {
		Row []string `json:"r"`
	}
	fmt.Println("Edit Record activated")
	var t dataIn
	err := json.NewDecoder(req.Body).Decode(&t)
	if err != nil {
		log.Println(err)
	}
	log.Println(t)
	// Find csv record
	index, cErr := strconv.Atoi(t.Row[0])
	if cErr != nil {
		log.Printf("Cannot make index %v", cErr)
	}
	index--
	// update the CSV record with AJAX data
	Data.RC[index] = t.Row[1:]
	fmt.Printf("New values: %v\n ", Data.RC[index])
	//marshal json for response to client
	msg, jErr := json.Marshal("Updated!")
	if jErr != nil {
		log.Printf("Cannot marshal msg: %v\n", jErr)
	}
	_, wErr := w.Write(msg)
	if wErr != nil {
		log.Println("Cannot write message", wErr)
	}
	return
}

// FindCSV POST ROUTE finds the schedule.
func (c Controller) FindCSV(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	fmt.Println("Selection ROUTE activated")
	// If POST request
	if req.Method == http.MethodPost {
		// Get fields from the form
		req.ParseMultipartForm(32 << 20) // limit your max input length!
		var buf bytes.Buffer
		file, header, err := req.FormFile("Upload")
		if err != nil {
			log.Printf("Cannot Upload CSV file: %v\n", err)
		}
		defer file.Close()
		name := strings.Split(header.Filename, ".")
		Name = name[0]
		fmt.Printf("File name %s\n", Name)
		// Copy the file data to my buffer
		io.Copy(&buf, file)
		// Do something with the contents...
		contents := buf.String()
		Data, err = readCSV(contents) // Read the contents of the csv file
		if err != nil {
			FileError = fmt.Sprintf("Error occured: %v", err)
		}
		buf.Reset()
		http.Redirect(w, req, "/mainV", http.StatusSeeOther) // Display the contents in web page
		return
	}
	http.Redirect(w, req, "/index", http.StatusSeeOther)
}

// SaveCSV GET ROUTE Saves the CSV file.
func (c Controller) SaveCSV(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	// Check if file was loaded
	if check() {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	fmt.Println("Save ROUTE activated")
	// Populate the template struct with empty values
	templateData := struct {
		Empty    bool
		FileName string
	}{
		Empty:    !check(),
		FileName: Name,
	}
	c.tpl.ExecuteTemplate(w, "saving.gohtml", templateData)
}

// SaveRecord POST ROUTE reads export filename.
func (c Controller) SaveRecord(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	fmt.Println("Save Record activated")
	// If POST request
	if req.Method == http.MethodPost {
		// Get fields from the form
		ExportName = req.FormValue("SaveName")
		// add path & create file in data dir
		FullName = filepath.Base(ExportName)
		// Checkint if csv extention was added
		if !strings.ContainsAny(FullName, ".") {
			FullName += ".csv"
		}
		// *************************************************************

		// Local file save

		// CSVFile, wErr := os.Create(filepath.Join("data", FullName))
		// if wErr != nil {
		// 	log.Printf("Cannot open file %v", wErr)
		// }
		// defer CSVFile.Close()
		// csvWriter := csv.NewWriter(CSVFile)
		// cErr := csvWriter.WriteAll(Data.RC)
		// if cErr != nil {
		// 	log.Printf("Cannot write CSV records %v", cErr)
		// }

		// *************************************************************
		http.Redirect(w, req, "/saveCSVFile", http.StatusSeeOther)
		return
	}

}

// SaveCSVFile GET ROUTE Saves the CSV file.
func (c Controller) SaveCSVFile(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	fmt.Printf("File name %s\n", ExportName)
	// serve file
	header := w.Header()
	header.Set("Content-Disposition", "attachment; filename="+FullName)
	header.Set("Content-Type", header.Get("Content-Type"))
	csvWriter := csv.NewWriter(w)
	cErr := csvWriter.WriteAll(Data.RC)
	if cErr != nil {
		log.Printf("Cannot write CSV records %v", cErr)
	}
	xPortNo := len(Data.RC)
	fmt.Println("Total rows exported ", xPortNo)
	return

}

// About provides a method for the information page
func (c Controller) About(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	// populate the template struct with empty values

	templateData := struct {
		FileName string
	}{
		FileName: Name,
	}
	c.tpl.ExecuteTemplate(w, "about.gohtml", templateData)
}

// Csv functions
// *******************************************
// readCSV reads csv file and stores in struct
func readCSV(f string) (CSVData, error) {
	// Create value of CSVData
	d := [][]string{}

	// *******************************************
	// Code for reading file from disc

	// for reading of a disk file
	// f, fErr := os.Open(filename)
	// if fErr != nil {
	// 	log.Fatalf("Cannot open area csv file %v\n", fErr)
	// }
	// defer f.Close()

	// *******************************************

	r := csv.NewReader(strings.NewReader(f))
	d, csvErr := r.ReadAll()
	if csvErr != nil {
		log.Printf("Cannot read csv file %v\n", csvErr)
		return CSVData{}, fmt.Errorf("corrupt CSV file %v", csvErr)
	}
	fmt.Printf("Completed read from csv records read %d\n", len(d))
	//Populate struct values
	data := CSVData{}
	data.RC = make([][]string, len(d))
	for r, record := range d {
		data.RC[r] = append(data.RC[r], record...)
	}
	return data, nil
}

// buildInputSelection builds the column names
func buildInputSelection(rowLen int) []string {
	s := []string{}
	for i := 0; i < rowLen+1; i++ {
		tmp := "rc" + fmt.Sprint(i+1)
		s = append(s, tmp)
	}
	return s
}

// newRow creates new cells inside new row for use in adding a row to table
func newRow(colCount int) ([]template.JS, []jsRow) {
	s := []template.JS{}
	st := []jsRow{}
	var (
		tmp  string
		tmpR jsRow
	)
	for i := 0; i < colCount+1; i++ {
		tmp = "cell" + fmt.Sprint(i+1) + " = "
		if i == colCount {
			tmp += "newRow.insertCell(" + fmt.Sprint(i) + "),"
			s = append(s, template.JS(tmp))
			//Build cell struct
			tmpR.CellName = "rc" + fmt.Sprint(i+1)
			tmpR.CellNameJS = template.JS(tmpR.CellName)
			tmpR.Cell = template.JS("cell" + fmt.Sprint(i+1))
			tmpR.EndMarker = template.JS(";")
			st = append(st, tmpR)
			return s, st
		}
		tmp += "newRow.insertCell(" + fmt.Sprint(i) + "),"
		s = append(s, template.JS(tmp))
		//Build cell struct
		tmpR.CellName = "rc" + fmt.Sprint(i+1)
		tmpR.CellNameJS = template.JS(tmpR.CellName)
		tmpR.Cell = template.JS("cell" + fmt.Sprint(i+1))
		tmpR.EndMarker = template.JS(",")
		st = append(st, tmpR)
	}
	return s, st
}

// **************************************************************************************************
// JSON ERROR handling
func (mr *malformedRequest) Error() string {
	return mr.msg
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			return &malformedRequest{status: http.StatusUnsupportedMediaType, msg: msg}
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			return &malformedRequest{status: http.StatusRequestEntityTooLarge, msg: msg}

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		msg := "Request body must only contain a single JSON object"
		return &malformedRequest{status: http.StatusBadRequest, msg: msg}
	}

	return nil
}

//removeIndex removes a slice element
func removeIndex(s [][]string, index int) [][]string {
	temp := [][]string{}
	for i := 0; i < len(s); i++ {
		if i != index-1 {
			temp = append(temp, s[i])
		}
	}
	return temp
}

// Check if the file has been loaded.
func check() bool {
	if len(Data.RC) == 0 {
		return true
	}
	return false
}

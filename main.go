package main

import (
	"log"
	"net/http"

	"github.com/nexlight101/CSV_project/conf"
	"github.com/nexlight101/CSV_project/modules"

	"github.com/julienschmidt/httprouter"
)

func main() {
	// Get a new router.
	r := httprouter.New()

	// Get a template controller value.
	c := modules.NewController(conf.TPL)

	// Handle dictionary routes
	r.GET("/", c.Main)                      // Landing page
	r.GET("/mainV", c.IndexV)               // GET ROUTE displays the csv file
	r.GET("/main", c.Index)                 // GET ROUTE displays the csv form row
	r.POST("/findCSV", c.FindCSV)           // POST ROUTE reads the csv filename
	r.POST("/saveRecord", c.SaveRecord)     // POST ROUTE reads the csv filename
	r.GET("/saveCSVFile", c.SaveCSVFile)    // GET ROUTE saves the csv filename
	r.GET("/saveCSV", c.SaveCSV)            // GET ROUTE saves the csv filename
	r.GET("/about", c.About)                // GET ROUTE displays the about information
	r.POST("/addNewRecord", c.AddNewRecord) // POST ROUTE adds a row from AJAX request
	r.POST("/deleteRecord", c.DeleteRecord) // POST ROUTE remove a row from AJAX request
	r.POST("/editRecord", c.EditRecord)     // POST ROUTE edit a row from AJAX request
	r.POST("/addNewColumn", c.AddNewColumn) // POST ROUTE adding new column from AJAX request
	// Handle icon
	http.Handle("/favicon.ico", http.NotFoundHandler())
	// Serve CSS
	// r.ServeFiles("/public/stylesheets/*filepath", http.Dir("public/stylesheets"))

	// if not found look for a static file
	static := httprouter.New()
	static.ServeFiles("/public/*filepath", http.Dir("public"))
	// r.NotFound = http.FileServer(http.Dir("public/stylesheets"))
	r.NotFound = static
	// Server
	log.Fatal(http.ListenAndServe(":80", r))

}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/labstack/echo"
)

type GPSData struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp string  `json:"timestamp"`
}

var dataSource = readDataFromJsonFile("data.json")
var totalData = len(dataSource)

func main() {
	r := echo.New()

	// routes here
	r.POST("/data", handleRequest)

	var stopChan = make(chan os.Signal, 2)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := r.Start(":9000"); err != nil {
			log.Fatal("error start echo web, err: ", err.Error())
		}
	}()

	fmt.Println("server started at :9000")
	fmt.Println("recieve signal ", <-stopChan) // wait for SIGINT

	if err := r.Shutdown(ctx); err != nil {
		log.Fatal("error when shutdown", err.Error())
	}
}

func handleRequest(ctx echo.Context) error {

	request := map[string]int{}
	if err := ctx.Bind(&request); err != nil {
		fmt.Println("error when err: ", err.Error())
		return ctx.String(500, "error : "+err.Error())
	}

	start := request["start"]
	end := request["end"]

	if start == 0 {
		return ctx.String(http.StatusBadRequest, "'start' should be higher than 0")
	}

	if end == 0 {
		end = totalData
	}

	if start >= end {
		return ctx.String(http.StatusBadRequest, "'start' can't be higher or equal with 'end'")
	}

	if end > totalData {
		return ctx.String(http.StatusBadRequest, "'end' can't be higher that 'total_data'= "+strconv.Itoa(totalData))
	}

	data := make([]GPSData, end-start+1)
	k := 0
	for i := start - 1; i < end; i++ {
		data[k] = dataSource[i]
		k++
	}

	response := map[string]interface{}{
		"totalData":        totalData,
		"totalRetriveData": end - start + 1,
		"start":            start,
		"end":              end,
		"data":             data,
	}

	return ctx.JSON(200, response)
}

func readDataFromJsonFile(fileName string) []GPSData {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	// Now let's unmarshall the data into `payload`
	var payload []GPSData
	err = json.Unmarshal(content, &payload)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	return payload
}

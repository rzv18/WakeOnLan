package main

import (
	"log"
	"os"

	"github.com/gocarina/gocsv"
)

// LoadComputerList loads the Computer and returns the read list or an error
func loadComputerList(computerCsvFilePath string) ([]Computer, error) {
	var computers []Computer

	computerCsvFilePointer, computerCsvFileOpenError := os.Open(computerCsvFilePath)
	if computerCsvFileOpenError != nil {
		log.Fatalf("Error on Opening the file")
		return computers, computerCsvFileOpenError
	}

	if computerCsvFileParserError := gocsv.UnmarshalFile(computerCsvFilePointer, &computers); computerCsvFileParserError != nil {
		computerCsvFilePointer.Close()
		return computers, computerCsvFileParserError
	}

	computerCsvFilePointer.Close()
	return computers, nil
}

// SaveComputerList saves the computer list to a CSV file
func saveComputerList(computerCsvFilePath string, computers []Computer) error {
	file, err := os.OpenFile(computerCsvFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	return gocsv.MarshalFile(&computers, file)
}

// Test if Path is a File and it exists
func FileExists(name string) bool {
	if fi, err := os.Stat(name); err == nil {
		if fi.Mode().IsRegular() {
			return true
		}
	}
	return false
}

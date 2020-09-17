package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	pathLogFile    = "/var/log/emptty"
	pathLogFileOld = "/var/log/emptty.old"
)

// propertyFunc defines method to be invoked during readProperties method for each record.
type propertyFunc func(key string, value string)

// readProperties reads defined filePath per line and parses each key-value pair.
// These pairs are used as parameters for invoking propertyFunc
func readProperties(filePath string, method propertyFunc) error {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return errors.New("Could not open file " + filePath)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "#") && strings.Index(line, "=") >= 0 {
			splitIndex := strings.Index(line, "=")
			key := strings.ReplaceAll(line[:splitIndex], "export ", "")
			value := line[splitIndex+1:]
			if strings.Index(value, "#") >= 0 {
				value = value[:strings.Index(value, "#")]
			}
			key = strings.ToUpper(strings.TrimSpace(key))
			value = strings.TrimSpace(value)
			method(key, value)
		}
	}
	return scanner.Err()
}

// Checks, if file on path exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Handles error passed as string and calls handleErr function.
func handleStrErr(err string) {
	if err != "" {
		handleErr(errors.New(err))
	}
}

// If error is not nil, otherwise it prints error, waits for user input and then exits the program.
func handleErr(err error) {
	if err != nil {
		log.Print(err)
		fmt.Printf("Error: %s\n", err)
		fmt.Printf("\nPress Enter to continue...")
		bufio.NewReader(os.Stdin).ReadString('\n')
		os.Exit(1)
	}
}

// Handles passed arguments.
func handleArgs() {
	for _, arg := range os.Args {
		switch arg {
		case "-v", "--version":
			fmt.Printf("emptty %s\nhttps://github.com/tvrzna/emptty\n\nReleased under the MIT License.\n\n", version)
			os.Exit(0)
		case "-d", "--daemon":
			startDaemon()
			os.Exit(0)
		}
	}
}

// Initialize logger to file defined by pathLogFile.
func initLogger() {
	if fileExists(pathLogFile) {
		os.Remove(pathLogFileOld)
		os.Rename(pathLogFile, pathLogFileOld)
	}

	f, err := os.OpenFile(pathLogFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err == nil {
		log.SetOutput(f)
	}
}

// Sanitize value.
func sanitizeValue(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}

// Makes directories up to last part of path (to make sure to not make dir, that is named as result file)
func mkDirsForFile(path string, perm os.FileMode, usr *sysuser) error {
	if !fileExists(path) && path != "" {
		arrPath := strings.Split(path, "/")
		for i, _ := range arrPath {
			if i < len(arrPath)-1 {
				newPath := strings.Join(arrPath[:i+1], "/")
				if newPath != "" && !fileExists(newPath) {
					err := os.Mkdir(newPath, perm)
					if err == nil {
						err = os.Chown(newPath, usr.uid, usr.gid)
					}
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

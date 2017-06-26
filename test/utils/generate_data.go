package main

import (
	"fmt"
	"os"
	"bufio"
	"math/rand"
	"strconv"
)


func random(min, max int) int {
   return rand.Intn(max - min) + min
}


// generateData generates random data to be used by scheduler.
func generateData(podCount int) error {
	// Create a file.
	file, err := os.Create("sample.txt")
	if err != nil {
		return fmt.Errorf("File creation failed.")
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	for i:=0; i< podCount ; i ++ {
		fmt.Fprintln(writer, "CPU " + strconv.Itoa(random(10,100)) +"m"  +"," + "Memory " + strconv.Itoa(random(50,500))+ "Mi")
	}
	writer.Flush()
	return nil
}


func main() {
	var podCount int
	fmt.Scanf("%d", &podCount)
	generateData(podCount)
}
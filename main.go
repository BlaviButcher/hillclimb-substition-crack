package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

var maxKey = strings.Split("ABCDEFGHIJKLMNOPQRSTUVWXYZ", "")
var maxScore = -10000000000.0
var parentKey = maxKey
var parentScore = maxScore
var iteration = 0
var maxDeciphered string
var parentMaxDeciphered string

// the order in which each letter (number duo) appears in the cipher
var cipherOrder []string

func main() {

	// clear old best file
	if err := os.Truncate("best.txt", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	// read in cipher and create array - cipher file is made of of numbers ie. 11 12 15 21 22 23 24 25
	cipher, err := readCipherFile()
	if err != nil {
		panic(err)
	}

	cipherOrder = removeDuplicateStr(cipher)

	// create a map of cipher cipherLetter: keyLetter
	// blank to start
	mappedCipher := mapCipher(cipher)

	runCipher(cipher, mappedCipher)

}

func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func runCipher(cipher []string, cipherKeyMap map[string]string) {
	quadgrams := readGram("quadgrams.txt")
	trigrams := readGram("trigrams.txt")
	bigrams := readGram("bigrams.txt")
	monograms := readGram("monograms.txt")
	quintgrams := readGram("quintgrams.txt")

	for true {
		iteration++
		fmt.Println("Iteration:", iteration)

		// shuffle parent key around
		rand.Shuffle(len(parentKey), func(i, j int) {
			parentKey[i], parentKey[j] = parentKey[j], parentKey[i]
		})

		// add key to cipherMap
		cipherKeyMap = mapKeyToCipher(parentKey, cipherKeyMap)
		deciphered := decipher(cipher, cipherKeyMap)
		parentScore := scoreDecipher(deciphered, quadgrams, trigrams, bigrams, monograms, quintgrams)
		parentMaxDeciphered = "null"

		count := 0
		for count < 1000 {
			a := rand.Intn(26)
			b := rand.Intn(26)
			child := parentKey

			// switch two characters in child
			child[a], child[b] = child[b], child[a]

			cipherKeyMap = mapKeyToCipher(child, cipherKeyMap)

			deciphered := decipher(cipher, cipherKeyMap)

			score := scoreDecipher(deciphered, quadgrams, trigrams, bigrams, monograms, quintgrams)

			if score > parentScore {
				parentScore = score
				parentKey = child
				parentMaxDeciphered = strings.Join(deciphered, "")
				count = 0
			}
			count++
		}

		if parentScore > maxScore {
			maxScore = parentScore
			maxKey = parentKey
			maxDeciphered = parentMaxDeciphered
			out := fmt.Sprintf("New max score: %d on iteration: %d ", maxScore, iteration)
			fmt.Printf(out)
			fmt.Printf("Key: %s\n", maxKey)
			writeToFile(out, maxKey, cipherKeyMap, maxDeciphered)

		}

	}
}
func writeToFile(data string, key []string, keyMap map[string]string, maxDeciphered string) {

	file, err := os.OpenFile("best.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	datawriter := bufio.NewWriter(file)

	_, _ = datawriter.WriteString(data + "\n" + strings.Join(key, "") + "\n" + keyMapToString(keyMap) + "\n" + maxDeciphered + "\n\n")

	datawriter.Flush()
	file.Close()
}

func getDecipherText(keyMap map[string]string, cipher []string) string {
	var deciphered string
	for _, letter := range cipher {
		deciphered += keyMap[letter]
	}
	return deciphered

}

func readGram(filename string) map[string]int {
	grams := make(map[string]int)
	// Open the file
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	for _, line := range strings.Split(string(file), "\n") {
		pair := strings.Split(line, " ")
		score, _ := strconv.Atoi(pair[1])
		grams[pair[0]] = score
	}

	return grams

}

func keyMapToString(keyMap map[string]string) string {
	var keyString string
	for key, letter := range keyMap {
		keyString += fmt.Sprintf("%s:%s ", key, letter)
	}
	return keyString
}

func scoreDecipher(deciphered []string, quadgrams map[string]int, trigrams map[string]int, bigrams map[string]int, monograms map[string]int, quintgrams map[string]int) float64 {
	// read in quadgram file and create map of quadgram: score

	score := 0.0
	// quadScore := 0
	// for i := 3; i < len(deciphered); i++ {
	// 	quadgram := deciphered[i-3] + deciphered[i-2] + deciphered[i-1] + deciphered[i]
	// 	if s, ok := quadgrams[quadgram]; ok {
	// 		quadScore += s
	// 	}
	// }
	// score += quadScore

	triScore := 0
	for i := 2; i < len(deciphered); i++ {
		trigram := deciphered[i-2] + deciphered[i-1] + deciphered[i]
		if s, ok := trigrams[trigram]; ok {
			triScore += s
		}
	}
	score += triScore

	biScore := 0
	for i := 1; i < len(deciphered); i++ {
		bigram := deciphered[i-1] + deciphered[i]
		if s, ok := bigrams[bigram]; ok {
			biScore += s
		}
	}
	score += biScore

	// monoScore := 0
	// for i := 0; i < len(deciphered); i++ {
	// 	monogram := deciphered[i]
	// 	if s, ok := monograms[monogram]; ok {
	// 		monoScore += s
	// 	}
	// }
	// score += monoScore

	// quintScore := 0
	// for i := 4; i < len(deciphered); i++ {
	// 	quintgram := deciphered[i-4] + deciphered[i-3] + deciphered[i-2] + deciphered[i-1] + deciphered[i]
	// 	if s, ok := quintgrams[quintgram]; ok {
	// 		quintScore += s
	// 	}
	// }
	// score += quintScore

	return score
}

func decipher(cipher []string, cipherKeyMap map[string]string) []string {
	decipheredArray := make([]string, len(cipher))
	for i, letter := range cipher {
		decipheredArray[i] = cipherKeyMap[letter]
	}

	return decipheredArray

}

// align cipher to key
func mapKeyToCipher(key []string, cipherMap map[string]string) map[string]string {

	for i := 0; i < len(key); i++ {
		cipherMap[cipherOrder[i]] = key[i]
	}
	return cipherMap
}

func readCipherFile() ([]string, error) {
	// Open the file
	file, err := ioutil.ReadFile("cipher.txt")
	if err != nil {
		return nil, fmt.Errorf("opening cipher: %v", err)
	}
	cipher := make([]string, 0)
	for _, letter := range strings.Split(string(file), " ") {
		cipher = append(cipher, letter)
	}

	return cipher, nil

}

func mapCipher(cipher []string) map[string]string {
	cipherMap := make(map[string]string)
	for _, letter := range cipher {
		cipherMap[letter] = ""
	}

	// if missing a letter, add a dummy to the map
	// it will act as our missing letter
	if len(cipherMap) < 26 {
		diff := 26 - len(cipherMap)
		for i := 0; i < diff; i++ {
			cipherMap[strconv.Itoa(i+1000)] = ""
			cipherOrder = append(cipherOrder, strconv.Itoa(i+1000))
		}
	}

	return cipherMap
}

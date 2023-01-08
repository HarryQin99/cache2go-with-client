package main

import (
	"bufio"
	cache2gowithclient "cache2go-with-client/cache"
	"fmt"
	"os"
	"strings"
	"time"
)

/*
CMD:
init table
*/
type cmd string

var (
	cacheTable *cache2gowithclient.CacheTable
)

const (
	DefaulfCacheTableName string = "default"
	Set                   cmd    = "set"
	Get                   cmd    = "get"
	Check                 cmd    = "check"
)

// Initialize a cachetable while the program starts
func Init() {
	cacheTable = cache2gowithclient.Cache(DefaulfCacheTableName)
}

func main() {

	// Init default cache table
	Init()
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Cache2go")
	fmt.Println("---------------------")

	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		args := strings.Split(text, " ")

		if strings.Compare("hi", text) == 0 {
			fmt.Println("hello, Yourself")
		} else if Set.Compare(args[0]) {
			err := Add(args)
			if err != nil {
				fmt.Println(err.Error())
			}
		} else if Get.Compare(args[0]) {
			cacheItem, err := Fetch(args)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println(cacheItem.Value())
			}
		} else if Check.Compare(args[0]) {
			exists, err := CheckExist(args)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println(exists)
			}
		}

	}
}

func (c cmd) Compare(s string) bool {
	return strings.Compare(string(c), s) == 0
}

func Add(args []string) error {
	if len(args) < 3 {
		return cache2gowithclient.ErrSetValFailed
	}

	key := args[1]
	value := args[2]
	lifeSpan, err := time.ParseDuration(args[3])
	if err != nil {
		return cache2gowithclient.ErrSetValFailed
	}
	cacheTable.Add(key, lifeSpan, value)
	return nil
}

func Fetch(args []string) (*cache2gowithclient.CacheItem, error) {
	if len(args) < 2 {
		return nil, cache2gowithclient.ErrGetValFailed
	}
	key := args[1]
	return cacheTable.Value(key)
}

func CheckExist(args []string) (bool, error) {
	if len(args) < 2 {
		return false, cache2gowithclient.ErrGetValFailed
	}
	key := args[1]
	return cacheTable.Exists(key), nil
}

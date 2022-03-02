package main

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const retweets_api = "https://api.twitter.com/1.1/statuses/retweeters/ids.json"
const users_data = "https://api.twitter.com/2/users/by/username/"
const user_data = "https://api.twitter.com/2/users/"
const followers_api = "https://api.twitter.com/1.1/followers/ids.json"
const api_v2 = "https://api.twitter.com/2"
const auth = "Bearer AAAAAAAAAAAAAAAAAAAAALJXXwEAAAAAsli9svNXW5Sm%2BiWIYYprPeEIwt0%3DJM4V7mrvdNEH8ksJYd9gzQdkehOkBdN7JaipWyYNMuyV6KMe32"

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/", hello)
	e.GET("/:user_id", get_retweet_users())
	e.Logger.Fatal(e.Start(":3000")) // コンテナ側の開放ポートと一緒にすること
}

type retweetData struct {
	Ids               []string `json:"ids"`
	NextCursor        int      `json:"next_cursor"`
	NextCursorStr     string   `json:"next_cursor_str"`
	PreviousCursor    int      `json:"previous_cursor"`
	PreviousCursorStr string   `json:"previous_cursor_str"`
}
type userData struct {
	Data struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
	} `json:"data"`
}

func get_retweet_users() echo.HandlerFunc {
	return func(c echo.Context) error {
		var myuserId = get_userId("DICE_prgmsc")
		var followers = get_followers(myuserId)
		var now = time.Now().Format("2006-01-02_15_04_05")
		followers_file, err1 := os.OpenFile(now+"_followers.txt", os.O_WRONLY|os.O_CREATE, 0666)
		if err1 != nil {
			//エラー処理
			log.Fatal(err1)
		}
		defer followers_file.Close()
		Include_followers_file, err2 := os.OpenFile(now+"_Include_followers.txt", os.O_WRONLY|os.O_CREATE, 0666)
		if err2 != nil {
			//エラー処理
			log.Fatal(err2)
		}
		defer Include_followers_file.Close()
		for _, val := range followers {
			fmt.Println(val)
		}
		var retweetUserFollowerList []string
		var retweetUserList []string
		var result []string
		for {
			userId := c.Param("user_id")
			req, _ := http.NewRequest("GET", fmt.Sprintf(retweets_api+"?id=%s&stringify_ids=true", userId), nil)
			req.Header.Set("Authorization", auth)
			client := new(http.Client)
			resp, _ := client.Do(req)
			byteArray, _ := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			//fmt.Println(string(byteArray))
			var data retweetData
			if err := json.NewDecoder(strings.NewReader(string(byteArray))).Decode(&data); err != nil {
				fmt.Println(err)
			}
			//for _, v := range data.Ids {
			//	fmt.Printf("%s \n", v)
			//}
			result = append(result, data.Ids...)
			if data.NextCursor == 0 {
				break
			}
		}
		for _, val := range result {
			if val != myuserId {
				users := get_user_data(val)
				//fmt.Println(users)
				retweetUserList = append(retweetUserList, users)
				fmt.Fprintln(Include_followers_file, users)
				if contains(followers, val) {
					retweetUserFollowerList = append(retweetUserFollowerList, users)
					fmt.Fprintln(followers_file, users)
				}
				time.Sleep(time.Millisecond * 10)
			}
		}
		//for _, val := range retweetUserList {
		//	fmt.Println(val)
		//}
		//fmt.Println(len(retweetUserList))
		return c.JSON(http.StatusOK, retweetUserList)
	}
}

// ユーザー情報ID,@xxx等取得
func get_userId(username string) string {
	api_path := users_data + username
	req, _ := http.NewRequest("GET", api_path, nil)
	req.Header.Set("Authorization", auth)
	client := new(http.Client)
	resp, _ := client.Do(req)
	byteArray, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	var data userData
	if err := json.NewDecoder(strings.NewReader(string(byteArray))).Decode(&data); err != nil {
		fmt.Println(err)
	}
	return data.Data.ID
}

// フォロワー一覧取得
func get_followers(userId string) []string {
	var result []string
	var url string
	var stringify_ids int = -1
	for {
		if stringify_ids != -1 {
			url = fmt.Sprintf(followers_api+"?user_id=%s&stringify_ids=true&cursor=%d", userId, stringify_ids)
		} else {
			url = fmt.Sprintf(followers_api+"?user_id=%s&stringify_ids=true", userId)
		}
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", auth)
		client := new(http.Client)
		resp, _ := client.Do(req)
		byteArray, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		fmt.Println(string(byteArray))
		var data retweetData
		if err := json.NewDecoder(strings.NewReader(string(byteArray))).Decode(&data); err != nil {
			fmt.Println(err)
		}
		result = append(result, data.Ids...)
		stringify_ids = data.NextCursor
		if data.NextCursor == 0 {
			break
		}
		time.Sleep(time.Millisecond * 10)
	}
	return result
}
func get_user_data(id string) string {
	api_path := user_data + id
	req, _ := http.NewRequest("GET", api_path, nil)
	req.Header.Set("Authorization", auth)
	client := new(http.Client)
	resp, _ := client.Do(req)
	byteArray, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	var data userData
	//if err := json.NewDecoder(strings.NewReader(string(byteArray))).Decode(&data); err != nil {
	if err := json.Unmarshal(byteArray, &data); err != nil {
		fmt.Println(err)
	}
	result := fmt.Sprintf("%s：@%s", data.Data.Name, data.Data.Username)
	return result
}

func contains(s []string, e string) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
}

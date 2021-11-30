package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

//var logger *zap.Logger

//точка входа
func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	logger.Info("start")


	botToken := "2102830230:AAH4ENaEOafawyYRxICfy5-ZPfWrqhHtMEg"
	//https://api.telegram.org/bot<token>/METHOD_NAME
	botApi := "https://api.telegram.org/bot"
	botUrl := botApi+botToken
	offset := 0

	for {
		updates, err := getUpdates(botUrl, offset)
		if err != nil {
			msg := fmt.Sprintf("func getUpdates was failed", err)
			logger.Error(msg)
		}
		for _, update := range updates {
			err = respond(botUrl, update)
			offset = update.UpdateId + 1
			logger.Info("new update")
			//msg := fmt.Sprintf("#: ", offset)
			//logger.Info(msg)
			if err != nil {
				msg := fmt.Sprintf("loop is not working", err)
				logger.Error(msg)
			}
		}

		fmt.Println(updates)
	}


}

//запрос обновлений
func getUpdates(botUrl string, offset int) ([]Update, error){
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	resp, err := http.Get(botUrl+ "/getUpdates" + "?offset=" + strconv.Itoa(offset))
	if err != nil {
		return nil, err
	}
	//defer  log.Fatal(resp.Body.Close())
	defer  resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("#70 body error" , err)
		logger.Error(msg)
		return nil, err
	}
	var restResponse RestResponse

	err = json.Unmarshal(body, &restResponse)
	if err != nil {
		msg := fmt.Sprintf("#79 json.Unmarshal", err)
		logger.Error(msg)
		return nil, err
	}

	return restResponse.Result, nil

}

//ответ на обновления
func respond(botUrl string, update Update) error {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	var botMessage BotMessage
	botMessage.ChatId = update.Message.Chat.ChatId
	botMessage.Text = update.Message.Text
	buf, err := json.Marshal(botMessage)
	if err != nil {
		msg := fmt.Sprintf("#100 buff error", err)
		logger.Error(msg)
		return err
	}
	_, err = http.Post(botUrl + "/sendMessage", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		msg := fmt.Sprintf("#106 http.Post error", err)
		logger.Error(msg)
		return err
	}


	return nil
}
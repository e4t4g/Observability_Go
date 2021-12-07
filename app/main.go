package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

//точка входа
func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	appLogger := logger.Sugar().Named("botLogg")

	appLogger.Info("start echo bot")

	botToken := "2102830230:AAH4ENaEOafawyYRxICfy5-ZPfWrqhHtMEg"
	//https://api.telegram.org/bot<token>/METHOD_NAME
	botApi := "https://api.telegram.org/bot"
	botUrl := botApi + botToken
	offset := 0

	for {
		updates, err := getUpdates(botUrl, appLogger.With("getUpdates"), offset)
		if err != nil {
			appLogger.Fatal("url is not set")
		}
		for _, update := range updates {
			err = respond(botUrl, appLogger.With("respond"), update)
			offset = update.UpdateId + 1
			appLogger.Info("got a new request")
			if err != nil {
				appLogger.Errorw("update error", "err", err)
			}
		}
		fmt.Println(updates)
	}
}

//запрос обновлений
func getUpdates(botUrl string, appLogger *zap.SugaredLogger, offset int) ([]Update, error) {

	resp, err := http.Get(botUrl + "/getUpdates" + "?offset=" + strconv.Itoa(offset))
	if err != nil {
		return nil, err
	}
	//defer resp.Body.Close()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("body error", err)
		appLogger.Error(msg)
		return nil, err
	}
	var restResponse RestResponse

	err = json.Unmarshal(body, &restResponse)
	if err != nil {
		msg := fmt.Sprintf("json.Unmarshal", err)
		appLogger.Error(msg)
		return nil, err
	}

	return restResponse.Result, nil

}

//ответ на обновления
func respond(botUrl string, appLogger *zap.SugaredLogger, update Update) error {

	var botMessage BotMessage
	botMessage.ChatId = update.Message.Chat.ChatId
	botMessage.Text = update.Message.Text
	buf, err := json.Marshal(botMessage)
	if err != nil {
		msg := fmt.Sprintf("buff error", err)
		appLogger.Error(msg)
		return err
	}
	_, err = http.Post(botUrl+"/sendMessage", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		msg := fmt.Sprintf("http.Post error", err)
		appLogger.Error(msg)
		return err
	}
	return nil
}


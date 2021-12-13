package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/spf13/pflag"
)

type bot struct {
	logger *zap.SugaredLogger
	cache  *cache.Cache
}

//точка входа
func main() {

	var configParse string

	pflag.StringVar(&configParse, "config", ".config.yml", "config file path")

	pflag.Parse()

	pwd, _ := os.Getwd()
	config, err := InitConfig(pwd + configParse)
	if err != nil {
		panic(err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	b := bot{
		logger: logger.Sugar(),
	}

	//https://api.telegram.org/bot<token>/METHOD_NAME

	botUrl := config.BotApi + config.BotToken
	fmt.Println(botUrl)
	offset := 0

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	b.InitCache()

	for {
		ctx, updates, err := b.getUpdates(ctx, botUrl, offset)
		if err != nil {
			b.logger.Fatal("url is not set")
		}
		for _, update := range updates {
			err := b.cache.Get(ctx, botUrl, &update)
			key := fmt.Sprintf("user_articles:%s", update)
			switch err {
			case cache.ErrCacheMiss:
				err = b.respond(ctx, botUrl, update)
				offset = update.UpdateId + 1
				b.logger.Info("got a new request")
				if err != nil {
					b.logger.Errorw("update error", "err", err)
				}
				fmt.Println(updates)
				err = b.cache.Set(&cache.Item{
					Ctx:   ctx,
					Key:   key,
					Value: update,
					TTL:   time.Hour,
				})
				if err != nil {
				}

			case nil:
				fmt.Println(updates)
			}

		}

	}
}

//запрос обновлений
func (b *bot) getUpdates(ctx context.Context, botUrl string, offset int) (context.Context, []Update, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "getUpdates")
	defer span.Finish()

	resp, err := http.Get(botUrl + "/getUpdates" + "?offset=" + strconv.Itoa(offset))
	if err != nil {
		b.logger.Errorw("error gert updates", err)
		span.LogFields(
			log.Error(err),
		)
		return nil, nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("body error", err)
		b.logger.Error(msg)
		span.LogFields(
			log.Error(err),
		)
		return nil, nil, err
	}
	var restResponse RestResponse

	err = json.Unmarshal(body, &restResponse)
	if err != nil {
		msg := fmt.Sprintf("json.Unmarshal", err)
		b.logger.Error(msg)
		span.LogFields(
			log.Error(err),
		)
		return nil, nil, err
	}

	return ctx, restResponse.Result, nil

}

//ответ на обновления
func (b *bot) respond(ctx context.Context, botUrl string, update Update) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "get respond")
	defer span.Finish()

	var botMessage BotMessage
	botMessage.ChatId = update.Message.Chat.ChatId
	botMessage.Text = update.Message.Text
	buf, err := json.Marshal(botMessage)
	if err != nil {
		msg := fmt.Sprintf("buff error", err)
		b.logger.Error(msg)
		return err
	}

	_, err = http.Post(botUrl+"/sendMessage", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		msg := fmt.Sprintf("http.Post error", err)
		b.logger.Error(msg)
		return err
	}
	return nil
}

func (b *bot) InitCache() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	b.cache = cache.New(&cache.Options{
		Redis:      rdb,
		LocalCache: cache.NewTinyLFU(24, time.Hour),
	})
}

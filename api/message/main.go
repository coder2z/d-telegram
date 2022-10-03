package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/coder2z/d-telegram/config"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
	"time"
)

var codeAuthenticatorFunc auth.CodeAuthenticatorFunc = func(
	ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter code: ")
	code, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(code), nil
}

func main() {
	// 初始化配置
	log, _ := zap.NewDevelopment(zap.IncreaseLevel(zapcore.InfoLevel), zap.AddStacktrace(zapcore.FatalLevel))
	defer func() { _ = log.Sync() }()

	init := config.Init()

	client := telegram.NewClient(init.AppID, init.AppHash, telegram.Options{
		Logger: log,
		SessionStorage: &session.FileStorage{
			Path: init.SessionFile,
		},
	})

	if err := client.Run(context.Background(), func(ctx context.Context) error {
		flow := auth.NewFlow(
			auth.CodeOnly("+15680544812", codeAuthenticatorFunc),
			auth.SendCodeOptions{},
		)
		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
			panic(err)
		}
		api := client.API()
		var data, err = api.ChannelsGetChannels(ctx, []tg.InputChannelClass{
			&tg.InputChannel{
				ChannelID: 1778419548,
			},
		})
		if err != nil {
			return err
		}
		g := data.GetChats()[0]
		gType := g.(*tg.Channel)
		//GetHistory
		data1, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
			Peer: &tg.InputPeerChannel{
				ChannelID:  1778419548,
				AccessHash: gType.AccessHash,
			},
			OffsetID:   0,
			OffsetDate: int(time.Now().Unix()),
			AddOffset:  0,
			Limit:      10,
			MaxID:      0,
			MinID:      0,
			Hash:       0,
		})
		if err != nil {
			panic(err)
		}
		messages := data1.(*tg.MessagesChannelMessages)

		fmt.Println(messages.Messages)

		// Return to close client connection and free up resources.
		return nil
	}); err != nil {
		panic(err)
	}
	// Client is closed.
}

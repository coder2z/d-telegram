package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/coder2z/d-telegram/config"
	"github.com/coder2z/d-telegram/download"
	"github.com/coder2z/d-telegram/xlog"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram/auth"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"strings"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/updates"
	updhook "github.com/gotd/td/telegram/updates/hook"
	"github.com/gotd/td/tg"
)

func InWatchChannelIDList(data []int64, c int64) bool {
	for _, v := range data {
		if v == c {
			return true
		}
	}
	return false
}

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
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx); err != nil {
		panic(err)
	}
}

func run(ctx context.Context) error {
	init := config.Get()
	d := tg.NewUpdateDispatcher()
	gaps := updates.New(updates.Config{
		Handler: d,
		Logger:  xlog.GetLogger(),
	})
	client := telegram.NewClient(init.AppID, init.AppHash, telegram.Options{
		UpdateHandler: gaps,
		Logger:        xlog.GetLogger(),
		SessionStorage: &session.FileStorage{
			Path: init.SessionFile,
		},
		Middlewares: []telegram.Middleware{
			updhook.UpdateHook(gaps.Handle),
		},
	})

	// 启动下载携程池
	download.Init(client.API(), init)

	// Setup message update handlers.
	d.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewChannelMessage) error {
		message, ok := update.GetMessage().(*tg.Message)
		if !ok {
			return nil
		}
		channel, ok := message.GetPeerID().(*tg.PeerChannel)
		if !ok {
			return nil
		}
		if !InWatchChannelIDList(init.WatchChannelIDList, channel.GetChannelID()) {
			return nil
		}

		xlog.GetLogger().Info("Channel message", zap.Any("message", message))

		// 判断是不是视频消息
		mediaDocument, ok := message.Media.(*tg.MessageMediaDocument)
		if !ok {
			return nil
		}
		document, ok := mediaDocument.Document.(*tg.Document)
		if !ok {
			return nil
		}
		if document.MimeType != "video/mp4" {
			return nil
		}

		// 判断大小 小文件就不走这里下载了
		if document.Size < init.MaxFileSize {
			return nil
		}

		// 判断是否有这些关键字的文件才下载
		var isd = true
		if len(init.WatchFileKeyWord) > 0 {
			isd = false
			for _, s := range init.WatchFileKeyWord {
				if strings.Contains(message.Message, s) {
					isd = true
					break
				}
			}
		}
		if !isd {
			return nil
		}
		download.Push(message)

		xlog.GetLogger().Info("Channel message", zap.Reflect("document", document))

		return nil
	})

	return client.Run(ctx, func(ctx context.Context) error {

		flow := auth.NewFlow(
			auth.CodeOnly(init.Phone, codeAuthenticatorFunc),
			auth.SendCodeOptions{},
		)
		err := client.Auth().IfNecessary(ctx, flow)
		if err != nil {
			return err
		}

		user, err := client.Self(ctx)
		if err != nil {
			return err
		}

		xlog.GetLogger().Info("Self user info", zap.Any("user", user))

		if err := gaps.Auth(ctx, client.API(), user.ID, user.Bot, true); err != nil {
			return err
		}
		defer func() { _ = gaps.Logout() }()

		<-ctx.Done()
		return ctx.Err()
	})
}

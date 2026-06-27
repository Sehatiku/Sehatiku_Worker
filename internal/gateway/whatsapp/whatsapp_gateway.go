package whatsapp

import (
	"context"
	"fmt"

	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	whatsmeow "go.mau.fi/whatsmeow"
)

type Sender interface {
	SendPatientReminder(ctx context.Context, phone, patientName, runType string) error
	SendCompanionReminder(ctx context.Context, phone, companionName, patientName string) error
}

type WhatsAppGateway struct {
	Client *whatsmeow.Client
	Log    *zap.Logger
}

func New(client *whatsmeow.Client, log *zap.Logger) *WhatsAppGateway {
	return &WhatsAppGateway{Client: client, Log: log}
}

func (g *WhatsAppGateway) SendPatientReminder(ctx context.Context, phone, patientName, runType string) error {
	if !g.Client.IsConnected() {
		return fmt.Errorf("whatsapp client not connected")
	}
	return g.send(ctx, phone, patientMessage(patientName, runType))
}

func (g *WhatsAppGateway) SendCompanionReminder(ctx context.Context, phone, companionName, patientName string) error {
	if !g.Client.IsConnected() {
		return fmt.Errorf("whatsapp client not connected")
	}
	return g.send(ctx, phone, companionMessage(companionName, patientName))
}

func (g *WhatsAppGateway) send(ctx context.Context, phone, text string) error {
	jid := types.NewJID(phone, types.DefaultUserServer)
	_, err := g.Client.SendMessage(ctx, jid, &waE2E.Message{
		Conversation: proto.String(text),
	})
	if err != nil {
		g.Log.Warn("whatsapp send failed", zap.String("phone", phone), zap.Error(err))
	}
	return err
}

func patientMessage(name, runType string) string {
	if runType == "noon" {
		return fmt.Sprintf(
			"Halo %s, jangan lupa catat data kesehatanmu hari ini ya! "+
				"Ketik data glucosa, tekanan darah, atau aktivitasmu lewat WhatsApp ini.",
			name,
		)
	}
	return fmt.Sprintf(
		"Halo %s, hari ini belum ada catatan kesehatanmu. "+
			"Masih sempat kok, catat sekarang supaya nakesmu bisa pantau kondisimu.",
		name,
	)
}

func companionMessage(companionName, patientName string) string {
	return fmt.Sprintf(
		"Halo %s, mohon bantu ingatkan %s untuk mencatat data kesehatannya hari ini ya.",
		companionName, patientName,
	)
}

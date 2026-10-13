package main

import (
	"database/sql"
	"fmt"

	waProto "go.mau.fi/whatsmeow/binary/proto"
)

// Old rows remain NULL: their original PTT flag was never recorded.
func ensureVoiceNoteSchema(db *sql.DB) error {
	rows, err := db.Query("PRAGMA table_info(messages)")
	if err != nil {
		return err
	}
	found := false
	for rows.Next() {
		var cid, notNull, pk int
		var name, kind string
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &kind, &notNull, &defaultValue, &pk); err != nil {
			rows.Close()
			return err
		}
		if name == "is_voice_note" {
			found = true
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	if !found {
		if _, err := db.Exec("ALTER TABLE messages ADD COLUMN is_voice_note BOOLEAN"); err != nil {
			return fmt.Errorf("voice note migration: %w", err)
		}
	}
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS messages_voice_inbox ON messages(timestamp) WHERE media_type = 'audio' AND is_voice_note = 1 AND is_from_me = 0")
	return err
}

func voiceNoteFlag(msg *waProto.Message) *bool {
	if msg == nil || msg.GetAudioMessage() == nil {
		return nil
	}
	flag := msg.GetAudioMessage().GetPTT()
	return &flag
}

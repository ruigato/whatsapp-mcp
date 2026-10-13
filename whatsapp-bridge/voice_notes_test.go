package main

import (
	"database/sql"
	"testing"
	"time"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

func TestVoiceNoteFlag(t *testing.T) {
	if voiceNoteFlag(nil) != nil || voiceNoteFlag(&waProto.Message{Conversation: proto.String("text")}) != nil {
		t.Fatal("non-audio must have unknown voice flag")
	}
	for _, want := range []bool{true, false} {
		msg := &waProto.Message{AudioMessage: &waProto.AudioMessage{PTT: proto.Bool(want)}}
		if got := voiceNoteFlag(msg); got == nil || *got != want {
			t.Fatalf("PTT = %v, want %v", got, want)
		}
	}
}

func TestVoiceMigrationPreservesOldRowsAndStoresPTT(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`CREATE TABLE messages(id TEXT, chat_jid TEXT, sender TEXT, content TEXT,
		timestamp TIMESTAMP, is_from_me BOOLEAN, media_type TEXT, filename TEXT, url TEXT,
		media_key BLOB, file_sha256 BLOB, file_enc_sha256 BLOB, file_length INTEGER,
		PRIMARY KEY(id,chat_jid)); INSERT INTO messages(id,chat_jid,media_type) VALUES('old','chat','audio')`)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := ensureVoiceNoteSchema(db); err != nil {
			t.Fatal(err)
		}
	}
	var old sql.NullBool
	if err := db.QueryRow("SELECT is_voice_note FROM messages WHERE id='old'").Scan(&old); err != nil || old.Valid {
		t.Fatalf("old row flag changed: %v / %v", old, err)
	}
	store := &MessageStore{db: db}
	flag := true
	if err := store.StoreMessage("new", "chat", "sender", "", time.Now(), false, "audio", "audio.ogg", "url", nil, nil, nil, 4, &flag); err != nil {
		t.Fatal(err)
	}
	var got bool
	if err := db.QueryRow("SELECT is_voice_note FROM messages WHERE id='new'").Scan(&got); err != nil || !got {
		t.Fatalf("new voice note not stored: %v / %v", got, err)
	}
}

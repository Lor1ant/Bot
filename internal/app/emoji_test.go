package app

import "testing"

func TestApplyPremiumEmojiOff(t *testing.T) {
	if got := applyPremiumEmoji("✅ готово", nil); got != "✅ готово" {
		t.Fatalf("без карты текст не должен меняться: %q", got)
	}
}

func TestApplyPremiumEmojiOn(t *testing.T) {
	m := map[string]string{"✅": "123"}
	got := applyPremiumEmoji("✅ готово", m)
	want := `<tg-emoji emoji-id="123">✅</tg-emoji> готово`
	if got != want {
		t.Fatalf("got=%q want=%q", got, want)
	}
}

func TestApplyPremiumEmojiEmptyID(t *testing.T) {
	m := map[string]string{"✅": ""}
	if got := applyPremiumEmoji("✅ ok", m); got != "✅ ok" {
		t.Fatalf("пустой id -> без изменений: %q", got)
	}
}

package service

import "testing"

func TestExtractReply(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain reply",
			input: "Thanks for the update!",
			want:  "Thanks for the update!",
		},
		{
			name:  "reply with quoted lines",
			input: "Sounds good.\n\n> On Mar 1 someone wrote:\n> original text",
			want:  "Sounds good.\n",
		},
		{
			name:  "reply with On...wrote: marker",
			input: "I agree.\n\nOn Monday, March 1 someone wrote:\n> blah",
			want:  "I agree.\n",
		},
		{
			name:  "reply with dashes separator",
			input: "Got it.\n\n---------- Forwarded message ----------\nold stuff",
			want:  "Got it.\n",
		},
		{
			name:  "reply with underscores separator",
			input: "Will do.\n\n______________________________\noriginal",
			want:  "Will do.\n",
		},
		{
			name:  "reply with Original Message",
			input: "OK.\n\n-----Original Message-----\nold",
			want:  "OK.\n",
		},
		{
			name:  "empty reply",
			input: "> quoted only",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractReply(tt.input)
			if got != tt.want {
				t.Errorf("extractReply() = %q, want %q", got, tt.want)
			}
		})
	}
}

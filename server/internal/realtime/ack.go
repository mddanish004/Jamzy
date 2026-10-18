package realtime

import "github.com/zishang520/socket.io/v2/socket"

type replyFunc func(payload any)

func noReply(any) {}

func replyTo(ack socket.Ack) replyFunc {
	if ack == nil {
		return noReply
	}
	return func(payload any) {
		ack([]any{payload}, nil)
	}
}

func splitArgs(args []any) (any, socket.Ack) {
	if len(args) == 0 {
		return nil, nil
	}

	last := args[len(args)-1]
	ack, ok := last.(socket.Ack)
	if !ok {
		return args[0], nil
	}

	if len(args) < 2 {
		return nil, ack
	}

	return args[0], ack
}

func field(payload any, key string) any {
	values, ok := payload.(map[string]any)
	if !ok {
		return nil
	}
	return values[key]
}

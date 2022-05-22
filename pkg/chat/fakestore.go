package chat

import "context"

type fakeStore struct{}

func (f fakeStore) GetAllChats(ctx context.Context, uuid string) ([]string, error) {
	return nil, nil
}

func (f fakeStore) SaveChat(ctx context.Context, uuid1, uuid2 string) error {
	return nil
}

func (f fakeStore) GetChat(ctx context.Context, uuid1, uuid2 string) error {
	return nil
}

func (f fakeStore) SaveMessage(ctx context.Context, m *Message) error {
	return nil
}

func (f fakeStore) LoadAllMessages(ctx context.Context, uuid1, uuid2 string) ([]*Message, error) {
	return nil, nil
}

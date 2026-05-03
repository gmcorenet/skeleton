package event

import "time"

type UserCreatedEvent struct {
	UserID    string
	Email     string
	CreatedAt time.Time
}

func NewUserCreatedEvent(userID, email string) *UserCreatedEvent {
	return &UserCreatedEvent{
		UserID:    userID,
		Email:     email,
		CreatedAt: time.Now(),
	}
}

func (e *UserCreatedEvent) GetName() string {
	return "user.created"
}

func (e *UserCreatedEvent) GetUserID() string {
	return e.UserID
}

func (e *UserCreatedEvent) GetEmail() string {
	return e.Email
}

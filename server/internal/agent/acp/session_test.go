package acp

import (
	"context"
	"testing"
	"time"

	acpsdk "github.com/coder/acp-go-sdk"
)

func TestListCommandsWaitsForAsyncAvailableCommands(t *testing.T) {
	const sessionKey = "test-session"
	state := &sessionState{ID: acpsdk.SessionId("agent-session")}
	proc := &Process{sessions: map[string]*sessionState{sessionKey: state}}
	sess := &session{proc: proc, sessionKey: sessionKey}

	go func() {
		time.Sleep(50 * time.Millisecond)
		state.setCommands([]acpsdk.AvailableCommand{{
			Name:        "resume",
			Description: "Resume a previous session",
		}})
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	commands, err := sess.ListCommands(ctx)
	if err != nil {
		t.Fatalf("ListCommands returned error: %v", err)
	}
	if len(commands.Commands) != 1 {
		t.Fatalf("expected delayed command to be returned, got %#v", commands.Commands)
	}
	if commands.Commands[0].Name != "resume" {
		t.Fatalf("expected resume command, got %#v", commands.Commands[0])
	}
}

func TestSessionUpdateStoresAvailableCommandsWithoutHandler(t *testing.T) {
	const agentSessionID = "agent-session"
	state := &sessionState{ID: acpsdk.SessionId(agentSessionID)}
	proc := &Process{
		sessions:     map[string]*sessionState{"test-session": state},
		sessionsByID: map[string]*sessionState{agentSessionID: state},
	}
	client := &mindfsClient{proc: proc}

	err := client.SessionUpdate(context.Background(), acpsdk.SessionNotification{
		SessionId: acpsdk.SessionId(agentSessionID),
		Update: acpsdk.SessionUpdate{
			AvailableCommandsUpdate: &acpsdk.SessionAvailableCommandsUpdate{
				AvailableCommands: []acpsdk.AvailableCommand{{
					Name:        "resume",
					Description: "Resume a previous session",
				}},
			},
		},
	})
	if err != nil {
		t.Fatalf("SessionUpdate returned error: %v", err)
	}

	commands := state.getCommands()
	if len(commands) != 1 || commands[0].Name != "resume" {
		t.Fatalf("expected available command to be stored without handler, got %#v", commands)
	}
}

func TestResumeSessionReservationReceivesEarlyAvailableCommands(t *testing.T) {
	const sessionKey = "test-session"
	const agentSessionID = "agent-session"
	proc := &Process{
		sessions:     map[string]*sessionState{},
		sessionsByID: map[string]*sessionState{},
	}
	state, ok := proc.reserveResumeSession(sessionKey, acpsdk.SessionId(agentSessionID))
	if !ok {
		t.Fatal("expected resume session reservation")
	}
	client := &mindfsClient{proc: proc}

	err := client.SessionUpdate(context.Background(), acpsdk.SessionNotification{
		SessionId: acpsdk.SessionId(agentSessionID),
		Update: acpsdk.SessionUpdate{
			AvailableCommandsUpdate: &acpsdk.SessionAvailableCommandsUpdate{
				AvailableCommands: []acpsdk.AvailableCommand{{Name: "resume"}},
			},
		},
	})
	if err != nil {
		t.Fatalf("SessionUpdate returned error: %v", err)
	}
	commands := state.getCommands()
	if len(commands) != 1 || commands[0].Name != "resume" {
		t.Fatalf("expected early available command on reserved session, got %#v", commands)
	}
}

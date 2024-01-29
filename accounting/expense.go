package accounting

import (
	"fmt"
	"strconv"
)

var nextExpenseId = 1000

type Expense struct {
	ID           string
	Name         string
	Owner        *Participant
	Participants []*Participant
}

func MakeExpense(name string, owner *Participant, participants ...*Participant) (*Expense, error) {
	if owner == nil {
		return nil, fmt.Errorf("owner cannot be nil")
	}
	if len(participants) < 1 {
		return nil, fmt.Errorf("must be at least two participant")
	}

	e := &Expense{
		ID:           strconv.Itoa(nextExpenseId),
		Name:         name,
		Owner:        owner,
		Participants: participants,
	}
	nextExpenseId += 1
	return e, nil
}

func MakeEvenSplitExpense(name string, owner string, cost int, participants ...string) (*Expense, error) {
	// The participants borrow money from the owner, so the owner is not a participant
	ps := make([]*Participant, len(participants))
	chunks := splitCostEvenly(cost, 1+len(participants))
	participantCost := 0
	for i, p := range participants {
		ps[i] = &Participant{
			Account: p,
			Cost:    chunks[i+1],
		}
		participantCost += chunks[i+1]
	}
	ownerParticpant := &Participant{
		Account: owner,
		Cost:    cost,
	}

	return MakeExpense(name, ownerParticpant, ps...)
}

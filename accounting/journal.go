package accounting

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/jedib0t/go-pretty/v6/table"
)

type DebtMap map[string]map[string]int

type Journal struct {
	Entries  []*Entry
	Expenses map[string]*Expense
}

func (j *Journal) EntriesForExpenseID(id string) []*Entry {
	entries := make([]*Entry, 0)
	for _, entry := range j.Entries {
		if entry.ExpenseID == id {
			entries = append(entries, entry)
		}
	}
	return entries
}

func (j *Journal) AddEzExpense(name string, owner string, cost int, participants ...string) error {
	e, err := MakeEvenSplitExpense(name, owner, cost, participants...)
	if err != nil {
		return err
	}
	return j.RegisterExpense(e)
}

func (j *Journal) RegisterExpense(expense *Expense) error {
	if _, ok := j.Expenses[expense.ID]; ok {
		return fmt.Errorf("expense %s already registered", expense.ID)
	}

	j.Expenses[expense.ID] = expense
	participantCosts := 0
	for _, p := range expense.Participants {
		pEntry := &Entry{
			Account:   p.Account,
			ExpenseID: expense.ID,
			Debit:     p.Cost,
		}
		participantCosts += p.Cost
		j.Entries = append(j.Entries, pEntry)
	}

	ownerEntry := &Entry{
		Account:   expense.Owner.Account,
		ExpenseID: expense.ID,
		Credit:    participantCosts,
	}
	j.Entries = append(j.Entries, ownerEntry)

	return nil
}

func ReportDebts(debts DebtMap) {
	dt := list.NewWriter()
	dt.SetStyle(list.StyleConnectedRounded)
	for owner, debtors := range debts {
		dt.AppendItem("to " + owner)
		dt.Indent()
		for debtor, cost := range debtors {
			dt.AppendItem(fmt.Sprintf("%s owes %d", debtor, cost))
		}
		dt.UnIndent()
	}

	fmt.Println()
	fmt.Println(dt.Render())
}

func (j *Journal) Report() {
	jt := table.NewWriter()
	jt.Style().Options.DrawBorder = false
	jt.Style().Options.SeparateColumns = false
	jt.Style().Options.SeparateFooter = true
	jt.Style().Options.SeparateHeader = true
	jt.Style().Options.SeparateRows = false
	jt.SetTitle("// JOURNAL")
	jt.AppendHeader(table.Row{"Account", "ExpenseID", "Debit", "Credit"})
	debits := 0
	credits := 0
	jt.SortBy([]table.SortBy{{Name: "ExpenseID"}})
	for _, e := range j.Entries {
		debits += e.Debit
		credits += e.Credit
		jt.AppendRow(table.Row{e.Account, e.ExpenseID, e.Debit, e.Credit})
	}
	jt.AppendFooter(table.Row{"Balance", "", debits, credits})
	fmt.Println(jt.Render())
	fmt.Println()

	et := table.NewWriter()
	et.Style().Options.DrawBorder = false
	et.Style().Options.SeparateColumns = false
	et.Style().Options.SeparateFooter = true
	et.Style().Options.SeparateHeader = true
	et.Style().Options.SeparateRows = false
	et.SetTitle("// EXPENSES")
	et.AppendHeader(table.Row{"ExpenseID", "Description", "Owner", "Cost"})
	et.SetIndexColumn(0)
	for _, e := range j.Expenses {
		et.AppendRow(table.Row{e.ID, e.Name, e.Owner.Account, e.Owner.Cost})
	}
	fmt.Println(et.Render())

	debts, err := j.CalculateDebts()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ReportDebts(debts)
}

func (j *Journal) CalculateDebts() (DebtMap, error) {
	// Map of owner -> debtors -> cost
	debts := make(DebtMap)

	// Go through every entry
	for _, entry := range j.Entries {
		// where someone has borrowed money
		if entry.Debit > 0 {
			expense, ok := j.Expenses[entry.ExpenseID]
			if !ok {
				return debts, fmt.Errorf("found entry %v with an invalid expense ID: %s", entry, entry.ExpenseID)
			}
			if _, ok := debts[expense.Owner.Account]; !ok {
				debts[expense.Owner.Account] = make(map[string]int)
			}

			// and add it to their debts to the owner of the expense
			debts[expense.Owner.Account][entry.Account] += entry.Debit
		}
	}

	// Now go over each debt and cancel out differences

	// For every person who has people that owe them money
	for owner, debtors := range debts {
		// Go through the people that owe them money
		for debtor, cost := range debtors {
			if cost <= 0 {
				continue
			}

			// And if there is a debt in the reverse direction
			if revcost, ok := debts[debtor][owner]; ok && revcost > 0 {
				// subtract from that debtors total of what they owe that person
				// if it's less than the total cost
				if revcost <= cost {
					diff := cost - revcost
					debts[owner][debtor] = diff
					debts[debtor][owner] -= revcost
				}
			}
		}
	}

	return debts, nil
}

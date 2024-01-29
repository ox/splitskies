package accounting

import (
	"fmt"
	"reflect"
	"testing"
)

type Split struct {
	name         string
	owner        string
	cost         int
	participants []string
}

func genDebts(splits []Split) (DebtMap, error) {
	j := &Journal{
		Entries:  make([]*Entry, 0),
		Expenses: make(map[string]*Expense),
	}

	for _, split := range splits {
		err := j.AddEzExpense(split.name, split.owner, split.cost, split.participants...)
		if err != nil {
			return nil, err
		}
	}

	return j.CalculateDebts()
}

func TestCalculateDebts(t *testing.T) {

	tests := []struct {
		splits   []Split
		expected DebtMap
	}{
		{
			[]Split{
				{"Beers", "Artem", 100, []string{"Bob"}},
				{"Dinner", "Bob", 100, []string{"Artem"}},
				{"Dessert", "Artem", 10, []string{"Bob"}},
			},
			DebtMap{
				"Artem": map[string]int{
					"Bob": 5,
				},
				"Bob": map[string]int{
					"Artem": 0,
				},
			},
		},
		{
			[]Split{
				{"AirBnB", "Artem", 1500, []string{"Abby", "Taylor", "Jess", "Eric"}},
				{"Dinner", "Abby", 100, []string{"Taylor", "Jess"}},
				{"Beer", "Eric", 200, []string{"Artem", "Abby", "Taylor", "Jess"}},
			},
			DebtMap{
				"Artem": map[string]int{
					"Abby":   300,
					"Taylor": 300,
					"Jess":   300,
					"Eric":   260,
				},
				"Abby": map[string]int{
					"Taylor": 33,
					"Jess":   34,
				},
				"Eric": map[string]int{
					"Artem":  0,
					"Abby":   40,
					"Taylor": 40,
					"Jess":   40,
				},
			},
		},
	}

	for _, debttest := range tests {
		debts, err := genDebts(debttest.splits)
		if err != nil {
			t.Error(err)
		}

		expected := debttest.expected
		if !reflect.DeepEqual(expected, debts) {
			fmt.Printf("expected %v, got %v", expected, debts)
			t.FailNow()
		}
	}

}

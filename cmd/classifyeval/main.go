// Command classifyeval replays labelled transactions through the real classifier,
// so a prompt, taxonomy or model change can be measured instead of eyeballed.
//
//	MORPH_AI_KEY=... go run ./cmd/classifyeval
//	MORPH_AI_MODEL=gpt-5.4-nano go run ./cmd/classifyeval -runs 3
//
// Cases live in testdata/cases.jsonl, one JSON object per line, captured from
// real traffic. Add a line whenever the bot gets something wrong.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/maximbilan/mcc"
	"github.com/morph/internal/aiservice"
	"github.com/morph/internal/category"
	"github.com/morph/third_party/openai"
)

type evalCase struct {
	// "mono" for a card transaction, "notification" for a bank push.
	Kind string `json:"kind"`
	// Group scores a slice of the set separately, e.g. transfers.
	Group string `json:"group"`

	MCC      int32   `json:"mcc,omitempty"`
	Merchant string  `json:"merchant,omitempty"`
	Amount   float64 `json:"amount,omitempty"`

	App     string `json:"app,omitempty"`
	Title   string `json:"title,omitempty"`
	Message string `json:"message,omitempty"`

	// Expected path, empty when the input is not a transaction.
	Want          string `json:"want"`
	WantIsExpense bool   `json:"wantIsExpense"`
}

type outcome struct {
	Case evalCase
	Got  string
	IsTx bool
	OK   bool
}

func main() {
	runs := flag.Int("runs", 1, "how many times to replay the whole set, to expose run-to-run variance")
	workers := flag.Int("workers", 8, "concurrent requests")
	casesPath := flag.String("cases", "cmd/classifyeval/testdata/cases.jsonl", "path to the labelled cases")
	verbose := flag.Bool("v", false, "print every miss")
	flag.Parse()

	if os.Getenv("MORPH_AI_KEY") == "" {
		log.Fatal("MORPH_AI_KEY is not set")
	}

	cases, err := loadCases(*casesPath)
	if err != nil {
		log.Fatalf("could not load cases: %v", err)
	}
	fmt.Printf("%d cases, %d run(s)\n\n", len(cases), *runs)

	for run := 1; run <= *runs; run++ {
		outcomes := replay(cases, *workers)
		report(run, outcomes, *verbose)
	}
}

func loadCases(path string) ([]evalCase, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cases []evalCase
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		var c evalCase
		if err := json.Unmarshal([]byte(line), &c); err != nil {
			return nil, fmt.Errorf("%q: %w", line, err)
		}
		cases = append(cases, c)
	}
	return cases, scanner.Err()
}

// prompts mirrors what the handlers send, so this measures production behaviour.
func prompts(c evalCase) (system string, user string) {
	system = category.ClassificationPrompt()
	switch c.Kind {
	case "notification":
		system += "\n\nThis input is a bank push notification. Set isTransaction to false for promotional, marketing, security, login and other informational messages, and true only for a real debit or credit. Read the amount out of the notification text."
		user = fmt.Sprintf("Classify this bank push notification.\nApp: %s\nTitle: %s\nMessage: %s", c.App, c.Title, c.Message)
	default:
		description := ""
		if got, err := mcc.GetCategory(fmt.Sprint(c.MCC)); err == nil {
			description = got
		}
		user = fmt.Sprintf("Classify this bank transaction.\nMerchant: %s\nMCC: %d (%s)\nAmount: %.2f",
			c.Merchant, c.MCC, description, c.Amount)
	}
	return system, user
}

func replay(cases []evalCase, workers int) []outcome {
	var service aiservice.AIService = openai.OpenAI{}
	paths := category.Paths()

	outcomes := make([]outcome, len(cases))
	queue := make(chan int)
	var wg sync.WaitGroup

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range queue {
				c := cases[i]
				system, user := prompts(c)
				ctx := context.Background()
				response := service.Classify(aiservice.Request{
					Name:         "MorphEval",
					Description:  "Classifies a transaction into: Category, Subcategory, Amount",
					SystemPrompt: system,
					UserPrompt:   user,
					AllowedPaths: paths,
				}, &ctx)

				result := outcome{Case: c, Got: "<no response>"}
				if response != nil {
					name, sub := category.SplitPath(response.CategoryPath)
					result.Got = name
					if sub != "" {
						result.Got = name + "/" + sub
					}
					result.IsTx = response.IsTransaction
					result.OK = result.IsTx == c.WantIsExpense && (!c.WantIsExpense || result.Got == c.Want)
				}
				outcomes[i] = result
			}
		}()
	}

	for i := range cases {
		queue <- i
	}
	close(queue)
	wg.Wait()
	return outcomes
}

func report(run int, outcomes []outcome, verbose bool) {
	// scored counts only cases whose path reaches MoneyWiz; a non-transaction
	// push's path is discarded, so counting it as "Other" would inflate the number.
	type tally struct{ total, ok, scored, other int }
	byGroup := map[string]*tally{}
	all := &tally{}

	for _, o := range outcomes {
		group := o.Case.Group
		if group == "" {
			group = "ungrouped"
		}
		if byGroup[group] == nil {
			byGroup[group] = &tally{}
		}
		isOther := o.Got == "Other" || o.Got == "Waste" || strings.HasSuffix(o.Got, "/Other")
		for _, t := range []*tally{byGroup[group], all} {
			t.total++
			if o.OK {
				t.ok++
			}
			if o.Case.WantIsExpense {
				t.scored++
				if isOther {
					t.other++
				}
			}
		}
	}

	groups := make([]string, 0, len(byGroup))
	for group := range byGroup {
		groups = append(groups, group)
	}
	sort.Strings(groups)

	fmt.Printf("run %d\n", run)
	for _, group := range groups {
		t := byGroup[group]
		fmt.Printf("  %-12s n=%-3d correct=%3.0f%%  landed on an Other leaf=%3.0f%%\n",
			group, t.total, pct(t.ok, t.total), pct(t.other, t.scored))
	}
	fmt.Printf("  %-12s n=%-3d correct=%3.0f%%  landed on an Other leaf=%3.0f%%\n",
		"TOTAL", all.total, pct(all.ok, all.total), pct(all.other, all.scored))

	if verbose {
		for _, o := range outcomes {
			if o.OK {
				continue
			}
			label := o.Case.Merchant
			if label == "" {
				label = o.Case.Message
			}
			fmt.Printf("    MISS %-42s want=%-24s got=%s\n", truncate(label, 42), o.Case.Want, o.Got)
		}
	}
	fmt.Println()
}

func pct(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return 100 * float64(part) / float64(total)
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}

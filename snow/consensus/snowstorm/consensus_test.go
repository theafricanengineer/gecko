// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package snowstorm

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/ava-labs/gecko/ids"
	"github.com/ava-labs/gecko/snow"
	"github.com/ava-labs/gecko/snow/choices"
	"github.com/ava-labs/gecko/snow/consensus/snowball"
)

var (
	Red   = &TestTx{Identifier: ids.Empty.Prefix(0)}
	Green = &TestTx{Identifier: ids.Empty.Prefix(1)}
	Blue  = &TestTx{Identifier: ids.Empty.Prefix(2)}
	Alpha = &TestTx{Identifier: ids.Empty.Prefix(3)}
)

//  R - G - B - A

func init() {
	X := ids.Empty.Prefix(4)
	Y := ids.Empty.Prefix(5)
	Z := ids.Empty.Prefix(6)

	Red.Ins.Add(X)

	Green.Ins.Add(X)
	Green.Ins.Add(Y)

	Blue.Ins.Add(Y)
	Blue.Ins.Add(Z)

	Alpha.Ins.Add(Z)
}

func Setup() {
	Red.Reset()
	Green.Reset()
	Blue.Reset()
	Alpha.Reset()
}

func ParamsTest(t *testing.T, factory Factory) {
	Setup()

	graph := factory.New()

	params := snowball.Parameters{
		Metrics: prometheus.NewRegistry(),
		K:       2, Alpha: 2, BetaVirtuous: 1, BetaRogue: 2,
	}
	graph.Initialize(snow.DefaultContextTest(), params)

	if p := graph.Parameters(); p.K != params.K {
		t.Fatalf("Wrong K parameter")
	} else if p := graph.Parameters(); p.Alpha != params.Alpha {
		t.Fatalf("Wrong Alpha parameter")
	} else if p := graph.Parameters(); p.BetaVirtuous != params.BetaVirtuous {
		t.Fatalf("Wrong Beta1 parameter")
	} else if p := graph.Parameters(); p.BetaRogue != params.BetaRogue {
		t.Fatalf("Wrong Beta2 parameter")
	}
}

func IssuedTest(t *testing.T, factory Factory) {
	Setup()

	graph := factory.New()

	params := snowball.Parameters{
		Metrics: prometheus.NewRegistry(),
		K:       2, Alpha: 2, BetaVirtuous: 1, BetaRogue: 1,
	}
	graph.Initialize(snow.DefaultContextTest(), params)

	if issued := graph.Issued(Red); issued {
		t.Fatalf("Haven't issued anything yet.")
	}

	graph.Add(Red)

	if issued := graph.Issued(Red); !issued {
		t.Fatalf("Have already issued.")
	}

	Blue.Accept()

	if issued := graph.Issued(Blue); !issued {
		t.Fatalf("Have already accepted.")
	}
}

func LeftoverInputTest(t *testing.T, factory Factory) {
	Setup()

	graph := factory.New()

	params := snowball.Parameters{
		Metrics: prometheus.NewRegistry(),
		K:       2, Alpha: 2, BetaVirtuous: 1, BetaRogue: 1,
	}
	graph.Initialize(snow.DefaultContextTest(), params)
	graph.Add(Red)
	graph.Add(Green)

	if prefs := graph.Preferences(); prefs.Len() != 1 {
		t.Fatalf("Wrong number of preferences.")
	} else if !prefs.Contains(Red.ID()) {
		t.Fatalf("Wrong preference. Expected %s got %s", Red.ID(), prefs.List()[0])
	} else if graph.Finalized() {
		t.Fatalf("Finalized too early")
	}

	r := ids.Bag{}
	r.SetThreshold(2)
	r.AddCount(Red.ID(), 2)
	graph.RecordPoll(r)

	if prefs := graph.Preferences(); prefs.Len() != 0 {
		t.Fatalf("Wrong number of preferences.")
	} else if !graph.Finalized() {
		t.Fatalf("Finalized too late")
	}

	if Red.Status() != choices.Accepted {
		t.Fatalf("%s should have been accepted", Red.ID())
	} else if Green.Status() != choices.Rejected {
		t.Fatalf("%s should have been rejected", Green.ID())
	}
}

func LowerConfidenceTest(t *testing.T, factory Factory) {
	Setup()

	graph := factory.New()

	params := snowball.Parameters{
		Metrics: prometheus.NewRegistry(),
		K:       2, Alpha: 2, BetaVirtuous: 1, BetaRogue: 1,
	}
	graph.Initialize(snow.DefaultContextTest(), params)
	graph.Add(Red)
	graph.Add(Green)
	graph.Add(Blue)

	if prefs := graph.Preferences(); prefs.Len() != 1 {
		t.Fatalf("Wrong number of preferences.")
	} else if !prefs.Contains(Red.ID()) {
		t.Fatalf("Wrong preference. Expected %s got %s", Red.ID(), prefs.List()[0])
	} else if graph.Finalized() {
		t.Fatalf("Finalized too early")
	}

	r := ids.Bag{}
	r.SetThreshold(2)
	r.AddCount(Red.ID(), 2)
	graph.RecordPoll(r)

	if prefs := graph.Preferences(); prefs.Len() != 1 {
		t.Fatalf("Wrong number of preferences.")
	} else if !prefs.Contains(Blue.ID()) {
		t.Fatalf("Wrong preference. Expected %s", Blue.ID())
	} else if graph.Finalized() {
		t.Fatalf("Finalized too early")
	}
}

func MiddleConfidenceTest(t *testing.T, factory Factory) {
	Setup()

	graph := factory.New()

	params := snowball.Parameters{
		Metrics: prometheus.NewRegistry(),
		K:       2, Alpha: 2, BetaVirtuous: 1, BetaRogue: 1,
	}
	graph.Initialize(snow.DefaultContextTest(), params)
	graph.Add(Red)
	graph.Add(Green)
	graph.Add(Alpha)
	graph.Add(Blue)

	if prefs := graph.Preferences(); prefs.Len() != 2 {
		t.Fatalf("Wrong number of preferences.")
	} else if !prefs.Contains(Red.ID()) {
		t.Fatalf("Wrong preference. Expected %s", Red.ID())
	} else if !prefs.Contains(Alpha.ID()) {
		t.Fatalf("Wrong preference. Expected %s", Alpha.ID())
	} else if graph.Finalized() {
		t.Fatalf("Finalized too early")
	}

	r := ids.Bag{}
	r.SetThreshold(2)
	r.AddCount(Red.ID(), 2)
	graph.RecordPoll(r)

	if prefs := graph.Preferences(); prefs.Len() != 1 {
		t.Fatalf("Wrong number of preferences.")
	} else if !prefs.Contains(Alpha.ID()) {
		t.Fatalf("Wrong preference. Expected %s", Alpha.ID())
	} else if graph.Finalized() {
		t.Fatalf("Finalized too early")
	}
}
func IndependentTest(t *testing.T, factory Factory) {
	Setup()

	graph := factory.New()

	params := snowball.Parameters{
		Metrics: prometheus.NewRegistry(),
		K:       2, Alpha: 2, BetaVirtuous: 2, BetaRogue: 2,
	}
	graph.Initialize(snow.DefaultContextTest(), params)
	graph.Add(Red)
	graph.Add(Alpha)

	if prefs := graph.Preferences(); prefs.Len() != 2 {
		t.Fatalf("Wrong number of preferences.")
	} else if !prefs.Contains(Red.ID()) {
		t.Fatalf("Wrong preference. Expected %s", Red.ID())
	} else if !prefs.Contains(Alpha.ID()) {
		t.Fatalf("Wrong preference. Expected %s", Alpha.ID())
	} else if graph.Finalized() {
		t.Fatalf("Finalized too early")
	}

	ra := ids.Bag{}
	ra.SetThreshold(2)
	ra.AddCount(Red.ID(), 2)
	ra.AddCount(Alpha.ID(), 2)
	graph.RecordPoll(ra)

	if prefs := graph.Preferences(); prefs.Len() != 2 {
		t.Fatalf("Wrong number of preferences.")
	} else if !prefs.Contains(Red.ID()) {
		t.Fatalf("Wrong preference. Expected %s", Red.ID())
	} else if !prefs.Contains(Alpha.ID()) {
		t.Fatalf("Wrong preference. Expected %s", Alpha.ID())
	} else if graph.Finalized() {
		t.Fatalf("Finalized too early")
	}

	graph.RecordPoll(ra)

	if prefs := graph.Preferences(); prefs.Len() != 0 {
		t.Fatalf("Wrong number of preferences.")
	} else if !graph.Finalized() {
		t.Fatalf("Finalized too late")
	}
}

func VirtuousTest(t *testing.T, factory Factory) {
	Setup()

	graph := factory.New()

	params := snowball.Parameters{
		Metrics: prometheus.NewRegistry(),
		K:       2, Alpha: 2, BetaVirtuous: 1, BetaRogue: 1,
	}
	graph.Initialize(snow.DefaultContextTest(), params)
	graph.Add(Red)

	if virtuous := graph.Virtuous(); virtuous.Len() != 1 {
		t.Fatalf("Wrong number of virtuous.")
	} else if !virtuous.Contains(Red.ID()) {
		t.Fatalf("Wrong virtuous. Expected %s", Red.ID())
	}

	graph.Add(Alpha)

	if virtuous := graph.Virtuous(); virtuous.Len() != 2 {
		t.Fatalf("Wrong number of virtuous.")
	} else if !virtuous.Contains(Red.ID()) {
		t.Fatalf("Wrong virtuous. Expected %s", Red.ID())
	} else if !virtuous.Contains(Alpha.ID()) {
		t.Fatalf("Wrong virtuous. Expected %s", Alpha.ID())
	}

	graph.Add(Green)

	if virtuous := graph.Virtuous(); virtuous.Len() != 1 {
		t.Fatalf("Wrong number of virtuous.")
	} else if !virtuous.Contains(Alpha.ID()) {
		t.Fatalf("Wrong virtuous. Expected %s", Alpha.ID())
	}

	graph.Add(Blue)

	if virtuous := graph.Virtuous(); virtuous.Len() != 0 {
		t.Fatalf("Wrong number of virtuous.")
	}
}

func IsVirtuousTest(t *testing.T, factory Factory) {
	Setup()

	graph := factory.New()

	params := snowball.Parameters{
		Metrics: prometheus.NewRegistry(),
		K:       2, Alpha: 2, BetaVirtuous: 1, BetaRogue: 1,
	}
	graph.Initialize(snow.DefaultContextTest(), params)

	if !graph.IsVirtuous(Red) {
		t.Fatalf("Should be virtuous")
	} else if !graph.IsVirtuous(Green) {
		t.Fatalf("Should be virtuous")
	} else if !graph.IsVirtuous(Blue) {
		t.Fatalf("Should be virtuous")
	} else if !graph.IsVirtuous(Alpha) {
		t.Fatalf("Should be virtuous")
	}

	graph.Add(Red)

	if !graph.IsVirtuous(Red) {
		t.Fatalf("Should be virtuous")
	} else if graph.IsVirtuous(Green) {
		t.Fatalf("Should not be virtuous")
	} else if !graph.IsVirtuous(Blue) {
		t.Fatalf("Should be virtuous")
	} else if !graph.IsVirtuous(Alpha) {
		t.Fatalf("Should be virtuous")
	}

	graph.Add(Green)

	if graph.IsVirtuous(Red) {
		t.Fatalf("Should not be virtuous")
	} else if graph.IsVirtuous(Green) {
		t.Fatalf("Should not be virtuous")
	} else if graph.IsVirtuous(Blue) {
		t.Fatalf("Should not be virtuous")
	}
}

func QuiesceTest(t *testing.T, factory Factory) {
	Setup()

	graph := factory.New()

	params := snowball.Parameters{
		Metrics: prometheus.NewRegistry(),
		K:       2, Alpha: 2, BetaVirtuous: 1, BetaRogue: 1,
	}
	graph.Initialize(snow.DefaultContextTest(), params)

	if !graph.Quiesce() {
		t.Fatalf("Should quiesce")
	}

	graph.Add(Red)

	if graph.Quiesce() {
		t.Fatalf("Shouldn't quiesce")
	}

	graph.Add(Green)

	if !graph.Quiesce() {
		t.Fatalf("Should quiesce")
	}
}

func AcceptingDependencyTest(t *testing.T, factory Factory) {
	Setup()

	graph := factory.New()

	purple := &TestTx{
		Identifier: ids.Empty.Prefix(7),
		Stat:       choices.Processing,
	}
	purple.Ins.Add(ids.Empty.Prefix(8))
	purple.Deps = []Tx{Red}

	params := snowball.Parameters{
		Metrics: prometheus.NewRegistry(),
		K:       1, Alpha: 1, BetaVirtuous: 1, BetaRogue: 2,
	}
	graph.Initialize(snow.DefaultContextTest(), params)

	graph.Add(Red)
	graph.Add(Green)
	graph.Add(purple)

	if prefs := graph.Preferences(); prefs.Len() != 2 {
		t.Fatalf("Wrong number of preferences.")
	} else if !prefs.Contains(Red.ID()) {
		t.Fatalf("Wrong preference. Expected %s", Red.ID())
	} else if !prefs.Contains(purple.ID()) {
		t.Fatalf("Wrong preference. Expected %s", purple.ID())
	} else if Red.Status() != choices.Processing {
		t.Fatalf("Wrong status. %s should be %s", Red.ID(), choices.Processing)
	} else if Green.Status() != choices.Processing {
		t.Fatalf("Wrong status. %s should be %s", Green.ID(), choices.Processing)
	} else if purple.Status() != choices.Processing {
		t.Fatalf("Wrong status. %s should be %s", purple.ID(), choices.Processing)
	}

	g := ids.Bag{}
	g.Add(Green.ID())

	graph.RecordPoll(g)

	if prefs := graph.Preferences(); prefs.Len() != 2 {
		t.Fatalf("Wrong number of preferences.")
	} else if !prefs.Contains(Green.ID()) {
		t.Fatalf("Wrong preference. Expected %s", Green.ID())
	} else if !prefs.Contains(purple.ID()) {
		t.Fatalf("Wrong preference. Expected %s", purple.ID())
	} else if Red.Status() != choices.Processing {
		t.Fatalf("Wrong status. %s should be %s", Red.ID(), choices.Processing)
	} else if Green.Status() != choices.Processing {
		t.Fatalf("Wrong status. %s should be %s", Green.ID(), choices.Processing)
	} else if purple.Status() != choices.Processing {
		t.Fatalf("Wrong status. %s should be %s", purple.ID(), choices.Processing)
	}

	rp := ids.Bag{}
	rp.Add(Red.ID(), purple.ID())

	graph.RecordPoll(rp)

	if prefs := graph.Preferences(); prefs.Len() != 2 {
		t.Fatalf("Wrong number of preferences.")
	} else if !prefs.Contains(Green.ID()) {
		t.Fatalf("Wrong preference. Expected %s", Green.ID())
	} else if !prefs.Contains(purple.ID()) {
		t.Fatalf("Wrong preference. Expected %s", purple.ID())
	} else if Red.Status() != choices.Processing {
		t.Fatalf("Wrong status. %s should be %s", Red.ID(), choices.Processing)
	} else if Green.Status() != choices.Processing {
		t.Fatalf("Wrong status. %s should be %s", Green.ID(), choices.Processing)
	} else if purple.Status() != choices.Processing {
		t.Fatalf("Wrong status. %s should be %s", purple.ID(), choices.Processing)
	}

	r := ids.Bag{}
	r.Add(Red.ID())

	graph.RecordPoll(r)

	if prefs := graph.Preferences(); prefs.Len() != 0 {
		t.Fatalf("Wrong number of preferences.")
	} else if Red.Status() != choices.Accepted {
		t.Fatalf("Wrong status. %s should be %s", Red.ID(), choices.Accepted)
	} else if Green.Status() != choices.Rejected {
		t.Fatalf("Wrong status. %s should be %s", Green.ID(), choices.Rejected)
	} else if purple.Status() != choices.Accepted {
		t.Fatalf("Wrong status. %s should be %s", purple.ID(), choices.Accepted)
	}
}

func RejectingDependencyTest(t *testing.T, factory Factory) {
	Setup()

	graph := factory.New()

	purple := &TestTx{
		Identifier: ids.Empty.Prefix(7),
		Stat:       choices.Processing,
	}
	purple.Ins.Add(ids.Empty.Prefix(8))
	purple.Deps = []Tx{Red, Blue}

	params := snowball.Parameters{
		Metrics: prometheus.NewRegistry(),
		K:       1, Alpha: 1, BetaVirtuous: 1, BetaRogue: 2,
	}
	graph.Initialize(snow.DefaultContextTest(), params)

	graph.Add(Red)
	graph.Add(Green)
	graph.Add(Blue)
	graph.Add(purple)

	if prefs := graph.Preferences(); prefs.Len() != 2 {
		t.Fatalf("Wrong number of preferences.")
	} else if !prefs.Contains(Red.ID()) {
		t.Fatalf("Wrong preference. Expected %s", Red.ID())
	} else if !prefs.Contains(purple.ID()) {
		t.Fatalf("Wrong preference. Expected %s", purple.ID())
	} else if Red.Status() != choices.Processing {
		t.Fatalf("Wrong status. %s should be %s", Red.ID(), choices.Processing)
	} else if Green.Status() != choices.Processing {
		t.Fatalf("Wrong status. %s should be %s", Green.ID(), choices.Processing)
	} else if Blue.Status() != choices.Processing {
		t.Fatalf("Wrong status. %s should be %s", Blue.ID(), choices.Processing)
	} else if purple.Status() != choices.Processing {
		t.Fatalf("Wrong status. %s should be %s", purple.ID(), choices.Processing)
	}

	gp := ids.Bag{}
	gp.Add(Green.ID(), purple.ID())

	graph.RecordPoll(gp)

	if prefs := graph.Preferences(); prefs.Len() != 2 {
		t.Fatalf("Wrong number of preferences.")
	} else if !prefs.Contains(Green.ID()) {
		t.Fatalf("Wrong preference. Expected %s", Green.ID())
	} else if !prefs.Contains(purple.ID()) {
		t.Fatalf("Wrong preference. Expected %s", purple.ID())
	} else if Red.Status() != choices.Processing {
		t.Fatalf("Wrong status. %s should be %s", Red.ID(), choices.Processing)
	} else if Green.Status() != choices.Processing {
		t.Fatalf("Wrong status. %s should be %s", Green.ID(), choices.Processing)
	} else if Blue.Status() != choices.Processing {
		t.Fatalf("Wrong status. %s should be %s", Blue.ID(), choices.Processing)
	} else if purple.Status() != choices.Processing {
		t.Fatalf("Wrong status. %s should be %s", purple.ID(), choices.Processing)
	}

	graph.RecordPoll(gp)

	if prefs := graph.Preferences(); prefs.Len() != 0 {
		t.Fatalf("Wrong number of preferences.")
	} else if Red.Status() != choices.Rejected {
		t.Fatalf("Wrong status. %s should be %s", Red.ID(), choices.Rejected)
	} else if Green.Status() != choices.Accepted {
		t.Fatalf("Wrong status. %s should be %s", Green.ID(), choices.Accepted)
	} else if Blue.Status() != choices.Rejected {
		t.Fatalf("Wrong status. %s should be %s", Blue.ID(), choices.Rejected)
	} else if purple.Status() != choices.Rejected {
		t.Fatalf("Wrong status. %s should be %s", purple.ID(), choices.Rejected)
	}
}

func VacuouslyAcceptedTest(t *testing.T, factory Factory) {
	Setup()

	graph := factory.New()

	purple := &TestTx{
		Identifier: ids.Empty.Prefix(7),
		Stat:       choices.Processing,
	}

	params := snowball.Parameters{
		Metrics: prometheus.NewRegistry(),
		K:       1, Alpha: 1, BetaVirtuous: 1, BetaRogue: 2,
	}
	graph.Initialize(snow.DefaultContextTest(), params)

	graph.Add(purple)

	if prefs := graph.Preferences(); prefs.Len() != 0 {
		t.Fatalf("Wrong number of preferences.")
	} else if status := purple.Status(); status != choices.Accepted {
		t.Fatalf("Wrong status. %s should be %s", purple.ID(), choices.Accepted)
	}
}

func ConflictsTest(t *testing.T, factory Factory) {
	Setup()

	graph := factory.New()

	params := snowball.Parameters{
		Metrics: prometheus.NewRegistry(),
		K:       1, Alpha: 1, BetaVirtuous: 1, BetaRogue: 2,
	}
	graph.Initialize(snow.DefaultContextTest(), params)

	conflictInputID := ids.Empty.Prefix(0)

	insPurple := ids.Set{}
	insPurple.Add(conflictInputID)

	purple := &TestTx{
		Identifier: ids.Empty.Prefix(7),
		Stat:       choices.Processing,
		Ins:        insPurple,
	}

	insOrange := ids.Set{}
	insOrange.Add(conflictInputID)

	orange := &TestTx{
		Identifier: ids.Empty.Prefix(6),
		Stat:       choices.Processing,
		Ins:        insPurple,
	}

	graph.Add(purple)

	if orangeConflicts := graph.Conflicts(orange); orangeConflicts.Len() != 1 {
		t.Fatalf("Wrong number of conflicts")
	} else if !orangeConflicts.Contains(purple.Identifier) {
		t.Fatalf("Conflicts does not contain the right transaction")
	}

	graph.Add(orange)

	if orangeConflicts := graph.Conflicts(orange); orangeConflicts.Len() != 1 {
		t.Fatalf("Wrong number of conflicts")
	} else if !orangeConflicts.Contains(purple.Identifier) {
		t.Fatalf("Conflicts does not contain the right transaction")
	}
}

func VirtuousDependsOnRogueTest(t *testing.T, factory Factory) {
	Setup()

	graph := factory.New()

	params := snowball.Parameters{
		Metrics: prometheus.NewRegistry(),
		K:       1, Alpha: 1, BetaVirtuous: 1, BetaRogue: 2,
	}
	graph.Initialize(snow.DefaultContextTest(), params)

	rogue1 := &TestTx{
		Identifier: ids.Empty.Prefix(0),
		Stat:       choices.Processing,
	}
	rogue2 := &TestTx{
		Identifier: ids.Empty.Prefix(1),
		Stat:       choices.Processing,
	}
	virtuous := &TestTx{
		Identifier: ids.Empty.Prefix(2),
		Deps:       []Tx{rogue1},
		Stat:       choices.Processing,
	}

	input1 := ids.Empty.Prefix(3)
	input2 := ids.Empty.Prefix(4)

	rogue1.Ins.Add(input1)
	rogue2.Ins.Add(input1)

	virtuous.Ins.Add(input2)

	graph.Add(rogue1)
	graph.Add(rogue2)
	graph.Add(virtuous)

	votes := ids.Bag{}
	votes.Add(rogue1.ID())
	votes.Add(virtuous.ID())

	graph.RecordPoll(votes)

	if status := rogue1.Status(); status != choices.Processing {
		t.Fatalf("Rogue Tx is %s expected %s", status, choices.Processing)
	} else if status := rogue2.Status(); status != choices.Processing {
		t.Fatalf("Rogue Tx is %s expected %s", status, choices.Processing)
	} else if status := virtuous.Status(); status != choices.Processing {
		t.Fatalf("Virtuous Tx is %s expected %s", status, choices.Processing)
	} else if !graph.Quiesce() {
		t.Fatalf("Should quiesce as there are no pending virtuous transactions")
	}
}

func StringTest(t *testing.T, factory Factory, prefix string) {
	Setup()

	graph := factory.New()

	params := snowball.Parameters{
		Metrics: prometheus.NewRegistry(),
		K:       2, Alpha: 2, BetaVirtuous: 1, BetaRogue: 2,
	}
	graph.Initialize(snow.DefaultContextTest(), params)
	graph.Add(Red)
	graph.Add(Green)
	graph.Add(Blue)
	graph.Add(Alpha)

	if prefs := graph.Preferences(); prefs.Len() != 1 {
		t.Fatalf("Wrong number of preferences.")
	} else if !prefs.Contains(Red.ID()) {
		t.Fatalf("Wrong preference. Expected %s got %s", Red.ID(), prefs.List()[0])
	} else if graph.Finalized() {
		t.Fatalf("Finalized too early")
	}

	rb := ids.Bag{}
	rb.SetThreshold(2)
	rb.AddCount(Red.ID(), 2)
	rb.AddCount(Blue.ID(), 2)
	graph.RecordPoll(rb)
	graph.Add(Blue)

	{
		expected := prefix + "(\n" +
			"    Choice[0] = ID:  LUC1cmcxnfNR9LdkACS2ccGKLEK7SYqB4gLLTycQfg1koyfSq Confidence: 1 Bias: 1\n" +
			"    Choice[1] = ID:  TtF4d2QWbk5vzQGTEPrN48x6vwgAoAmKQ9cbp79inpQmcRKES Confidence: 0 Bias: 0\n" +
			"    Choice[2] = ID:  Zda4gsqTjRaX6XVZekVNi3ovMFPHDRQiGbzYuAb7Nwqy1rGBc Confidence: 0 Bias: 0\n" +
			"    Choice[3] = ID: 2mcwQKiD8VEspmMJpL1dc7okQQ5dDVAWeCBZ7FWBFAbxpv3t7w Confidence: 1 Bias: 1\n" +
			")"
		if str := graph.String(); str != expected {
			t.Fatalf("Expected %s, got %s", expected, str)
		}
	}

	if prefs := graph.Preferences(); prefs.Len() != 2 {
		t.Fatalf("Wrong number of preferences.")
	} else if !prefs.Contains(Red.ID()) {
		t.Fatalf("Wrong preference. Expected %s", Red.ID())
	} else if !prefs.Contains(Blue.ID()) {
		t.Fatalf("Wrong preference. Expected %s", Blue.ID())
	} else if graph.Finalized() {
		t.Fatalf("Finalized too early")
	}

	ga := ids.Bag{}
	ga.SetThreshold(2)
	ga.AddCount(Green.ID(), 2)
	ga.AddCount(Alpha.ID(), 2)
	graph.RecordPoll(ga)

	{
		expected := prefix + "(\n" +
			"    Choice[0] = ID:  LUC1cmcxnfNR9LdkACS2ccGKLEK7SYqB4gLLTycQfg1koyfSq Confidence: 0 Bias: 1\n" +
			"    Choice[1] = ID:  TtF4d2QWbk5vzQGTEPrN48x6vwgAoAmKQ9cbp79inpQmcRKES Confidence: 1 Bias: 1\n" +
			"    Choice[2] = ID:  Zda4gsqTjRaX6XVZekVNi3ovMFPHDRQiGbzYuAb7Nwqy1rGBc Confidence: 1 Bias: 1\n" +
			"    Choice[3] = ID: 2mcwQKiD8VEspmMJpL1dc7okQQ5dDVAWeCBZ7FWBFAbxpv3t7w Confidence: 0 Bias: 1\n" +
			")"
		if str := graph.String(); str != expected {
			t.Fatalf("Expected %s, got %s", expected, str)
		}
	}

	if prefs := graph.Preferences(); prefs.Len() != 2 {
		t.Fatalf("Wrong number of preferences.")
	} else if !prefs.Contains(Red.ID()) {
		t.Fatalf("Wrong preference. Expected %s", Red.ID())
	} else if !prefs.Contains(Blue.ID()) {
		t.Fatalf("Wrong preference. Expected %s", Blue.ID())
	} else if graph.Finalized() {
		t.Fatalf("Finalized too early")
	}

	empty := ids.Bag{}
	graph.RecordPoll(empty)

	{
		expected := prefix + "(\n" +
			"    Choice[0] = ID:  LUC1cmcxnfNR9LdkACS2ccGKLEK7SYqB4gLLTycQfg1koyfSq Confidence: 0 Bias: 1\n" +
			"    Choice[1] = ID:  TtF4d2QWbk5vzQGTEPrN48x6vwgAoAmKQ9cbp79inpQmcRKES Confidence: 0 Bias: 1\n" +
			"    Choice[2] = ID:  Zda4gsqTjRaX6XVZekVNi3ovMFPHDRQiGbzYuAb7Nwqy1rGBc Confidence: 0 Bias: 1\n" +
			"    Choice[3] = ID: 2mcwQKiD8VEspmMJpL1dc7okQQ5dDVAWeCBZ7FWBFAbxpv3t7w Confidence: 0 Bias: 1\n" +
			")"
		if str := graph.String(); str != expected {
			t.Fatalf("Expected %s, got %s", expected, str)
		}
	}

	if prefs := graph.Preferences(); prefs.Len() != 2 {
		t.Fatalf("Wrong number of preferences.")
	} else if !prefs.Contains(Red.ID()) {
		t.Fatalf("Wrong preference. Expected %s", Red.ID())
	} else if !prefs.Contains(Blue.ID()) {
		t.Fatalf("Wrong preference. Expected %s", Blue.ID())
	} else if graph.Finalized() {
		t.Fatalf("Finalized too early")
	}

	graph.RecordPoll(ga)

	{
		expected := prefix + "(\n" +
			"    Choice[0] = ID:  LUC1cmcxnfNR9LdkACS2ccGKLEK7SYqB4gLLTycQfg1koyfSq Confidence: 0 Bias: 1\n" +
			"    Choice[1] = ID:  TtF4d2QWbk5vzQGTEPrN48x6vwgAoAmKQ9cbp79inpQmcRKES Confidence: 1 Bias: 2\n" +
			"    Choice[2] = ID:  Zda4gsqTjRaX6XVZekVNi3ovMFPHDRQiGbzYuAb7Nwqy1rGBc Confidence: 1 Bias: 2\n" +
			"    Choice[3] = ID: 2mcwQKiD8VEspmMJpL1dc7okQQ5dDVAWeCBZ7FWBFAbxpv3t7w Confidence: 0 Bias: 1\n" +
			")"
		if str := graph.String(); str != expected {
			t.Fatalf("Expected %s, got %s", expected, str)
		}
	}

	if prefs := graph.Preferences(); prefs.Len() != 2 {
		t.Fatalf("Wrong number of preferences.")
	} else if !prefs.Contains(Green.ID()) {
		t.Fatalf("Wrong preference. Expected %s", Green.ID())
	} else if !prefs.Contains(Alpha.ID()) {
		t.Fatalf("Wrong preference. Expected %s", Alpha.ID())
	} else if graph.Finalized() {
		t.Fatalf("Finalized too early")
	}

	graph.RecordPoll(ga)

	{
		expected := prefix + "()"
		if str := graph.String(); str != expected {
			t.Fatalf("Expected %s, got %s", expected, str)
		}
	}

	if prefs := graph.Preferences(); prefs.Len() != 0 {
		t.Fatalf("Wrong number of preferences.")
	} else if !graph.Finalized() {
		t.Fatalf("Finalized too late")
	}

	if Green.Status() != choices.Accepted {
		t.Fatalf("%s should have been accepted", Green.ID())
	} else if Alpha.Status() != choices.Accepted {
		t.Fatalf("%s should have been accepted", Alpha.ID())
	} else if Red.Status() != choices.Rejected {
		t.Fatalf("%s should have been rejected", Red.ID())
	} else if Blue.Status() != choices.Rejected {
		t.Fatalf("%s should have been rejected", Blue.ID())
	}

	graph.RecordPoll(rb)

	{
		expected := prefix + "()"
		if str := graph.String(); str != expected {
			t.Fatalf("Expected %s, got %s", expected, str)
		}
	}

	if prefs := graph.Preferences(); prefs.Len() != 0 {
		t.Fatalf("Wrong number of preferences.")
	} else if !graph.Finalized() {
		t.Fatalf("Finalized too late")
	}

	if Green.Status() != choices.Accepted {
		t.Fatalf("%s should have been accepted", Green.ID())
	} else if Alpha.Status() != choices.Accepted {
		t.Fatalf("%s should have been accepted", Alpha.ID())
	} else if Red.Status() != choices.Rejected {
		t.Fatalf("%s should have been rejected", Red.ID())
	} else if Blue.Status() != choices.Rejected {
		t.Fatalf("%s should have been rejected", Blue.ID())
	}
}

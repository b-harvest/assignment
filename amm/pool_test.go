package amm_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"amm/amm"
)

func TestPoolPrice(t *testing.T) {
	for _, tc := range []struct {
		name   string
		rx, ry int64   // reserve balance
		ps     int64   // pool coin supply
		p      sdk.Dec // expected pool price
	}{
		{
			name: "normal pool",
			ps:   10000,
			rx:   20000,
			ry:   100,
			p:    sdk.NewDec(200),
		},
		{
			name: "decimal rounding",
			ps:   10000,
			rx:   200,
			ry:   300,
			p:    sdk.MustNewDecFromStr("0.666666666666666667"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool := amm.CreatePool(sdk.NewInt(tc.rx), sdk.NewInt(tc.ry), sdk.NewInt(tc.ps))
			require.True(sdk.DecEq(t, tc.p, pool.Price()))
		})
	}

	// panicking cases
	for _, tc := range []struct {
		rx, ry int64
		ps     int64
	}{
		{
			rx: 0,
			ry: 1000,
			ps: 1000,
		},
		{
			rx: 1000,
			ry: 0,
			ps: 1000,
		},
	} {
		t.Run("panics", func(t *testing.T) {
			require.Panics(t, func() {
				pool := amm.CreatePool(sdk.NewInt(tc.rx), sdk.NewInt(tc.ry), sdk.NewInt(tc.ps))
				pool.Price()
			})
		})
	}
}

func TestIsDepleted(t *testing.T) {
	for _, tc := range []struct {
		name       string
		rx, ry     int64 // reserve balance
		ps         int64 // pool coin supply
		isDepleted bool
	}{
		{
			name:       "empty pool",
			rx:         0,
			ry:         0,
			ps:         0,
			isDepleted: true,
		},
		{
			name:       "depleted, with some coins from outside",
			rx:         100,
			ry:         0,
			ps:         0,
			isDepleted: true,
		},
		{
			name:       "depleted, with some coins from outside #2",
			rx:         100,
			ry:         100,
			ps:         0,
			isDepleted: true,
		},
		{
			name:       "normal pool",
			rx:         10000,
			ry:         10000,
			ps:         10000,
			isDepleted: false,
		},
		{
			name:       "not depleted, but reserve coins are gone",
			rx:         0,
			ry:         10000,
			ps:         10000,
			isDepleted: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool := amm.CreatePool(sdk.NewInt(tc.rx), sdk.NewInt(tc.ry), sdk.NewInt(tc.ps))
			require.Equal(t, tc.isDepleted, pool.IsDepleted())
		})
	}
}

func TestDeposit(t *testing.T) {
	for _, tc := range []struct {
		name   string
		rx, ry int64 // reserve balance
		ps     int64 // pool coin supply
		x, y   int64 // depositing coin amount
		ax, ay int64 // expected accepted coin amount
		pc     int64 // expected minted pool coin amount
	}{
		{
			name: "ideal deposit",
			rx:   2000,
			ry:   100,
			ps:   10000,
			x:    200,
			y:    10,
			ax:   200,
			ay:   10,
			pc:   1000,
		},
		{
			name: "unbalanced deposit",
			rx:   2000,
			ry:   100,
			ps:   10000,
			x:    100,
			y:    2000,
			ax:   100,
			ay:   5,
			pc:   500,
		},
		{
			name: "decimal truncation",
			rx:   222,
			ry:   333,
			ps:   333,
			x:    100,
			y:    100,
			ax:   66,
			ay:   99,
			pc:   99,
		},
		{
			name: "decimal truncation #2",
			rx:   200,
			ry:   300,
			ps:   333,
			x:    80,
			y:    80,
			ax:   53,
			ay:   80,
			pc:   88,
		},
		{
			name: "zero minting amount",
			ps:   100,
			rx:   10000,
			ry:   10000,
			x:    99,
			y:    99,
			ax:   0,
			ay:   0,
			pc:   0,
		},
		{
			name: "tiny minting amount",
			rx:   10000,
			ry:   10000,
			ps:   100,
			x:    100,
			y:    100,
			ax:   100,
			ay:   100,
			pc:   1,
		},
		{
			name: "tiny minting amount #2",
			rx:   10000,
			ry:   10000,
			ps:   100,
			x:    199,
			y:    199,
			ax:   100,
			ay:   100,
			pc:   1,
		},
		{
			name: "zero minting amount",
			rx:   10000,
			ry:   10000,
			ps:   999,
			x:    10,
			y:    10,
			ax:   0,
			ay:   0,
			pc:   0,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool := amm.CreatePool(sdk.NewInt(tc.rx), sdk.NewInt(tc.ry), sdk.NewInt(tc.ps))
			ax, ay, pc := pool.Deposit(sdk.NewInt(tc.x), sdk.NewInt(tc.y))
			require.True(sdk.IntEq(t, sdk.NewInt(tc.ax), ax))
			require.True(sdk.IntEq(t, sdk.NewInt(tc.ay), ay))
			require.True(sdk.IntEq(t, sdk.NewInt(tc.pc), pc))
			// Additional assertions
			if !pool.IsDepleted() {
				require.True(t, (ax.Int64()*tc.ps) >= (pc.Int64()*tc.rx)) // (ax / Rx) > (pc / Ps)
				require.True(t, (ay.Int64()*tc.ps) >= (pc.Int64()*tc.ry)) // (ay / Ry) > (pc / Ps)
			}
		})
	}
}

func TestWithdraw(t *testing.T) {
	for _, tc := range []struct {
		name    string
		rx, ry  int64 // reserve balance
		ps      int64 // pool coin supply
		pc      int64 // redeeming pool coin amount
		feeRate sdk.Dec
		x, y    int64 // withdrawn coin amount
	}{
		{
			name:    "ideal withdraw",
			rx:      2000,
			ry:      100,
			ps:      10000,
			pc:      1000,
			feeRate: sdk.ZeroDec(),
			x:       200,
			y:       10,
		},
		{
			name:    "ideal withdraw - with fee",
			rx:      2000,
			ry:      100,
			ps:      10000,
			pc:      1000,
			feeRate: sdk.MustNewDecFromStr("0.003"),
			x:       199,
			y:       9,
		},
		{
			name:    "withdraw all",
			rx:      123,
			ry:      567,
			ps:      10,
			pc:      10,
			feeRate: sdk.MustNewDecFromStr("0.003"),
			x:       123,
			y:       567,
		},
		{
			name:    "advantageous for pool",
			rx:      100,
			ry:      100,
			ps:      10000,
			pc:      99,
			feeRate: sdk.ZeroDec(),
			x:       0,
			y:       0,
		},
		{
			name:    "advantageous for pool",
			rx:      10000,
			ry:      100,
			ps:      10000,
			pc:      99,
			feeRate: sdk.ZeroDec(),
			x:       99,
			y:       0,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool := amm.CreatePool(sdk.NewInt(tc.rx), sdk.NewInt(tc.ry), sdk.NewInt(tc.ps))
			x, y := pool.Withdraw(sdk.NewInt(tc.pc), tc.feeRate)
			require.True(sdk.IntEq(t, sdk.NewInt(tc.x), x))
			require.True(sdk.IntEq(t, sdk.NewInt(tc.y), y))
			// Additional assertions
			require.True(t, (tc.pc*tc.rx) >= (x.Int64()*tc.ps))
			require.True(t, (tc.pc*tc.ry) >= (y.Int64()*tc.ps))
		})
	}
}

func TestXtoY(t *testing.T) {
	for _, tc := range []struct {
		name            string
		rx, ry          int64 // reserve balance
		inputX, outputY int64 // withdrawn coin amount
	}{
		{
			name:    "x to y swap 1",
			rx:      100000,
			ry:      10000000000,
			inputX:  100000,
			outputY: 5000000000,
		},
		{
			name:    "x to y swap 2",
			rx:      2000000,
			ry:      3000000,
			inputX:  500000,
			outputY: 600000,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool := amm.CreatePool(sdk.NewInt(tc.rx), sdk.NewInt(tc.ry), sdk.NewInt(1000000))
			beforeK := pool.K()
			y := pool.XtoY(sdk.NewInt(tc.inputX))
			require.True(sdk.IntEq(t, sdk.NewInt(tc.outputY), y))
			require.True(t, pool.K().GTE(beforeK))
		})
	}
}

func TestYtoX(t *testing.T) {
	for _, tc := range []struct {
		name            string
		rx, ry          int64 // reserve balance
		inputY, outputX int64 //
	}{
		{
			name:    "y to x swap 1",
			rx:      2000000,
			ry:      3000000,
			inputY:  1000000,
			outputX: 500000,
		},
		{
			name:    "y to x swap 2",
			rx:      3000000,
			ry:      2000000,
			inputY:  100,
			outputX: 149,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool := amm.CreatePool(sdk.NewInt(tc.rx), sdk.NewInt(tc.ry), sdk.NewInt(1000000))
			beforeK := pool.K()
			x := pool.YtoX(sdk.NewInt(tc.inputY))
			require.True(sdk.IntEq(t, sdk.NewInt(tc.outputX), x))
			require.True(t, pool.K().GTE(beforeK))
		})
	}
}

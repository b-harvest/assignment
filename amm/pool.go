package amm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Pool is the basic pool type.
type Pool struct {
	// Rx and Ry are the pool's reserve balance of each x/y coin.
	Rx, Ry sdk.Int
	// Ps is the pool's pool coin supply.
	Ps sdk.Int
}

// CreatePool returns a new Pool.
// It is OK to pass an empty sdk.Int to Ps when Ps is not going to be used.
func CreatePool(rx, ry, ps sdk.Int) *Pool {
	return &Pool{
		Rx: rx,
		Ry: ry,
		Ps: ps,
	}
}

// Balances returns the balances of the pool.
func (pool *Pool) Balances() (rx, ry sdk.Int) {
	return pool.Rx, pool.Ry
}

// PoolCoinSupply returns the pool coin supply.
func (pool *Pool) PoolCoinSupply() sdk.Int {
	return pool.Ps
}

// Price returns the pool price.
func (pool *Pool) Price() sdk.Dec {
	if pool.Rx.IsZero() || pool.Ry.IsZero() {
		panic("pool price is not defined for a depleted pool")
	}
	return pool.Rx.ToDec().Quo(pool.Ry.ToDec())
}

// K returns the pool k.
func (pool *Pool) K() sdk.Int {
	return pool.Rx.Mul(pool.Ry)
}

// IsDepleted returns whether the pool is depleted or not.
func (pool *Pool) IsDepleted() bool {
	return pool.Ps.IsZero() || pool.Rx.IsZero() || pool.Ry.IsZero()
}

// Deposit returns accepted x and y coin amount and minted pool coin amount
// when someone deposits x and y coins.
func (pool *Pool) Deposit(x, y sdk.Int) (ax, ay, pc sdk.Int) {
	// Calculate accepted amount and minting amount.
	// Note that we take as many coins as possible(by ceiling numbers)
	// from depositor and mint as little coins as possible.

	// TODO: implement calculating logic for ax, ay, pc =================
	// ..
	// ==================================================================

	// update pool states
	pool.Rx = pool.Rx.Add(ax)
	pool.Ry = pool.Ry.Add(ay)
	pool.Ps = pool.Ps.Add(pc)
	return
}

// Withdraw returns withdrawn x and y coin amount when someone withdraws
// pc pool coin.
// Withdraw also takes care of the fee rate.
func (pool *Pool) Withdraw(pc sdk.Int, feeRate sdk.Dec) (x, y sdk.Int) {
	// TODO: implement calculating logic for x, y =======================
	// ..
	// ==================================================================

	// update pool states
	pool.Rx = pool.Rx.Sub(x)
	pool.Ry = pool.Ry.Sub(y)
	pool.Ps = pool.Ps.Sub(pc)
	return
}

func (pool *Pool) XtoY(xDelta sdk.Int) (yDelta sdk.Int) {
	// TODO: implement x to y swap logic  ===============================
	// ..
	// ==================================================================

	// update pool states
	pool.Rx = pool.Rx.Add(xDelta)
	pool.Ry = pool.Ry.Sub(yDelta)
	return
}

func (pool *Pool) YtoX(yDelta sdk.Int) (xDelta sdk.Int) {
	// TODO: implement y to x swap logic  ===============================
	// ..
	// ==================================================================

	// update pool states
	pool.Rx = pool.Rx.Sub(xDelta)
	pool.Ry = pool.Ry.Add(yDelta)
	return
}

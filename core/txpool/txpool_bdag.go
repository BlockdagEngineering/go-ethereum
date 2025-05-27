package txpool

func (txPool *TxPool) Subpools() []SubPool {
	return txPool.subpools
}
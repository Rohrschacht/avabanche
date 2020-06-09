package main

type transactionFactory struct {
	amount uint
	to     string
	user   user
}

func (tf *transactionFactory) newRegularTransaction(rpcID uint) transaction {
	tx := new(regularTransaction)
	tx.amount = tf.amount
	tx.to = tf.to
	tx.user = tf.user
	tx.rpcID = rpcID
	return tx
}

func (tf *transactionFactory) newSensorTransaction(rpcID uint) transaction {
	tx := new(sensorTransaction)
	tx.amount = tf.amount
	tx.to = tf.to
	tx.user = tf.user
	tx.rpcID = rpcID
	return tx
}

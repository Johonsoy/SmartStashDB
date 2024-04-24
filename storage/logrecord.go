package storage

import (
	"encoding/binary"
)

type LogRecordType = byte

const MaxLogRecordLength = 1 + binary.MaxVarintLen64*2

type LogRecord struct {
	Key     []byte
	Value   []byte
	Type    LogRecordType
	BatchId uint64
}

func NewLogRecord(key []byte, value []byte, batchId uint64) *LogRecord {
	return &LogRecord{}
}

// Encode Serialize LogRecord, header + batchId + keySize + valueSize + key + value /*
func (logRecord *LogRecord) Encode() []byte {
	header := make([]byte, MaxLogRecordLength)
	header[0] = logRecord.Type

	index := 1

	index += binary.PutUvarint(header[index:], logRecord.BatchId)
	index += binary.PutVarint(header[index:], int64(len(logRecord.Key)))
	index += binary.PutVarint(header[index:], int64(len(logRecord.Value)))

	value := make([]byte, index+len(logRecord.Key)+len(logRecord.Value))

	// copy header.
	copy(value, header[:index])

	copy(value[index:], logRecord.Key)

	copy(value[index+len(logRecord.Key):], logRecord.Value)

	return value
}

func (logRecord *LogRecord) Decode(b []byte) {
	logRecord.Type = b[0]

	index := 1
	n := 0
	logRecord.BatchId, n = binary.Uvarint(b[index:])
	index += n
	keyLength, n := binary.Uvarint(b[index:])
	index += n

	valueLength, n := binary.Uvarint(b[index:])
	index += n

	key := make([]byte, keyLength)

	value := make([]byte, valueLength)

	copy(key, b[index:index+int(keyLength)])
	index += int(keyLength)

	copy(value, b[index:index+int(valueLength)])

	logRecord.Key = key
	logRecord.Value = value

}

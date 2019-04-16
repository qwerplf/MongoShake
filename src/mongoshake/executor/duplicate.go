package executor

import (
	"mongoshake/common"
	"mongoshake/collector/configure"
	"mongoshake/oplog"

	"github.com/vinllen/mgo"
	"github.com/vinllen/mgo/bson"
	LOG "github.com/vinllen/log4go"
)

func HandleDuplicated(collection *mgo.Collection, records []*OplogRecord, op int8) {
	for _, record := range records {
		log := record.original.partialLog
		switch conf.Options.ReplayerConflictWriteTo {
		case DumpConflictToDB:
			// general process : write record to specific database
			session := collection.Database.Session
			// discard conflict again
			session.DB(utils.APPConflictDatabase).C(collection.Name).Insert(log.Object)
		case DumpConflictToSDK, NoDumpConflict:
		}

		if utils.SentinelOptions.DuplicatedDump {
			SnapshotDiffer{op: op, log: log}.dump(collection)
		}
	}
}

type SnapshotDiffer struct {
	op        int8
	log       *oplog.PartialLog
	foundInDB bson.M
}

func (s SnapshotDiffer) write2Log() {
	LOG.Info("Found in DB ==> %v", s.foundInDB)
	LOG.Info("Oplog ==> %v", s.log.Object)
}

func (s SnapshotDiffer) dump(coll *mgo.Collection) {
	if s.op == OpUpdate {
		coll.Find(s.log.Query).One(s.foundInDB)
	} else {
		coll.Find(bson.M{"_id": s.log.Object["_id"]}).One(s.foundInDB)
	}
	s.write2Log()
}

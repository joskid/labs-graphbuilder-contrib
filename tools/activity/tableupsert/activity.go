/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package tableupsert

import (
	"encoding/json"
	"reflect"
	"sync"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/TIBCOSoftware/labs-graphbuilder-lib/table"
	"github.com/TIBCOSoftware/labs-graphbuilder-lib/util"
)

// activityLogger is the default logger for the Filter Activity
var log = logger.GetLogger("activity-table")

const (
	setting_Table = "Table"
	input         = "Mapping"
	output_Data   = "Data"
	output_Exists = "Exists"
)

// TableUpsertActivity is an Activity that is used to Filter a message to the console
type TableUpsertActivity struct {
	metadata        *activity.Metadata
	activityToTable map[string]string
	mux             sync.Mutex
}

// NewActivity creates a new AppActivity
func NewActivity(metadata *activity.Metadata) activity.Activity {
	aTableActivity := &TableUpsertActivity{
		metadata:        metadata,
		activityToTable: make(map[string]string),
	}
	return aTableActivity
}

// Metadata returns the activity's metadata
func (a *TableUpsertActivity) Metadata() *activity.Metadata {
	return a.metadata
}

// Eval implements api.Activity.Eval - Filters the Message
func (a *TableUpsertActivity) Eval(ctx activity.Context) (done bool, err error) {

	myTable, err := a.getTable(ctx)

	if nil != err {
		return false, err
	}

	iData := ctx.GetInput(input).(*data.ComplexObject).Value

	log.Info("iData.Value = ", iData, ", type = ", reflect.TypeOf(iData))

	outputTuple, exists := myTable.Upsert(iData.(map[string]interface{}))

	log.Info("output tuple = ", outputTuple, ", exist = ", exists, ", row count = ", myTable.RowCount())

	complexdata := &data.ComplexObject{Metadata: "Data", Value: outputTuple}
	ctx.SetOutput(output_Data, complexdata)
	ctx.SetOutput(output_Exists, exists)

	return true, nil
}

func (a *TableUpsertActivity) getTable(context activity.Context) (*table.Table, error) {
	myId := util.ActivityId(context)

	myTable := table.GetTableManager().GetTable(a.activityToTable[myId])
	if nil == myTable {
		a.mux.Lock()
		defer a.mux.Unlock()

		myTable = table.GetTableManager().GetTable(a.activityToTable[myId])
		if nil == myTable {

			log.Info("[TableUpsertActivity] init : ", "initialize table begin ....")

			iTableInfo, exist := context.GetSetting(setting_Table)
			if !exist {
				return nil, activity.NewError("Table is not configured", "TABLE_UPSERT-4002", nil)
			}

			//Read table details
			tableInfo, _ := data.CoerceToObject(iTableInfo)
			if tableInfo == nil {
				return nil, activity.NewError("Unable extract table details", "TABLE_UPSERT-4001", nil)
			}

			var tablename string
			var schema []interface{}
			tableSettings, _ := tableInfo["settings"].([]interface{})
			if tableSettings != nil {
				for _, v := range tableSettings {
					setting, _ := data.CoerceToObject(v)

					if nil != setting {
						if setting["name"] == "schema" {
							iSchema := setting["value"]
							if nil == iSchema {
								return nil, activity.NewError("Unable to get model string", "TABLE_UPSERT-4004", nil)
							}
							err := json.Unmarshal([]byte(iSchema.(string)), &schema)
							if nil != err {
								return nil, err
							}
						} else if setting["name"] == "name" {
							tablename = setting["value"].(string)
						}
					}
				}
			}

			if "" == tablename {
				return nil, activity.NewError("Unable to get table name", "TABLE_UPSERT-4003", nil)
			}

			log.Info("-============= TABLE SCHEMA ================-")
			log.Info(schema)
			log.Info("-===========================================-")

			keyName := make([]string, 0)
			tableSchema := make([](map[string]interface{}), len(schema))
			for index, field := range schema {
				tableSchema[index] = field.(map[string]interface{})
				if "yes" == tableSchema[index]["IsKey"].(string) {
					keyName = append(keyName, tableSchema[index]["Name"].(string))
				}
			}

			myTable = table.GetTableManager().CreateTable(
				tablename,
				keyName,
				tableSchema,
			)
			log.Info("[TableUpsertActivity] init : ", "initialize table done : myTable = ", myTable)
			a.activityToTable[myId] = tablename
		}
	}

	return myTable, nil
}

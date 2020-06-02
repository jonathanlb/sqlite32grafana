package sqlite3

// GetTagKeys returns the column names and and underlying types available
// to label timeseries observations.
func (tsm *sqliteTimeSeriesManager) GetTagKeys(target string, dest *[]TagKey) error {
	valueColumn, keyColumns := tsm.target2tokens(target)
	if err := tsm.getSchema(tsm.table, dest); err != nil {
		return err
	}
	// remove the specified columns from the result
	for _, col := range append(keyColumns, tsm.timeColumn, valueColumn) {
		for idx, i := range *dest {
			if col == i.Text {
				n1 := len(*dest) - 1
				(*dest)[idx] = (*dest)[n1]
				*dest = (*dest)[0:n1]
				break
			}
		}
	}
	// rename for grafana
	deref := *dest
	for i, x := range deref {
		deref[i].Type = sql2grafanaType(x.Type)
	}
	return nil
}

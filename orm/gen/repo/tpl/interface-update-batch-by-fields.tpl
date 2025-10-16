// UpdateBatchBy{{.upperFields}} 根据字段{{.upperFields}}批量更新,零值会被更新
UpdateBatchBy{{.upperFields}}(ctx context.Context,{{.fieldAndDataTypes}}, data map[string]interface{}) error
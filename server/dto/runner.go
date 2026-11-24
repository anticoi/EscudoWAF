package dto

// RunnerControlRequest petición de control del runner
type RunnerControlRequest struct {
	Action string `json:"action" binding:"required,oneof=start stop restart force_stop reload"` // Acción de control
}

// RunnerControlResponse respuesta de control del runner
type RunnerControlResponse struct {
	Success bool   `json:"success" example:"true"`     // Si la operación tuvo éxito
	Action  string `json:"action" example:"start"`     // Acción ejecutada
	Message string `json:"message" example:"El runner se inició con éxito"` // Mensaje de la operación
	State   string `json:"state" example:"running"`    // Estado después de la operación
}

// RunnerStatusResponse respuesta del estado del runner
type RunnerStatusResponse struct {
	State     string `json:"state" example:"running"`  // Estado: running, stopped, error
	IsRunning bool   `json:"isRunning" example:"true"` // Si está en ejecución
}

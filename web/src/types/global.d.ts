declare global {
    // Añadir el tipo ApiError a window
    // interface Window {
    //     ApiError: typeof import('../api').ApiError
    // }

    // Declarar tipos globales directamente
    type ApiError = import('../api').ApiError
}

// Esta exportación es necesaria para que este archivo se trate como un módulo
export { }

import { useState } from 'react'
import { useNavigate } from 'react-router' // Cambiar la ruta de importación de react-router-dom a react-router
import { useQuery, useMutation } from '@tanstack/react-query'
import { authApi } from '@/api/services'
import useAuthStore from '@/store/auth'
import { LoginFormValues, PasswordResetFormValues } from '@/validation/auth'
import { useEffect } from 'react'
import { GetUserInfoResponseData } from '@/types/auth'

export const useLogin = () => {
    const navigate = useNavigate()
    const { login } = useAuthStore()
    const [error, setError] = useState<string | null>(null)

    const mutation = useMutation({
        mutationFn: (values: LoginFormValues) => authApi.login(values),
        onSuccess: (data) => {
            // Store token and user data
            login(data.token, data.user)

            // Redirect based on password reset requirement
            if (data.user.needReset) {
                navigate('/reset-password')
            } else {
                navigate('/')
            }
        },
        onError: (error: ApiError) => {
            setError(error.message || 'Error de inicio de sesión, compruebe el usuario y la contraseña')
        }
    })

    return {
        login: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}

export const useResetPassword = () => {
    const navigate = useNavigate()
    const [error, setError] = useState<string | null>(null)
    const { user, logout } = useAuthStore()

    const mutation = useMutation({
        mutationFn: (values: Omit<PasswordResetFormValues, 'confirmPassword'>) => {
            return authApi.resetPassword({
                oldPassword: values.oldPassword,
                newPassword: values.newPassword,
            })
        },
        onSuccess: () => {
            // If the user was required to reset password, log them out
            // to force a new login with the new password
            if (user?.needReset) {
                logout()
                navigate('/login', { state: { message: 'Contraseña restablecida, inicie sesión con la nueva contraseña' } })
            } else {
                navigate('/')
            }
        },
        onError: (error: ApiError) => {
            setError(error.message || 'Error al restablecer la contraseña, inténtelo de nuevo')
        },
    })

    return {
        resetPassword: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}


export const useCurrentUser = () => {
    const { user, setUser } = useAuthStore()

    const { data, isLoading, error } = useQuery<GetUserInfoResponseData, ApiError>({
        queryKey: ['currentUser'],
        queryFn: () => authApi.getCurrentUser(),
        enabled: !!useAuthStore.getState().token,
    })

    // Escuchar cambios de datos, sustituye a onSuccess
    useEffect(() => {
        if (data) {
            setUser(data)
        }
    }, [data, setUser])

    return {
        user: data || user,
        isLoading,
        error,
    }
}
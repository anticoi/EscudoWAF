import { z } from 'zod';

// Login form validation schema
export const loginSchema = z.object({
  username: z
    .string()
    .min(3, { message: 'El nombre de usuario debe tener al menos 3 caracteres' })
    .max(20, { message: 'El nombre de usuario puede tener máximo 20 caracteres' }),
  password: z
    .string()
    .min(6, { message: 'La contraseña debe tener al menos 6 caracteres' }),
});

// Password reset validation schema
export const passwordResetSchema = z
  .object({
    oldPassword: z.string().min(1, { message: 'Por favor, introduzca la contraseña actual' }),
    newPassword: z.string().min(6, { message: 'La nueva contraseña debe tener al menos 6 caracteres' }),
    confirmPassword: z.string().min(6, { message: 'Por favor, confirme la nueva contraseña' }),
  })
  .refine((data) => data.newPassword === data.confirmPassword, {
    message: 'Las contraseñas introducidas no coinciden',
    path: ['confirmPassword'],
  });

// Types based on the schemas
export type LoginFormValues = z.infer<typeof loginSchema>;
export type PasswordResetFormValues = z.infer<typeof passwordResetSchema>; 
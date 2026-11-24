import * as z from 'zod';

export const certificateFormSchema = z.object({
  name: z.string().min(1, 'El nombre del certificado no puede estar vacío'),
  description: z.string().optional(),
  publicKey: z.string().min(1, 'La clave pública no puede estar vacía'),
  privateKey: z.string().min(1, 'La clave privada no puede estar vacía'),
});

export const certificateUpdateSchema = z.object({
  name: z.string().optional(),
  description: z.string().optional(),
  publicKey: z.string().optional(),
  privateKey: z.string().optional(),
}); 
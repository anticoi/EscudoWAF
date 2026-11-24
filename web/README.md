# EscudoWAF Frontend

Interfaz web moderna para el sistema de firewall de aplicaciones web (WAF) EscudoWAF, construida con React, TypeScript y Vite.

## Funcionalidades

- **Sistema de autenticación**: inicio de sesión seguro y restablecimiento de contraseña
- **Monitorización en panel**: seguimiento en tiempo real de amenazas de seguridad web
- **Análisis de registros**: registros detallados de detección de ataques y eventos de protección
- **Gestión de reglas**: configuración y administración de reglas de seguridad del WAF
- **Gestión de sitios**: administración de sitios web y dominios protegidos
- **Gestión de certificados**: administración de certificados SSL/TLS
- **Soporte multi-idioma**: internacionalización en español, inglés y chino

## Stack tecnológico

- **Núcleo**: React + TypeScript + Vite
- **Estilos**: TailwindCSS + Shadcn UI
- **Enrutamiento**: React Router v7
- **Gestión de estado**: Zustand
- **Formularios**: React Hook Form + validación Zod
- **Obtención de datos**: TanStack Query (React Query v4)
- **Tablas**: TanStack Table
- **Internacionalización**: react-i18next con i18next-http-backend y i18next-browser-languagedetector
- **Iconos**: Lucide React

## Estructura del proyecto

```
├── public/                # Recursos estáticos
│   └── locales/           # Archivos de internacionalización
│       ├── es/            # Traducciones al español
│       ├── en/            # Traducciones al inglés
│       └── zh/            # Traducciones al chino
├── src/
│   ├── api/               # Servicios de API y manejo de peticiones
│   ├── assets/            # Recursos estáticos del proyecto
│   ├── components/        # Componentes de UI
│   │   ├── common/        # Componentes comunes compartidos
│   │   ├── layout/        # Componentes de layout
│   │   ├── table/         # Componentes de tabla (wrappers de TanStack Table)
│   │   └── ui/            # Componentes de Shadcn UI
│   ├── feature/           # Módulos de funcionalidad
│   │   └── auth/          # Componentes y hooks de autenticación
│   ├── hooks/             # Hooks personalizados de React
│   ├── lib/               # Librerías de utilidades
│   ├── pages/             # Páginas de la aplicación
│   │   ├── auth/          # Páginas de autenticación
│   │   ├── logs/          # Páginas de análisis de registros
│   │   ├── monitor/       # Páginas del panel de monitorización
│   │   ├── rule/          # Páginas de gestión de reglas
│   │   └── setting/       # Páginas de configuración
│   ├── routes/            # Configuración de enrutamiento
│   ├── store/             # Gestión de estado con Zustand
│   ├── types/             # Definiciones de tipos TypeScript
│   ├── utils/             # Funciones de utilidad
│   ├── validation/        # Esquemas de validación de formularios (Zod)
│   ├── App.tsx            # Componente principal de la app
│   ├── i18n.ts            # Configuración de i18n
│   ├── index.css          # Estilos globales
│   └── main.tsx           # Punto de entrada de la aplicación
├── .eslintrc.js           # Configuración de ESLint
├── components.json        # Configuración de componentes Shadcn
├── package.json           # Dependencias y scripts
├── tailwind.config.js     # Configuración de Tailwind CSS
├── tsconfig.json          # Configuración de TypeScript
└── vite.config.ts         # Configuración de Vite
```

## Primeros pasos

### Requisitos previos

- Node.js 22+
- pnpm (gestor de paquetes recomendado)

### Instalación

1. Clonar el repositorio:
   ```bash
   git clone https://github.com/anticoi/EscudoWAF.git
   cd web
   ```

2. Instalar dependencias:
   ```bash
   pnpm install
   ```

3. Iniciar el servidor de desarrollo:
   ```bash
   pnpm dev
   ```

4. Compilar para producción:
   ```bash
   pnpm build
   ```

## Funcionalidades clave

### Estructura del layout
- Barra lateral izquierda para la navegación principal
- Área de contenido derecha con migas de pan para la subnavegación
- Rutas de migas de pan según la navegación seleccionada en la barra lateral

### Autenticación
- Inicio de sesión con usuario/contraseña
- Restablecimiento de contraseña obligatorio para nuevos usuarios
- Rutas protegidas que requieren autenticación

### Internacionalización
- Soporte multi-idioma (español, inglés y chino)
- Detección automática del idioma
- Selector de idioma

## Contribuir

1. Seguir la estructura del código y las convenciones de nomenclatura del proyecto
2. Asegurar que el código esté correctamente formateado y comentado
3. Escribir tests para las nuevas funcionalidades
4. Crear un pull request con una descripción clara de los cambios

## Licencia

Copyright © 2025 huahua (RuiQi) / anticoi (EscudoWAF). Licencia MIT.

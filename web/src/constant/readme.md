```
// Ejemplo de uso
import {
  getConstant,
  setConstant,
  hasConstant,
  getCategoryConstants,
  setBatchConstants,
  ConstantCategory
} from '@/constant';

// Obtener constante
const apiTimeout = getConstant(ConstantCategory.SYSTEM, 'API_TIMEOUT', 5000);
const pageSize = getConstant(ConstantCategory.CONFIG, 'DEFAULT_PAGE_SIZE', 10);

// Establecer constante
setConstant(ConstantCategory.UI, 'SIDEBAR_WIDTH', 250);

// Comprobar si la constante existe
if (hasConstant(ConstantCategory.FEATURE, 'WAF_MODES')) {
  // Usar constantes...
}

// Obtener todas las constantes de una categoría
const allUiConstants = getCategoryConstants(ConstantCategory.UI);

// Establecer constantes por lotes
setBatchConstants(ConstantCategory.CONFIG, {
  'MAX_FILE_SIZE': 10485760, // 10MB
  'ALLOWED_FILE_TYPES': 'jpg,png,pdf',
  'MAX_UPLOAD_FILES': 5
});
```

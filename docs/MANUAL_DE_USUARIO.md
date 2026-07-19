# Manual de Usuario — FINEXIA

**Versión del documento:** 1.0
**Fecha:** Julio 2026
**Aplicación:** Finexia — Plataforma de gestión y seguimiento de portafolios de inversión

---

## Tabla de contenido

1. [Introducción](#1-introducción)
2. [Requisitos y acceso](#2-requisitos-y-acceso)
3. [Primeros pasos: registro e inicio de sesión](#3-primeros-pasos-registro-e-inicio-de-sesión)
4. [Interfaz general de la aplicación](#4-interfaz-general-de-la-aplicación)
5. [Dashboard (panel principal)](#5-dashboard-panel-principal)
6. [Portafolios](#6-portafolios)
7. [Posiciones y activos](#7-posiciones-y-activos)
8. [Inversiones](#8-inversiones)
9. [Plataformas](#9-plataformas)
10. [Transacciones](#10-transacciones)
11. [Importación masiva de transacciones (Excel/CSV)](#11-importación-masiva-de-transacciones-excelcsv)
12. [Reportes y exportaciones](#12-reportes-y-exportaciones)
13. [Notificaciones](#13-notificaciones)
14. [Configuración de la cuenta](#14-configuración-de-la-cuenta)
15. [Seguridad: 2FA y sesiones](#15-seguridad-2fa-y-sesiones)
16. [Panel de administración](#16-panel-de-administración)
17. [Preguntas frecuentes (FAQ)](#17-preguntas-frecuentes-faq)
18. [Solución de problemas](#18-solución-de-problemas)
19. [Glosario](#19-glosario)

---

## 1. Introducción

### 1.1 ¿Qué es Finexia?

Finexia es una aplicación web para **gestionar y hacer seguimiento de tus portafolios de inversión** en un solo lugar. Te permite registrar tus activos (acciones, criptomonedas, fondos y otros instrumentos), las plataformas donde los tienes, y todas tus transacciones de compra y venta, para luego visualizar el rendimiento, la distribución y el crecimiento de tu patrimonio con gráficas y reportes descargables.

### 1.2 Principios clave

- **Tú controlas tus datos.** Finexia **no se conecta a tus brokers ni a tus plataformas de inversión**, y nunca te pedirá las credenciales de esas cuentas. Toda la información se registra manualmente (o mediante importación de archivos Excel/CSV), por lo que siempre está bajo tu control.
- **Multi-portafolio.** Puedes crear tantos portafolios como necesites (por ejemplo: "Retiro", "Cripto", "Fondo de emergencia") y analizarlos de forma individual o agregada.
- **Multi-moneda.** Cada portafolio tiene una moneda base y la aplicación utiliza tasas de cambio para consolidar valores entre monedas.
- **Seguridad primero.** La aplicación incluye verificación de correo electrónico, autenticación en dos pasos (2FA), gestión de sesiones activas y protección contra intentos de acceso masivos.

### 1.3 ¿A quién va dirigido este manual?

Este manual está dirigido a:

- **Usuarios finales** que utilizan Finexia para el seguimiento de sus inversiones personales (secciones 3 a 15).
- **Administradores** de la plataforma, que además gestionan usuarios, el catálogo de activos y las tasas de cambio (sección 16).

---

## 2. Requisitos y acceso

### 2.1 Requisitos

- Un navegador web moderno y actualizado (Chrome, Firefox, Safari o Edge).
- Conexión a internet.
- Una dirección de correo electrónico válida (necesaria para verificar la cuenta y recuperar la contraseña).
- Opcional, pero recomendado: una aplicación de autenticación (Google Authenticator, Authy, 1Password, etc.) para activar la verificación en dos pasos.

La aplicación es **responsive**: funciona tanto en computadoras de escritorio como en tabletas y teléfonos móviles. En pantallas pequeñas el menú lateral se oculta y se abre con el botón de menú de la cabecera.

### 2.2 Página de inicio (landing) y lista de espera

Al visitar la dirección pública de Finexia sin haber iniciado sesión verás la página de presentación, con la descripción del producto, preguntas frecuentes y las páginas legales (Términos, Privacidad y Cookies).

Si el registro directo no está habilitado, la página ofrece unirte a la **lista de espera (waitlist)** dejando tu correo electrónico. Un administrador podrá luego enviarte una **invitación** para crear tu cuenta.

---

## 3. Primeros pasos: registro e inicio de sesión

### 3.1 Crear una cuenta

Existen dos vías para obtener una cuenta:

**A) Registro directo** (si está habilitado):

1. En la página de inicio, pulsa **Registrarse / Crear cuenta**.
2. Completa tus datos: nombre, correo electrónico y contraseña.
3. Envía el formulario. Recibirás un **correo de verificación**.
4. Abre el enlace del correo para **verificar tu dirección**. Sin este paso no podrás iniciar sesión (el sistema mostrará un aviso de "correo sin verificar").

> **Nota:** si el registro directo está desactivado, el formulario mostrará un aviso indicando que el registro está deshabilitado. En ese caso, únete a la lista de espera o solicita una invitación a un administrador.

**B) Por invitación:**

1. Un administrador te envía una invitación a tu correo.
2. Abre el enlace **Aceptar invitación** del correo. La aplicación validará el token de la invitación.
3. Define tu contraseña y confirma. Tu cuenta quedará creada y verificada, lista para iniciar sesión.

Las invitaciones tienen caducidad; si el enlace expiró, pide al administrador que la **reenvíe**.

### 3.2 Iniciar sesión

1. Ve a la página de **Iniciar sesión**.
2. Introduce tu correo y contraseña.
3. Si tienes **2FA activado**, la aplicación te pedirá un segundo paso: introduce el **código de 6 dígitos** de tu aplicación de autenticación (o uno de tus códigos de recuperación).
4. Al completar el acceso entrarás directamente al **Dashboard**.

La sesión se mantiene mediante un token de acceso de corta duración y una cookie segura de renovación (*refresh token*) que se rota automáticamente; no necesitas hacer nada para mantener la sesión activa mientras uses la aplicación.

### 3.3 ¿Olvidaste tu contraseña?

1. En la pantalla de inicio de sesión, pulsa **¿Olvidaste tu contraseña?**
2. Introduce tu correo electrónico. Si existe una cuenta asociada, recibirás un enlace de restablecimiento.
3. Abre el enlace, define tu **nueva contraseña** y confírmala.
4. Inicia sesión con la nueva contraseña.

Los enlaces de restablecimiento caducan por seguridad; si el tuyo expiró, solicita uno nuevo.

### 3.4 Verificación de correo

Si intentas iniciar sesión sin haber verificado tu correo, la aplicación te lo indicará y te permitirá **reenviar el correo de verificación**. Revisa también la carpeta de spam.

### 3.5 Cerrar sesión

Desde el menú de usuario (cabecera) o desde **Configuración → Sesiones activas** puedes cerrar la sesión actual. También puedes cerrar sesiones abiertas en otros dispositivos (ver sección 15.3).

---

## 4. Interfaz general de la aplicación

Una vez dentro, la aplicación se organiza en tres zonas:

### 4.1 Cabecera (header)

Situada en la parte superior. Contiene:

- El botón de **menú** (en pantallas pequeñas) para mostrar/ocultar la barra lateral.
- El acceso al **perfil de usuario** (avatar) y el cierre de sesión.

### 4.2 Barra lateral (sidebar)

Es el menú principal de navegación. Sus secciones son:

| Sección | ¿Para qué sirve? |
|---|---|
| **Dashboard** | Vista general: resumen financiero y crecimiento del patrimonio |
| **Portafolios** | Crear y gestionar tus portafolios y sus posiciones |
| **Inversiones** | Catálogo de productos y oportunidades de inversión |
| **Plataformas** | Registrar los brokers/exchanges/bancos donde tienes activos |
| **Transacciones** | Historial de operaciones e importación masiva |
| **Reportes** | Estadísticas, calendario de rendimiento y descargas en Excel |
| **Notificaciones** | Preferencias de avisos por correo y en la app |
| **Configuración** | Perfil, apariencia, contraseña, 2FA y sesiones |

Si tu cuenta tiene rol de **administrador**, verás además el grupo **Panel Admin** con: **Usuarios**, **Activos** y **Tasas de Cambio** (ver sección 16).

### 4.3 Área de contenido

Es la zona central donde se muestra cada página. En la mayoría de listados encontrarás **paginación** (selector de página y de cantidad de elementos por página) y acciones contextuales (crear, editar, eliminar).

---

## 5. Dashboard (panel principal)

El Dashboard es la primera pantalla tras iniciar sesión y ofrece una fotografía general de tus finanzas:

- **Resumen financiero:** el valor total de tus portafolios, la ganancia o pérdida acumulada y sus variaciones. Los valores positivos se muestran en verde y los negativos en rojo.
- **Crecimiento del portafolio:** gráfica de evolución de tu patrimonio a lo largo del tiempo. Puedes ajustar el periodo mostrado.
- **Asignación de activos:** distribución porcentual de tu dinero entre los distintos tipos de activos.
- **Transacciones recientes:** las últimas operaciones registradas, con acceso rápido al detalle.

El resumen puede consolidarse en una **moneda** concreta; la conversión se realiza con las tasas de cambio registradas en el sistema.

> **Consejo:** si acabas de crear tu cuenta, el Dashboard aparecerá vacío. Sigue este orden para empezar: (1) registra tus **Plataformas**, (2) crea un **Portafolio**, (3) añade **posiciones** y (4) registra o importa tus **transacciones**.

---

## 6. Portafolios

Un **portafolio** agrupa un conjunto de posiciones (activos) con un objetivo común: por ejemplo, "Jubilación", "Cripto" o "Inversión a corto plazo".

### 6.1 Ver tus portafolios

En **Portafolios** verás:

- Un **encabezado con el valor total** y la ganancia/pérdida global (en verde si es positiva, en rojo si es negativa).
- La sección **"Tus Portafolios"**, con una tarjeta por portafolio mostrando su valor, rendimiento y moneda.

Pulsa sobre cualquier portafolio para abrir su **detalle**.

### 6.2 Crear un portafolio

1. En **Portafolios**, pulsa **Crear / Añadir portafolio**.
2. Completa el formulario:
   - **Nombre del Portafolio** (obligatorio) — p. ej. "Mi Portafolio Principal".
   - **Descripción** (opcional) — el propósito del portafolio.
   - **Tipo de Portafolio** (obligatorio) — la clase de portafolio que corresponda.
   - **Moneda** (obligatorio) — la moneda base en la que se valorará el portafolio.
   - **Nivel de Riesgo** (obligatorio) — elige entre los niveles del catálogo (p. ej. conservador, moderado, agresivo).
   - **Monto Objetivo** (opcional) — la meta de valor que quieres alcanzar.
3. Guarda. El nuevo portafolio aparecerá en tu lista.

### 6.3 Detalle de un portafolio

La página de detalle muestra:

- **Indicadores principales:** valor actual, ganancia/pérdida total, y métricas del portafolio.
- **Distribución por tipo:** gráfica con el reparto del portafolio por tipo de activo.
- **Posiciones:** la lista de activos que componen el portafolio, con su cantidad, valor y rendimiento. Desde aquí puedes añadir posiciones nuevas o entrar al detalle de cada activo.
- **Crecimiento:** evolución histórica del valor del portafolio.
- **Mayor transacción:** la operación de mayor importe registrada en el portafolio.

### 6.4 Editar un portafolio

En el detalle del portafolio, pulsa **Editar portafolio** para modificar su nombre, descripción, tipo, riesgo, moneda o monto objetivo, y guarda los cambios.

---

## 7. Posiciones y activos

Una **posición** (entry) representa la tenencia de un activo concreto dentro de un portafolio: por ejemplo, 10 acciones de AAPL en tu portafolio "Principal".

### 7.1 Añadir una posición

1. Entra al **detalle del portafolio** y pulsa **Añadir posición / activo**.
2. **Busca el activo** escribiendo su ticker o nombre (p. ej. `AAPL`, `Bitcoin`). La aplicación busca en el catálogo de activos del sistema.
3. Indica:
   - **Cantidad** — número de unidades que posees.
   - **Precio** — precio de compra por unidad (p. ej. `150.50`).
   - **Plataforma** — dónde tienes el activo (de tu lista de Plataformas).
   - **Notas** (opcional) — cualquier comentario sobre la posición.
4. Guarda. La posición aparecerá en la lista del portafolio y sus importes se sumarán al valor total.

> **Nota:** si el activo que buscas no existe en el catálogo, contacta con un administrador para que lo añada (sección 16.3).

### 7.2 Detalle de un activo en el portafolio

Al pulsar sobre una posición se abre la vista del activo dentro de ese portafolio, con:

- **Resumen de Posición:** cantidad, valor actual, coste, ganancia/pérdida.
- **Información del Activo:** tipo, exchange donde cotiza, moneda, porcentaje de **asignación** dentro del portafolio, número de **transacciones** y **ROI** (retorno sobre la inversión).
- **Historial de Transacciones:** todas las compras y ventas de ese activo, paginadas.

Desde esta vista también puedes **registrar nuevas transacciones** del activo (ver sección 10).

### 7.3 Precios de los activos

Los precios de los activos del catálogo se actualizan mediante sincronización con proveedores de datos de mercado (tarea que ejecutan los administradores) o mediante precio manual fijado por un administrador. El valor de tus posiciones se recalcula con el último precio disponible.

---

## 8. Inversiones

La sección **Inversiones** reúne el catálogo de productos y oportunidades de inversión:

- **Portafolio de crecimiento equilibrado:** una vista de productos organizados según su perfil.
- **Oportunidades activas:** productos de inversión disponibles, con su información clave.

### 8.1 Ver una inversión

Pulsa sobre cualquier producto para ver su detalle: descripción, tipo de instrumento, categoría, retorno esperado, plazo, nivel de riesgo, inversión mínima y estado.

### 8.2 Añadir un producto de inversión

1. En **Inversiones**, pulsa **Añadir**.
2. Completa el formulario:
   - **Nombre** — p. ej. "Fondo Crecimiento Global".
   - **Tipo de Instrumento** — fondo, bono, acción, etc.
   - **Categoría** — clasificación del producto.
   - **Retorno esperado (%)** — p. ej. `12.5`.
   - **Plazo** — horizonte temporal (p. ej. `18` meses).
   - **Nivel de Riesgo** — perfil de riesgo del producto.
   - **Inversión Mínima ($)** — importe mínimo de entrada (p. ej. `1000`).
   - **Estado del Producto** — activo, cerrado, etc.
3. Guarda para publicarlo en el listado.

---

## 9. Plataformas

Las **plataformas** (también llamadas *fuentes*) son los lugares donde custodias tus activos: brokers, exchanges de criptomonedas, bancos, etc. Registrarlas te permite saber siempre **dónde** está cada inversión.

> **Recuerda:** Finexia solo guarda el nombre y la descripción de la plataforma. **Nunca** se conecta a ella ni almacena tus credenciales.

### 9.1 Ver tus plataformas

En **Plataformas** aparece la sección **"Tus Plataformas"** con las que hayas registrado. Si aún no tienes ninguna, verás el mensaje "No hay plataformas registradas" con la opción de crear la primera.

### 9.2 Añadir una plataforma

1. Pulsa **Añadir plataforma**.
2. Introduce:
   - **Nombre** — p. ej. "Interactive Brokers", "Binance", "Mi banco".
   - **Descripción** — información adicional (opcional).
3. Guarda. La plataforma quedará disponible para asociarla a posiciones y transacciones.

### 9.3 Editar o eliminar una plataforma

Desde la tarjeta o el detalle de la plataforma puedes **editar** sus datos o **eliminarla**. Ten en cuenta que la eliminación es permanente; si la plataforma tiene posiciones asociadas, revisa primero su contenido.

---

## 10. Transacciones

Las **transacciones** son las operaciones de compra y venta que dan forma a tus posiciones. Mantenerlas al día es la clave para que los valores, ganancias y reportes sean fieles a la realidad.

### 10.1 Ver el historial

La sección **Transacciones** muestra el historial de operaciones más recientes de todos tus portafolios, con fecha, activo, tipo de operación, cantidad, precio e importe. Desde el detalle de cada posición (sección 7.2) puedes ver el historial filtrado por activo.

### 10.2 Registrar una transacción

1. Entra a la posición correspondiente (portafolio → posición) o usa la acción **Nueva transacción**.
2. Indica:
   - **Tipo de operación** — compra o venta.
   - **Fecha** de la operación.
   - **Cantidad** de unidades.
   - **Precio** por unidad.
   - **Comisiones** u otros costes, si aplica.
3. Guarda. La posición y el portafolio se recalculan automáticamente.

### 10.3 Editar una transacción

Si cometiste un error, abre la transacción desde el historial y pulsa **Editar**. Corrige los datos y guarda; los totales se actualizarán.

---

## 11. Importación masiva de transacciones (Excel/CSV)

Si ya llevas tu registro en una hoja de cálculo, no necesitas volver a teclearlo todo: Finexia puede **importar tus transacciones desde un archivo Excel o CSV** en tres pasos: subir el archivo, mapear las columnas y confirmar.

### 11.1 Paso 1 — Subir el archivo

1. Ve a **Transacciones → Importar**.
2. **Arrastra tu Excel** a la zona de carga **o haz clic para buscarlo** en tu equipo.
3. La aplicación analizará el archivo ("Analizando tu archivo…"). Si el libro tiene varias hojas, selecciona la **hoja** que contiene las transacciones.

### 11.2 Paso 2 — Asignar columnas (mapeo)

En la pantalla **"Asigna tus columnas"** debes indicar qué columna de tu archivo corresponde a cada dato que Finexia necesita (activo/símbolo, fecha, tipo de operación, cantidad, precio, etc.):

- La aplicación muestra una **vista previa** de tus datos para ayudarte a identificar cada columna.
- En **Valores por defecto** puedes definir datos que se aplicarán a las filas donde tu archivo no tenga ese dato (por ejemplo, una moneda o un tipo de operación por defecto).
- Selecciona también el **portafolio de destino** y la **plataforma (fuente)** a la que se asociarán las transacciones importadas.

### 11.3 Paso 3 — Vista previa y confirmación

1. Revisa la **vista previa del import**: la aplicación muestra cómo quedará cada transacción interpretada.
2. Consulta la sección **"Filas omitidas"**: son filas que no se importarán por tener datos incompletos o no interpretables; se listan con el motivo para que puedas corregir tu archivo si lo deseas.
3. Si todo es correcto, pulsa **Confirmar importación**. Las transacciones se crearán en el portafolio elegido y verás el resultado del proceso.

> **Consejos para un buen import:**
> - Usa una fila de encabezados clara en tu hoja.
> - Mantén formatos de fecha y número consistentes.
> - Los símbolos/tickers deben coincidir con los del catálogo de activos.
> - Puedes repetir la vista previa tantas veces como necesites antes de confirmar; nada se guarda hasta el paso final.

---

## 12. Reportes y exportaciones

La sección **Reportes** concentra el análisis de rendimiento:

### 12.1 Reportes en pantalla

- **Performance Calendar (%):** calendario de rendimiento con el porcentaje de ganancia/pérdida por periodo (mes a mes), para identificar de un vistazo los mejores y peores tramos.
- **Key Statistics:** estadísticas clave del portafolio (rendimiento, riesgo/volatilidad y otros indicadores agregados).
- **Growth Projection:** proyección de crecimiento estimada a futuro según la evolución de tu portafolio.

### 12.2 Descargas en Excel (XLSX)

Desde **Reportes → Descargar** puedes exportar tus datos en archivos Excel:

| Archivo | Contenido |
|---|---|
| `resumen-mensual.xlsx` | Resumen mensual del portafolio |
| `transacciones.xlsx` | Historial completo de transacciones |
| `riesgo-volatilidad.xlsx` | Métricas de riesgo y volatilidad |

Los archivos se descargan directamente a tu equipo y puedes abrirlos con Excel, LibreOffice o Google Sheets.

---

## 13. Notificaciones

En **Notificaciones** configuras cómo quieres que Finexia te avise:

### 13.1 Correo electrónico

- **Alertas de actividad:** recibe un correo cuando ocurra actividad relevante en tu cuenta.
- **Resumen semanal:** un correo periódico con el resumen de la evolución de tus portafolios.

### 13.2 Alertas en la app

Activa o desactiva los avisos que se muestran dentro de la propia aplicación.

Marca o desmarca cada opción según tu preferencia; los cambios se guardan en tus preferencias de usuario.

---

## 14. Configuración de la cuenta

La página **Configuración** agrupa todo lo relativo a tu cuenta, en secciones:

### 14.1 Perfil

- Edita tu **nombre** y datos personales.
- **Sube o cambia tu avatar** (imagen de perfil), que se mostrará en la cabecera y en tu perfil.
- Tu correo electrónico identifica tu cuenta.

### 14.2 Apariencia

Ajusta las preferencias visuales de la aplicación (tema/aspecto de la interfaz) a tu gusto. Los cambios se aplican de inmediato y quedan guardados en tus preferencias.

### 14.3 Seguridad

- **Cambiar contraseña:** introduce tu contraseña actual y la nueva. Usa contraseñas largas y únicas.
- **Verificación en dos pasos (2FA)** y **Sesiones activas:** ver sección 15.

---

## 15. Seguridad: 2FA y sesiones

### 15.1 Activar la verificación en dos pasos (2FA)

La 2FA añade una segunda barrera al inicio de sesión: además de tu contraseña, necesitarás un código temporal generado por tu aplicación de autenticación.

1. Ve a **Configuración → Verificación en dos pasos (2FA)** y pulsa **Activar**.
2. **Escanea el código QR** con tu aplicación de autenticación (Google Authenticator, Authy, 1Password, etc.). Si no puedes escanear, usa la **"Clave para ingreso manual"** que se muestra junto al QR.
3. Introduce el **código de 6 dígitos** que genera la aplicación para confirmar la activación.
4. **Guarda tus códigos de recuperación.** La aplicación te mostrará una lista de códigos de un solo uso: descárgalos o cópialos y guárdalos en un lugar seguro (gestor de contraseñas, papel en lugar protegido). Son tu única vía de acceso si pierdes el teléfono.

A partir de ese momento, cada inicio de sesión pedirá el código TOTP.

### 15.2 Códigos de recuperación y desactivación

- **Usar un código de recuperación:** en el segundo paso del login, introduce uno de tus códigos de recuperación en lugar del código TOTP. Cada código sirve **una sola vez**.
- **Regenerar códigos:** desde la sección 2FA puedes generar una nueva lista (la anterior queda invalidada).
- **Desactivar 2FA:** desde la misma sección, confirmando con un código válido. Tu cuenta volverá a protegerse solo con contraseña (no recomendado).

### 15.3 Sesiones activas

En **Configuración → Sesiones activas** verás todos los dispositivos/navegadores con sesión abierta en tu cuenta, con información para identificarlos.

- **Revocar una sesión:** cierra la sesión de un dispositivo concreto.
- **Cerrar las demás sesiones:** cierra todas las sesiones excepto la actual. Útil si sospechas que alguien más accedió a tu cuenta (en ese caso, cambia también tu contraseña).

---

## 16. Panel de administración

> Esta sección aplica solo a cuentas con rol **administrador**. Las rutas de administración están protegidas: un usuario estándar recibirá un error de acceso denegado.

El grupo **Panel Admin** de la barra lateral incluye un panel con **accesos rápidos** y tres áreas:

### 16.1 Usuarios

Gestión completa de las cuentas de la plataforma:

- **Invitar a un nuevo usuario:** introduce el correo del invitado y envía la invitación. El invitado recibirá un enlace para aceptar y fijar su contraseña (sección 3.1-B).
- **Invitaciones pendientes:** lista de invitaciones enviadas y no aceptadas, con acciones para **reenviar** (si caducó o no llegó) o **revocar** (anula el enlace).
- **Lista de espera:** correos apuntados a la waitlist desde la página pública; desde aquí puedes invitarlos.
- **Usuarios registrados:** listado paginado de todos los usuarios. Para cada uno puedes:
  - **Ver y editar** sus datos.
  - **Banear/desbanear:** un usuario baneado no puede iniciar sesión.
  - **Eliminar** la cuenta (acción permanente; úsala con precaución).

### 16.2 Registro directo

El registro público puede estar habilitado o deshabilitado a nivel de sistema. Cuando está deshabilitado, la única vía de alta es la invitación.

### 16.3 Activos

Administración del catálogo de activos que los usuarios pueden añadir a sus portafolios:

- **Nuevo activo:** crea un activo indicando su símbolo/ticker, nombre, tipo, exchange y moneda.
- **Importar activos desde CSV/Excel:** alta masiva de activos a partir de un archivo.
- **Sincronizar precios:** actualiza los precios de todos los activos (o de uno concreto) desde el proveedor de datos de mercado.
- **Precio manual:** fija manualmente el precio de un activo cuando no exista cotización automática.

### 16.4 Tasas de Cambio

Gestión de las tasas de conversión entre monedas, usadas para consolidar portafolios multi-moneda:

- **Nueva tasa de cambio:** crea una tasa entre dos monedas.
- **Importar tasas desde CSV/Excel:** alta masiva de tasas.
- **Sincronizar:** actualiza todas las tasas (o una concreta) desde el proveedor.
- **Editar:** corrige el valor de una tasa existente.

---

## 17. Preguntas frecuentes (FAQ)

**¿Finexia se conecta a mis brokers o plataformas?**
No. Finexia nunca accede a tus plataformas ni te pide credenciales. Tú registras manualmente dónde tienes tus activos, así que la información siempre está bajo tu control.

**¿Puedo tener varios portafolios?**
Sí, puedes crear tantos como necesites, cada uno con su moneda, tipo, nivel de riesgo y monto objetivo propios.

**¿Cómo se calculan los valores de mis posiciones?**
Con el último precio disponible de cada activo en el catálogo (sincronizado desde proveedores de mercado o fijado por un administrador), multiplicado por tu cantidad, y convertido a la moneda del portafolio con las tasas de cambio del sistema.

**¿Puedo importar mi histórico desde Excel?**
Sí. Usa **Transacciones → Importar**: sube el archivo, asigna las columnas, revisa la vista previa (incluidas las filas omitidas) y confirma. Nada se guarda hasta que confirmas.

**No encuentro un activo al crear una posición. ¿Qué hago?**
El activo no está en el catálogo del sistema. Solicita a un administrador que lo añada desde **Panel Admin → Activos**.

**¿Qué pasa si pierdo mi teléfono con la app de autenticación?**
Usa uno de tus **códigos de recuperación** para entrar y luego reconfigura la 2FA. Si tampoco tienes los códigos, contacta con un administrador.

**¿Puedo usar Finexia en el móvil?**
Sí. La interfaz es adaptable; en pantallas pequeñas el menú lateral se abre desde el botón de la cabecera.

**¿Cómo exporto mis datos?**
Desde **Reportes** puedes descargar en Excel el resumen mensual, el historial de transacciones y las métricas de riesgo/volatilidad.

---

## 18. Solución de problemas

| Problema | Causa probable | Solución |
|---|---|---|
| "Correo sin verificar" al iniciar sesión | No abriste el enlace de verificación | Reenvía el correo de verificación desde la propia pantalla y revisa spam |
| "Demasiados intentos" (error 429) | Límite de peticiones por seguridad | Espera unos minutos y vuelve a intentarlo |
| El enlace de invitación o de reset no funciona | Token caducado o ya usado | Solicita un nuevo enlace (o el reenvío de la invitación al administrador) |
| No puedo registrarme | El registro directo está deshabilitado | Únete a la lista de espera o pide una invitación |
| El código 2FA no es aceptado | Reloj del teléfono desincronizado o código expirado | Sincroniza la hora del dispositivo y usa el código vigente; como alternativa, un código de recuperación |
| Mi sesión se cerró sola | La sesión fue revocada o expiró | Inicia sesión de nuevo; revisa **Sesiones activas** si no fuiste tú |
| Una importación omite filas | Datos incompletos o formatos no interpretables | Revisa el detalle de **Filas omitidas**, corrige el archivo y repite la vista previa |
| Los valores no cuadran con mi broker | Falta registrar transacciones o el precio no está actualizado | Completa el historial de transacciones; los precios se actualizan con la sincronización del catálogo |
| Acceso denegado (403) en páginas de Admin | Tu cuenta no tiene rol administrador | Solicita el rol a un administrador si corresponde |

Si el problema persiste, contacta con el administrador de tu instancia de Finexia.

---

## 19. Glosario

| Término | Definición |
|---|---|
| **Activo (asset)** | Instrumento financiero del catálogo: acción, criptomoneda, fondo, etc., identificado por su símbolo/ticker |
| **Posición (entry)** | Tenencia de un activo dentro de un portafolio (cantidad + coste) |
| **Portafolio** | Conjunto de posiciones agrupadas bajo un objetivo, con moneda y nivel de riesgo propios |
| **Plataforma (fuente)** | Broker, exchange o entidad donde custodias tus activos |
| **Transacción** | Operación de compra o venta que modifica una posición |
| **Asignación** | Porcentaje que representa un activo o tipo de activo dentro del total |
| **ROI** | *Return on Investment*: retorno sobre la inversión, en porcentaje |
| **Volatilidad** | Medida de la variabilidad del valor de un activo o portafolio |
| **2FA / TOTP** | Verificación en dos pasos con códigos temporales de 6 dígitos |
| **Códigos de recuperación** | Códigos de un solo uso para acceder si pierdes tu aplicación de autenticación |
| **Waitlist** | Lista de espera pública para solicitar acceso a la plataforma |
| **Invitación** | Enlace enviado por un administrador para crear una cuenta |
| **Tasa de cambio** | Relación de conversión entre dos monedas, usada para consolidar valores |
| **XLSX** | Formato de archivo de Excel usado en las exportaciones |

---

*Este manual describe la funcionalidad de Finexia a la fecha indicada en la portada. Las pantallas y textos pueden variar ligeramente según la versión desplegada.*

# Manual de Usuario — FINEXIA

**Versión del documento:** 1.8
**Fecha:** Septiembre 2026
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
8. [Mis Activos (vista consolidada)](#8-mis-activos-vista-consolidada)
9. [Plataformas](#9-plataformas)
10. [Transacciones](#10-transacciones)
11. [Importación masiva de transacciones (Excel/CSV)](#11-importación-masiva-de-transacciones-excelcsv)
12. [Reportes y exportaciones](#12-reportes-y-exportaciones)
13. [Notificaciones](#13-notificaciones)
14. [Configuración de la cuenta](#14-configuración-de-la-cuenta)
15. [Seguridad: 2FA y sesiones](#15-seguridad-2fa-y-sesiones)
16. [Conectar un asistente de IA (MCP)](#16-conectar-un-asistente-de-ia-mcp)
17. [Preguntas frecuentes (FAQ)](#17-preguntas-frecuentes-faq)
18. [Solución de problemas](#18-solución-de-problemas)
19. [Glosario](#19-glosario)

---

## 1. Introducción

### 1.1 ¿Qué es Finexia?

Finexia es una aplicación web para **gestionar y hacer seguimiento de tus portafolios de inversión** en un solo lugar. Te permite registrar tus activos (acciones, criptomonedas, ETFs, fondos y otros instrumentos), las plataformas donde los tienes, y todas tus transacciones de compra y venta, para luego visualizar el rendimiento, la distribución y el crecimiento de tu patrimonio con gráficas y reportes descargables.

### 1.2 Principios clave

- **Tú controlas tus datos.** Finexia **no se conecta a tus brokers ni a tus plataformas de inversión**, y nunca te pedirá las credenciales de esas cuentas. Toda la información se registra manualmente (o mediante importación de archivos Excel/CSV), por lo que siempre está bajo tu control.
- **Multi-portafolio.** Puedes crear tantos portafolios como necesites (por ejemplo: "Retiro", "Cripto", "Fondo de emergencia") y analizarlos de forma individual o agregada.
- **Multi-moneda.** Cada portafolio tiene una moneda base y la aplicación convierte automáticamente los valores entre monedas para mostrarte tu patrimonio consolidado.
- **Seguridad primero.** La aplicación incluye verificación de correo electrónico, autenticación en dos pasos (2FA), gestión de sesiones activas y protección contra intentos de acceso masivos.

---

## 2. Requisitos y acceso

### 2.1 Requisitos

- Un navegador web moderno y actualizado (Chrome, Firefox, Safari o Edge).
- Conexión a internet.
- Una dirección de correo electrónico válida (necesaria para verificar la cuenta y recuperar la contraseña).
- Opcional, pero recomendado: una aplicación de autenticación (Google Authenticator, Authy, 1Password, etc.) para activar la verificación en dos pasos.

La aplicación es **responsive**: funciona tanto en computadoras de escritorio como en tabletas y teléfonos móviles. En pantallas pequeñas el menú lateral se oculta y se abre con el botón de menú de la cabecera.

### 2.2 Página de inicio (landing) y lista de espera

Al visitar la dirección pública de Finexia sin haber iniciado sesión verás la página de presentación. Además de la descripción del producto y sus beneficios, incluye un recorrido por las cuatro vistas del panel (**El producto**), las garantías de seguridad de la plataforma (**Seguridad**), las preguntas frecuentes y las páginas legales (Términos, Privacidad y Cookies). En pantallas pequeñas el menú se abre con el botón de la esquina superior derecha.

![Página de inicio de Finexia con el formulario de acceso anticipado](img/manual/01-landing.png)

Si el registro directo no está habilitado, la página te ofrece unirte a la **lista de espera** dejando tu correo electrónico en el formulario de **"Acceso anticipado"**. El equipo de Finexia podrá luego enviarte una **invitación** para crear tu cuenta.

---

## 3. Primeros pasos: registro e inicio de sesión

### 3.1 Crear una cuenta

Existen dos vías para obtener una cuenta:

**A) Registro directo** (si está habilitado):

1. En la página de inicio, pulsa **Registrarse / Crear cuenta**.
2. Completa tus datos: nombre, correo electrónico y contraseña.
3. Envía el formulario. Recibirás un **correo de verificación**.
4. Abre el enlace del correo para **verificar tu dirección**. Sin este paso no podrás iniciar sesión (el sistema mostrará un aviso de "correo sin verificar").

> **Nota:** si el registro directo está desactivado, el formulario mostrará un aviso indicando que el registro está deshabilitado. En ese caso, únete a la lista de espera para recibir una invitación.

**B) Por invitación:**

1. Recibirás en tu correo una invitación para unirte a Finexia.
2. Abre el enlace **Aceptar invitación** del correo. La aplicación validará la invitación.
3. Define tu contraseña y confirma. Tu cuenta quedará creada y verificada, lista para iniciar sesión.

Las invitaciones tienen caducidad; si el enlace expiró, solicita que te la reenvíen.

### 3.2 Iniciar sesión

![Pantalla de inicio de sesión de Finexia](img/manual/02-login.png)

1. Ve a la página de **Iniciar sesión**.
2. Introduce tu correo y contraseña.
3. Si tienes **2FA activado**, la aplicación te pedirá un segundo paso: introduce el **código de 6 dígitos** de tu aplicación de autenticación (o uno de tus códigos de recuperación).
4. Al completar el acceso entrarás directamente al **Dashboard**.

La sesión se mantiene mediante un token de acceso de corta duración y una cookie segura de renovación que se rota automáticamente; no necesitas hacer nada para mantener la sesión activa mientras uses la aplicación.

### 3.3 ¿Olvidaste tu contraseña?

1. En la pantalla de inicio de sesión, pulsa **¿Olvidaste tu contraseña?**
2. Introduce tu correo electrónico. Si existe una cuenta asociada, recibirás un enlace de restablecimiento.
3. Abre el enlace, define tu **nueva contraseña** y confírmala.
4. Inicia sesión con la nueva contraseña.

Los enlaces de restablecimiento caducan por seguridad; si el tuyo expiró, solicita uno nuevo.

### 3.4 Verificación de correo

Si intentas iniciar sesión sin haber verificado tu correo, la aplicación te lo indicará y te permitirá **reenviar el correo de verificación**. Revisa también la carpeta de spam.

### 3.5 Cerrar sesión

Usa el botón **Cerrar Sesión** situado en la parte inferior de la barra lateral. También puedes cerrar sesiones abiertas en otros dispositivos desde **Configuración → Sesiones activas** (ver sección 15.3).

---

## 4. Interfaz general de la aplicación

Una vez dentro, la aplicación se organiza en tres zonas:

### 4.1 Cabecera (header)

Situada en la parte superior. Contiene:

- El botón de **menú** (en pantallas pequeñas) para mostrar/ocultar la barra lateral.
- Accesos rápidos y el indicador de **notificaciones**.
- Tu **avatar**, nombre y correo de la cuenta.

### 4.2 Barra lateral (menú principal)

Es el menú de navegación. Sus secciones son:

| Sección | ¿Para qué sirve? |
|---|---|
| **Dashboard** | Vista general: patrimonio neto, crecimiento y actividad reciente |
| **Portafolios** | Crear y gestionar tus portafolios y sus posiciones |
| **Mis Activos** | Cuánto tienes de cada activo, sumando todos tus portafolios |
| **Plataformas** | Registrar los brokers/exchanges/bancos donde tienes activos |
| **Transacciones** | Historial de operaciones e importación desde Excel/CSV |
| **Reportes** | Estadísticas, calendario de rendimiento y descargas en Excel |
| **Notificaciones** | Preferencias de avisos por correo y en la app |
| **Guía de usuario** | Este manual, para consultarlo en pantalla o descargarlo |
| **Configuración** | Perfil, apariencia, contraseña, 2FA, sesiones, claves de datos de mercado y acceso para asistentes |

En la parte inferior de la barra lateral está el botón **Cerrar Sesión**.

### 4.3 Área de contenido

Es la zona central donde se muestra cada página. En la mayoría de listados encontrarás paginación y acciones contextuales (crear, editar, eliminar).

### 4.4 Uso en el móvil

La interfaz se adapta automáticamente a pantallas pequeñas: el contenido ocupa todo el ancho y la barra lateral queda oculta.

![Dashboard de Finexia en un teléfono móvil](img/manual/15-movil-dashboard.png)

Para navegar, pulsa el **botón de menú** (las tres líneas de la esquina superior izquierda): la barra lateral se despliega con las mismas secciones que en escritorio y se cierra al elegir una opción o al tocar fuera de ella.

![Menú lateral desplegado en la versión móvil](img/manual/16-movil-menu.png)

Todas las funciones descritas en este manual están disponibles también desde el móvil.

### 4.5 La guía de usuario dentro de la aplicación

No hace falta buscar este manual fuera de Finexia: la entrada **Guía de usuario** del menú lateral lo abre dentro de la propia aplicación.

![Página de la guía de usuario con la ficha del documento y el índice de secciones](img/manual/17-guia.png)

La página muestra la **versión** y la **fecha** del documento que se está sirviendo, junto con su tamaño, y ofrece tres formas de usarlo:

- **Ver la guía aquí** — la abre incrustada en la página, sin salir de la aplicación.
- **Abrir en una pestaña nueva** — útil para consultarla mientras trabajas en otra sección.
- **Descargar PDF** — la guarda en tu equipo para leerla sin conexión.

Debajo aparece el **índice de secciones**, con la misma numeración que este documento, para saber de un vistazo dónde está cada cosa.

> **Sobre la versión que ves:** el PDF se genera a partir del manual del repositorio y la aplicación comprueba en cada integración que el archivo publicado corresponde al texto vigente. Si la fecha que ves es antigua, es porque el manual no ha cambiado desde entonces, no porque se haya quedado sin actualizar.

---

## 5. Dashboard (panel principal)

El Dashboard es la primera pantalla tras iniciar sesión y ofrece una fotografía general de tus finanzas.

![Dashboard de Finexia con el patrimonio neto, el crecimiento del portafolio, la asignación de activos y la actividad reciente](img/manual/03-dashboard.png)

Sus bloques principales son:

- **Patrimonio Neto:** el valor total de todos tus portafolios, con la ganancia acumulada en importe y porcentaje, el número de portafolios y de posiciones. Con el selector de moneda (p. ej. **USD/COP**) puedes ver el total consolidado en la moneda que prefieras.
- **Crecimiento del portafolio:** gráfica de evolución de tu patrimonio que compara el **valor de mercado** (línea continua ámbar) con el **capital invertido** (línea discontinua). Puedes cambiar el periodo mostrado (1M, 3M, 6M, 1Y o Todo), pasar la gráfica a **porcentaje** con el conmutador **Valor / %**, y ver la ganancia total, el rendimiento, la **rentabilidad real** del periodo y el valor actual.
- **Portafolios:** tabla resumen con cada portafolio, su tipo, valor actual, importe invertido y ganancia/pérdida (en verde si es positiva, en rojo si es negativa), junto con los totales. El nombre de cada fila lleva a su detalle.
- **Asignación de Activos:** gráfica de dona con la distribución porcentual de tu dinero entre tipos de activos (acciones, ETF, criptomonedas, fondos…).
- **Actividad Reciente:** las últimas compras y ventas registradas, con acceso a **Ver todo** y a **Descargar extracto**.

### 5.1 Consultar un día concreto en la gráfica

La gráfica de crecimiento no es solo una imagen: **pasa el cursor por encima** y una línea vertical marcará el día bajo el ratón. Justo encima de la gráfica aparecerá su detalle: la fecha, el **valor de mercado**, el **capital invertido** y la **ganancia** de ese día.

Si prefieres el teclado, pulsa **Tab** hasta llegar a la gráfica y recórrela con las **flechas izquierda y derecha**; **Inicio** y **Fin** saltan al primer y último día, y **Esc** quita la selección. Cada punto se anuncia en voz alta, de modo que la gráfica también se puede leer con un lector de pantalla.

### 5.2 Ver la gráfica en porcentaje

Junto a los botones de periodo hay un conmutador con dos posiciones: **Valor** y **%**. El primero es la vista de siempre, en dinero. El segundo dibuja la misma serie en **rentabilidad**, y es la que contesta a la pregunta de cuánto has ganado, no de cuánto tienes.

La diferencia importa más de lo que parece. En dinero, tu patrimonio sube cuando el mercado te da la razón **y también cuando metes un depósito**: las dos cosas empujan la curva hacia arriba y la gráfica no las distingue. En porcentaje se dibujan dos líneas que sí lo hacen:

| Línea | Qué mide |
|---|---|
| **Rentabilidad acumulada** (ámbar) | Lo que rindió tu dinero desde el inicio del periodo que estés viendo, **descontando aportes y retiros**. Un depósito no la mueve: solo el mercado la mueve |
| **Ganancia sobre coste** (gris discontinua) | Lo que vale tu cartera frente a lo que te costó, día a día. Esta sí depende de *cuándo* aportaste |

Cuando las dos líneas se separan, la distancia entre ellas es exactamente el efecto de tus movimientos de dinero: aportar antes de una subida las junta, aportar después las separa. La línea horizontal más marcada es el **equilibrio (0 %)**, la frontera entre ganar y perder.

Sobre la gráfica, la métrica **Rentabilidad real** lleva el periodo en su etiqueta (`Rentabilidad real · 3M`) porque, a diferencia de las otras tres, mide el tramo que estás viendo. Es la misma cifra que la línea ámbar al final de su recorrido.

> **Por qué no coincide con «Rendimiento».** El **Rendimiento** de al lado divide tu ganancia de hoy entre lo que llevas invertido hoy. Si metiste una cantidad grande justo después de una subida, esa cifra se hunde aunque tu cartera no haya perdido nada: el mismo beneficio repartido entre mucho más capital. La **Rentabilidad real** encadena lo que pasó en cada tramo y solo se mueve con el mercado. Las dos son correctas; contestan a preguntas distintas.

### 5.3 Leer la asignación de activos

En la dona, al señalar una porción —o su entrada en la leyenda— el resto se atenúa y el centro pasa a mostrar **esa categoría** con su porcentaje, en lugar del total. La leyenda funciona igual con el teclado y con un clic: cada entrada indica el importe y el peso de esa categoría dentro de tu patrimonio.

> **Consejo:** si acabas de crear tu cuenta, el Dashboard aparecerá vacío. Cada bloque te dirá qué falta y te ofrecerá el siguiente paso (crear un portafolio, importar transacciones…). El orden recomendado es: (1) registra tus **Plataformas**, (2) crea un **Portafolio**, (3) añade **posiciones** y (4) registra o importa tus **transacciones**.

---

## 6. Portafolios

Un **portafolio** agrupa un conjunto de posiciones (activos) con un objetivo común: por ejemplo, "Jubilación", "Cripto" o "Inversión a corto plazo".

### 6.1 Ver tus portafolios

![Listado de portafolios con el valor total y la ganancia global](img/manual/04-portafolios.png)

En **Portafolios** verás:

- Un **encabezado con el valor total** y la ganancia/pérdida global de todos tus portafolios.
- La sección **"Tus Portafolios"**, con una tarjeta por portafolio mostrando su valor, rendimiento, tipo y moneda.

Pulsa sobre cualquier portafolio para abrir su **detalle**.

### 6.2 Crear un portafolio

![Formulario de creación de un portafolio](img/manual/06-crear-portafolio.png)

1. En **Portafolios**, pulsa **Crear / Añadir portafolio**.
2. Completa el formulario:
   - **Nombre del Portafolio** (obligatorio) — p. ej. "Mi Portafolio Principal".
   - **Descripción** (opcional) — el propósito del portafolio.
   - **Tipo de Portafolio** (obligatorio) — la clase de portafolio que corresponda (largo plazo, especulativo…).
   - **Moneda** (obligatorio) — la moneda base en la que se valorará el portafolio.
   - **Nivel de Riesgo** (obligatorio) — elige entre los perfiles disponibles: **Conservador** (prioriza preservar el capital), **Moderado** (equilibrio entre crecimiento y estabilidad) o **Agresivo** (busca máximo crecimiento asumiendo alta volatilidad).
   - **Monto Objetivo** (opcional) — la meta de valor que quieres alcanzar.
3. Guarda. El nuevo portafolio aparecerá en tu lista.

### 6.3 Detalle de un portafolio

![Detalle de un portafolio con sus indicadores, la distribución por tipo y las posiciones](img/manual/05-portafolio-detalle.png)

La página de detalle muestra:

- **Indicadores principales:** valor de mercado, ganancia/pérdida total, **rentabilidad real** y riesgo/número de activos. La rentabilidad real es la del portafolio limpia de aportes y retiros —la misma cuenta que explica el apartado 5.2—, así que puede no parecerse a la ganancia sobre costo de al lado; aparece como `—` hasta que el portafolio acumule al menos dos cierres diarios.
- **Distribución por tipo:** gráfica con el reparto del portafolio por tipo de activo.
- **Posiciones:** la lista de activos que componen el portafolio, con su cantidad, precio de compra, precio de mercado y rendimiento. Desde aquí puedes añadir posiciones nuevas o entrar al detalle de cada activo.
- **Crecimiento:** evolución histórica del portafolio, con el mismo conmutador **Valor / %** que la gráfica del Dashboard.
- **Mayor transacción:** la operación de mayor importe registrada en el portafolio.

### 6.4 Editar un portafolio

En el detalle del portafolio, pulsa **Editar portafolio**: el formulario se abre en una ventana sobre la página con el nombre, la descripción, el tipo, el nivel de riesgo y el monto objetivo. Al guardar, la ventana se cierra y la página confirma que se actualizó.

---

## 7. Posiciones y activos

Una **posición** representa la tenencia de un activo concreto dentro de un portafolio: por ejemplo, 42 acciones de AAPL en tu portafolio "Principal".

### 7.1 Añadir una posición

1. Entra al **detalle del portafolio** y pulsa **Añadir posición / activo**.
2. Elige la **plataforma** donde tienes el activo, de tu lista de Plataformas.
3. **Busca el activo** escribiendo su ticker o nombre (p. ej. `AAPL`, `Bitcoin`). La aplicación busca en el catálogo de activos disponible.
4. Completa los **detalles de compra**:
   - **Cantidad** — número de unidades que posees.
   - **Precio de compra** — precio por unidad (p. ej. `150.50`), con el selector de **moneda de la operación** al lado: llega puesta la del activo, y se cambia si tu bróker la ejecutó en otra. Se admiten todos los decimales que haga falta, así que un precio de `0.00004182` se guarda tal cual y no redondeado.
   - **Fecha de compra**.
5. Si tu cuenta liquidó en una moneda distinta de la de la operación, marca **"Mi cuenta liquidó en otra moneda"** e indica la moneda y la tasa (apartado 7.4).
6. Añade las **notas** que quieras y guarda. La posición aparecerá en la lista del portafolio y sus importes se sumarán al valor total.

> **¿No aparece el activo que buscas?** El buscador te ofrece **Crear "TICKER"**. Rellena nombre, tipo y moneda (el mercado es opcional) y el activo queda disponible al instante para seguir con la posición. Los activos que creas los ves solo tú, hasta que el equipo de Finexia los incorpore al catálogo general. Puedes añadir hasta 50 activos nuevos cada 24 horas.

### 7.2 Detalle de un activo en el portafolio

Al pulsar sobre una posición se abre la vista del activo dentro de ese portafolio (el botón **Volver** te devuelve al detalle del portafolio).

![Detalle del activo AAPL con el resumen de posición, la información del activo y el historial de transacciones](img/manual/13-activo-detalle.png)

La vista incluye:

- El **precio de mercado** actual del activo, junto a su ticker, nombre, tipo y exchange.
- **Resumen de Posición:** cantidad total, precio promedio de compra, precio actual, costo total, valor de mercado y **ganancia/pérdida** en importe y porcentaje.
- **Información del Activo:** tipo, exchange donde cotiza, moneda, porcentaje de **asignación** dentro del portafolio, número de **transacciones** y **ROI** (retorno sobre la inversión).
- **Historial de Transacciones:** todas las compras y ventas de ese activo, con tipo, fecha, cantidad, precio, comisión, total y notas.

Desde esta misma vista se gestiona todo el historial. Cada acción abre su propia ventana sobre la página:

- **+ Agregar** — registra una transacción nueva sobre esta posición (sección 10.2).
- **Vender** — abre el panel de venta con el lote de esa fila, entero o en parte (sección 10.4).
- **Editar** — corrige cualquier dato de una operación ya registrada.
- **Eliminar** — borra una operación. El diálogo enseña el tipo, la fecha y el total de la que vas a quitar, porque en una tabla todas las filas se parecen y esto no se puede deshacer. La posición se recalcula con las que queden; si era la última, la cantidad pasa a 0.

Al final de la página está **Eliminar posición**, que no es lo mismo: quita la posición del portafolio **junto con todas sus transacciones**, y por eso el diálogo dice cuántas se van con ella. Ten en cuenta que hay una posición por plataforma —el mismo ticker comprado en dos brókers son dos posiciones—, así que cada una se elimina por separado.

### 7.3 Precios de los activos

Los precios vienen de proveedores de datos de mercado (Finnhub, Alpha Vantage), y **Finexia no tiene cuenta con ninguno de ellos**: la clave la pones tú, en *Configuración → Datos de mercado* (sección 14.4). Es gratis y se tarda un minuto; el manual lo explica ahí.

Qué cambia según tengas clave o no:

| | Con tu clave configurada | Sin clave |
|---|---|---|
| Precio de tus posiciones | Última cotización disponible | Tu **precio de compra** |
| Actualización | Automática cada día, más el botón **Sincronizar** cuando quieras | No hay nada que actualizar |
| Ganancia/pérdida | Real frente al mercado | Sale 0: estás comparando el coste consigo mismo |

Sin clave la aplicación **no inventa un precio ni usa el de otro usuario**: los datos que trae una clave personal pertenecen a quien la puso, y las condiciones de uso de los proveedores no permiten compartirlos. El panel principal te avisa cuando no hay ninguna clave usable, para que no confundas una valoración a coste con una de mercado.

Un administrador puede además fijar a mano el precio de un activo del catálogo. Ese precio sí lo ve todo el mundo, y se usa solo cuando tú no tienes dato propio para ese activo.

### 7.4 Cuando la operación y tu cuenta van en monedas distintas

Un activo puede cotizar en una moneda y tu cuenta liquidar en otra: compras un ETF que cotiza en euros desde una cuenta en dólares y el bróker convierte por su cuenta. Son dos cifras distintas —el precio al que se ejecutó y lo que salió de tu cuenta— y Finexia guarda las dos.

En el formulario de compra hay un interruptor: **"Mi cuenta liquidó en otra moneda"**. Mientras esté apagado no hay nada que decidir, porque la operación y la cuenta van en la misma moneda. Al encenderlo aparecen dos campos:

- **Moneda de la cuenta** — aquella en la que el bróker debitó el importe. Es la moneda en la que queda el coste de la posición.
- **Tasa de la operación** — cuántas unidades de la moneda de la cuenta costaba una de la moneda de la operación **ese día**.

> **Copia la tasa de la confirmación del bróker, no la de hoy.** La de hoy convierte una compra vieja a un precio que nunca pagaste, y deja mal el coste de la posición y, con él, toda su ganancia. Por eso los campos están detrás de un interruptor en vez de a la vista: una casilla de tasa siempre visible invita a rellenarla con la cotización del momento.

Si dejas las dos monedas iguales con el interruptor encendido, la aplicación te lo dice junto al campo y guarda la tasa como 1: una moneda no se convierte en sí misma.

Lo mismo aparece al **vender** y al **registrar o editar una transacción** de una posición de este tipo, con un añadido: un selector para decir en cuál de las dos monedas te cobraron la **comisión**. Y junto a la tasa verás siempre el importe que resulta en la moneda de tu cuenta, para contrastarlo con el que te debitaron antes de guardar.

---

## 8. Mis Activos (vista consolidada)

Los portafolios contestan a "cómo tengo repartido el dinero". **Mis Activos** contesta a la otra pregunta: **cuánto tengo de cada cosa**, sin importar en qué portafolio esté. Si compraste AAPL en tres portafolios distintos, aquí es una sola fila con la suma.

![Vista Mis Activos con el valor total, la barra de concentración y la lista consolidada](img/manual/18-mis-activos.png)

La página se abre desde **Mis Activos**, en el menú lateral, y tiene tres partes:

- **El valor total:** lo que suma todo lo que tienes, en la moneda que elijas arriba a la derecha, y en cuántos activos distintos está repartido. Es la cifra a la que se refieren todos los porcentajes de la página.
- **Cómo está repartido:** una sola barra con la cartera entera, cortada de mayor a menor. Cada franja es un activo y su ancho es lo que pesa, así que **el punto medio de la barra es la mitad de tu dinero**: la marca del centro lo señala y el pie dice cuántos activos caben a su izquierda. Dos, y tu dinero está en pocas manos; quince, y está repartido. Al señalar una franja —o la fila del activo en la lista de abajo— el pie pasa a decir su importe y su peso.
- **Tus activos:** la lista, con una fila por activo. El **buscador** de la cabecera filtra por nombre o por símbolo.

| Columna | Qué dice |
|---|---|
| **Activo** | Ticker y nombre. Si lo tienes en más de un portafolio, lo dice debajo |
| **Clase** | Acción, ETF, cripto, bono, efectivo… |
| **Posición** | Las unidades sumadas de todos tus portafolios, y a cómo está una |
| **Valor** | Lo que suma esa tenencia, en la moneda de la cabecera |
| **Peso** | Cuánto pesa sobre el total |

El fondo de cada fila es una barra: mide la posición contra la mayor que tengas, así que de un vistazo se ve cuáles son las grandes sin tener que leer los porcentajes. La escala es la de la cartera entera, no la de la hoja que estés viendo.

La lista se pagina de quince en quince; la barra de concentración siempre reparte la cartera entera, no la página ni el resultado de la búsqueda.

Dos detalles al leerla:

- El **precio por unidad** va en la moneda en la que cotiza el activo, no en la de la columna *Valor*: es lo que vale una unidad en su mercado, no lo que se convirtió. Cuando no hay precio de mercado (apartado 7.3), la fila lo dice con la etiqueta **a coste, sin precio de mercado**.
- Si algún activo está en una moneda para la que no hay tasa de cambio, entra en el total **sin convertir** y la página lo avisa arriba. En ese caso el total y los pesos mezclan monedas, así que tómalos como una aproximación.

---

## 9. Plataformas

Las **plataformas** son los lugares donde custodias tus activos: brokers, exchanges de criptomonedas, bancos, etc. Registrarlas te permite saber siempre **dónde** está cada inversión y **qué parte** guarda cada sitio.

> **Recuerda:** Finexia solo guarda el nombre y la descripción de la plataforma. **Nunca** se conecta a ella ni almacena tus credenciales.

### 9.1 Ver tus plataformas

![Listado de plataformas ordenadas por lo que guarda cada una, con la barra de reparto encima](img/manual/07-plataformas.png)

La página abre con el **reparto de tu dinero**: el total de la cuenta y una barra en la que cada tramo es una plataforma, del tamaño de lo que guarda. Debajo, la tabla las ordena **de mayor a menor**, y la regla de color bajo cada nombre repite su tramo de la barra con el mismo tono.

| Columna | Qué dice |
|---|---|
| **Plataforma** | Nombre, tipo (bróker, casa de bolsa, billetera cripto…) y qué porcentaje de la cuenta guarda |
| **Posiciones** | Cuántas posiciones abiertas custodia |
| **Invertido** | Lo que te costó lo que sigue en cartera, en la moneda de tu cuenta |
| **Ganancia** | La diferencia frente al valor de mercado, en importe y en porcentaje |

La lista se pagina de doce en doce. Si todavía no tienes ninguna plataforma, la página te ofrece crear la primera.

> **La columna de ganancia aparece y desaparece.** Solo se dibuja si al menos una plataforma tiene con qué calcularla: sin precios de mercado (apartado 7.3) sería una columna de guiones, que se lee como un dato que falta y no como uno que no existe.

### 9.2 Añadir una plataforma

1. Pulsa **Añadir plataforma**.
2. Introduce:
   - **Nombre** — p. ej. "Interactive Brokers", "Binance", "Mi banco".
   - **Descripción** — información adicional (opcional).
   - **Tipo de plataforma** — bróker, banco de inversión, plataforma de trading, neobank, DeFi, billetera cripto, fondos mutuos, casa de bolsa u otro.
3. Guarda. La plataforma quedará disponible para asociarla a posiciones y transacciones.

### 9.3 Detalle de una plataforma

![Detalle de una plataforma con lo invertido, lo que vale hoy, la diferencia y las posiciones](img/manual/19-plataforma-detalle.png)

Al pulsar el nombre de una plataforma se abre su ficha. Bajo el título van su tipo y la fecha de alta, y después las cifras, en línea porque solo significan algo leídas juntas —la diferencia **es** la resta de las otras dos—:

- **Invertido** — lo que te costaron las posiciones que custodia, y qué porcentaje de tu cuenta representa.
- **Vale hoy** — su valor de mercado. Debajo, cuántas posiciones se están valorando a su propio coste por no tener precio.
- **Diferencia** — la ganancia o la pérdida, en importe y en porcentaje sobre lo invertido.
- **Posiciones** — cuántas hay, y sobre cuántos **activos** y **portafolios** se reparten: diez posiciones son una cosa muy distinta si son diez empresas que si son una empresa repetida en diez portafolios.

La ficha avisa cuando toca. Si **ninguna** posición tiene precio de mercado, explica que el valor de mercado repite lo invertido y que la ganancia sale en cero porque no hay con qué compararla, no porque la plataforma no se haya movido. Y si alguna posición se contó en su propia moneda por falta de tasa de cambio, también lo dice: entonces el total suma monedas distintas.

### 9.4 Editar o eliminar una plataforma

**Editar** abre el formulario en una ventana sobre la ficha, con el nombre, la descripción y el tipo.

**Eliminar** pide confirmación y es permanente. Una plataforma que todavía tenga posiciones registradas —**incluidas las que ya vendiste**, porque siguen siendo tu historial— **no se puede eliminar**: Finexia se niega y te lo explica en el propio diálogo en lugar de arrastrar tus posiciones con ella. Elimina primero esas posiciones (apartado 7.2) y vuelve a intentarlo.

---

## 10. Transacciones

Las **transacciones** son las operaciones de compra y venta que dan forma a tus posiciones. Mantenerlas al día es la clave para que los valores, ganancias y reportes sean fieles a la realidad.

### 10.1 Ver el historial

![Historial de transacciones con fecha, activo, tipo de operación e importe](img/manual/08-transacciones.png)

La sección **Transacciones** muestra el historial de operaciones más recientes de todos tus portafolios, con fecha, activo, tipo de operación (compra o venta), cantidad, precio e importe. Desde el detalle de cada posición (sección 7.2) puedes ver el historial filtrado por activo.

### 10.2 Registrar una transacción

Las transacciones se registran desde la posición a la que pertenecen: entra al **portafolio**, abre la **posición** y pulsa **+ Agregar**. (Para cargar muchas de golpe, usa la importación de la sección 11.)

1. Elige el **tipo de operación**. El formulario se adapta a lo que elijas:

   | Tipo | Qué pide |
   |---|---|
   | **Compra**, **Venta** | Cantidad, precio unitario y comisión |
   | **Dividendo**, **Interés**, **Comisión** | Solo el importe |
   | **Transferencia de entrada**, **Transferencia de salida** | Cantidad y precio |
   | **Split** | Las nuevas unidades recibidas |

2. Indica la **fecha** y, si tienes el mismo activo en más de una plataforma, en cuál de ellas fue.
3. Añade la **comisión**, si la hubo, y las **notas** que quieras.
4. Si la operación se liquidó en otra moneda, indica la **tasa** (apartado 7.4). Al lado verás el **coste en la moneda de tu cuenta** que sale de lo que llevas tecleado, para contrastarlo con lo que te debitaron.
5. Guarda. La posición y el portafolio se recalculan automáticamente.

### 10.3 Editar o eliminar una transacción

En el historial de la posición, cada fila tiene sus acciones:

- **Editar** — abre la operación en una ventana con todos sus datos, la tasa incluida. Corrige lo que haga falta y guarda; los totales se actualizan.
- **Eliminar** — la borra tras una confirmación que muestra el tipo, la fecha y el total de esa fila, porque en la tabla todas se parecen. La posición se recalcula con las que queden; si era la última, la cantidad pasa a 0. No se puede deshacer.

### 10.4 Registrar una venta

El botón **Vender** de una fila del historial abre el panel de venta con ese lote:

- Vende **el lote entero** o **una parte**; en ese caso indícala **por unidades o por importe**, lo que tengas a mano de la confirmación del bróker.
- El **precio** llega sugerido del mercado: cámbialo por el que se ejecutó de verdad.
- Si la venta liquidó en otra moneda, indica la **tasa** y en cuál de las dos monedas te cobraron la **comisión** (apartado 7.4). Debajo verás lo que recibiría tu cuenta con lo tecleado.

La venta se registra como una transacción más, así que aparece en el historial y se puede editar o eliminar como cualquier otra.

---

## 11. Importación masiva de transacciones (Excel/CSV)

Si ya llevas tu registro en una hoja de cálculo, no necesitas volver a teclearlo todo: Finexia puede **importar tus transacciones desde un archivo Excel o CSV**. La propia página lo resume así: *"Sube el Excel donde llevas tu registro de inversiones: detectamos tus columnas y las adaptamos automáticamente."*

El proceso tiene tres pasos, indicados en la parte superior: **1 · Archivo**, **2 · Columnas y vista previa** y **3 · Resultado**.

![Primer paso de la importación: elegir portafolio de destino, plataforma y subir el archivo](img/manual/09-importar-transacciones.png)

### 11.1 Paso 1 — Archivo

1. Ve a **Transacciones → Importar**.
2. Selecciona el **Portafolio destino** (dónde se crearán las transacciones) y la **Plataforma / broker** a la que corresponden.
3. **Arrastra tu Excel** a la zona de carga **o haz clic para buscarlo** en tu equipo. Se admiten los formatos **.xlsx** y **.csv**, con un tamaño máximo de **8 MB**. No importa cómo se llamen tus columnas: podrás asignarlas en el siguiente paso.
4. La aplicación analizará el archivo ("Analizando tu archivo…"). Si el libro tiene varias hojas, selecciona la **hoja** que contiene las transacciones.

### 11.2 Paso 2 — Columnas y vista previa

En la pantalla **"Asigna tus columnas"** la aplicación te dice cuántas columnas detectó en tu archivo y en qué fila están los encabezados, y te propone una **asignación sugerida** que puedes ajustar.

![Paso 2 de la importación: columnas asignadas, valores por defecto y vista previa con filas válidas y con errores](img/manual/14-importar-columnas.png)

En esta pantalla:

- Si el libro tiene varias hojas, elige la **Hoja** correcta (al cambiarla, la aplicación vuelve a sugerir la asignación).
- Asigna cada dato a una columna de tu archivo. Los campos **Fecha**, **Ticker/Símbolo**, **Cantidad** y **Precio** son obligatorios (marcados con `*`); **Tipo de operación**, **Nombre del activo**, **Comisiones**, **Moneda**, **Tasa de cambio**, **Categoría** y **Notas** son opcionales — puedes dejarlos en **"— No usar —"**.
- En **Valores por defecto** defines lo que se aplicará a las filas donde tu archivo no tenga ese dato: tipo de operación, moneda, **moneda de la cuenta**, categoría y formato de fecha (con detección automática).
- La **moneda de la cuenta** es aquella en la que tu bróker debitó (apartado 7.4). Déjala vacía si tu extracto no convirtió nada; si la rellenas, cada fila que venga en otra moneda necesita su **tasa de cambio**, así que conviene mapear también esa columna. La vista previa trae una columna **Tasa** para comprobarlo fila a fila.
- La **vista previa** resume el resultado con contadores —en el ejemplo de arriba, "8 filas · 6 listas para importar · 2 con errores (se omitirán)"— y muestra las filas interpretadas: las válidas con ✓ y las que se omitirán con ✗ y el **motivo del error** en la columna **Detalle** (fecha no reconocida, precio vacío…), para que puedas corregir tu archivo si lo deseas.

### 11.3 Paso 3 — Resultado

1. Cuando la asignación sea correcta, pulsa **Importar N transacciones** (el botón indica cuántas filas válidas se crearán). Si prefieres empezar de nuevo, usa **Elegir otro archivo**.
2. Las transacciones se crearán en el portafolio elegido y verás el **resultado** del proceso: cuántas se importaron, cuántas se omitieron y el detalle de los errores, si los hubo.

> **Consejos para una buena importación:**
> - Usa una fila de encabezados clara en tu hoja.
> - Mantén formatos de fecha y número consistentes.
> - Los tickers que no estén en el catálogo se dan de alta automáticamente con el nombre, tipo y moneda del archivo, y quedan visibles solo para ti.
> - Puedes repetir la vista previa tantas veces como necesites antes de confirmar; nada se guarda hasta el paso final.

---

## 12. Reportes y exportaciones

La sección **Reportes** ("Gestiona y descarga documentos financieros de tu cuenta") concentra el análisis de rendimiento.

![Página de reportes con el calendario de rendimiento, las estadísticas clave, la proyección de crecimiento y las descargas](img/manual/10-reportes.png)

### 12.1 Reportes en pantalla

- **Rentabilidad mensual (%):** calendario con el rendimiento de cada mes, un panel por año (verás tantos paneles como años cubra tu historial). Arriba de cada panel está el **acumulado del año**, y debajo la **leyenda de la escala de color** (de rojo intenso para caídas de más del 1 % a verde intenso para subidas del 2 % o más). Los meses sin dato aparecen con un guion.

  Lo que ves es **rendimiento de lo invertido, no variación del saldo**: los aportes y retiros del mes se descuentan antes de calcular el porcentaje, así que ingresar dinero no pinta el mes en verde. El descuento se hace con tus movimientos reales, uno a uno: una compra entra como dinero puesto, una venta sale por lo que cobraste —no por lo que te costó en su día, que es lo que convertiría una plusvalía en pérdida— y un dividendo cuenta como renta que sí ganaste. Los meses que tu historial no cubre enteros van marcados con un asterisco —el mes en el que arranca y el que está en curso—, porque cubren menos días que un mes completo. Esos mismos meses quedan fuera del mejor y el peor mes de las estadísticas: tres días no se comparan con treinta.

  Un detalle a tener en cuenta al empezar: cuando registras una posición que ya tenías comprada, la aplicación la incorpora el día en que se la cuentas, no el día en que la compraste —de tu historial anterior no sabe nada—. Si esa posición llega con ganancia acumulada, esa ganancia aparece entera en su primer día dentro de la aplicación. No es un error de cálculo, es todo lo que se puede saber de un periodo que nadie observó; a partir de ahí, el seguimiento ya es día a día.

- **Estadísticas clave:** doce métricas repartidas en tres bloques.

  | Bloque | Métricas |
  |---|---|
  | **Rendimiento** | Rentabilidad del periodo, rentabilidad anualizada, ganancia / pérdida, ganancia sobre coste, mejor mes y peor mes |
  | **Riesgo** | Volatilidad anualizada, máxima caída y ratio de Sharpe |
  | **Historial** | Periodo cubierto, capital invertido y valor actual |

  Al pasar el cursor por cualquier métrica verás qué mide exactamente. Las que necesitan más historial del que tienes aparecen como **N/A** con el motivo escrito debajo. Las cifras anuales —rentabilidad anualizada y ratio de Sharpe— comparten el mismo umbral de **90 días**, porque las dos son el mismo historial llevado a un año y anualizar unas pocas semanas da cifras sin sentido.

  La volatilidad es un caso aparte y por eso cambia de nombre: medir cuánto oscila tu cartera de un día al otro necesita tramos (al menos 10), no meses, así que por debajo de los 90 días se publica como **volatilidad por tramo**, sin anualizar y con el aviso al lado. A partir de los 90 pasa a llamarse **volatilidad anualizada**. No compares una con otra: en una serie diaria se llevan un factor de veinte.

  Dos cosas que conviene tener claras al leer la tarjeta, porque las cifras parecen contradecirse y no lo hacen:

  - **La rentabilidad del periodo no se traduce en la ganancia.** La primera encadena lo que rindió tu dinero e ignora *cuándo* aportaste; la segunda es dinero contante y sí depende de eso. Un +30 % de periodo puede convivir con un +10 % sobre coste: significa que la mayor parte de tu capital entró después de la subida. Ninguna de las dos está mal, miden cosas distintas.
  - **El ratio de Sharpe no cuadra con la rentabilidad anualizada.** Se calcula por la convención de siempre (media de los tramos anualizada ÷ volatilidad), que no es la rentabilidad compuesta de la primera fila dividida entre la volatilidad. Si multiplicas el Sharpe por la volatilidad no te sale la anualizada, y no es un error. Va en gris y con un aviso al lado a propósito: es una estimación, y con pocos meses de historial su margen de error es lo bastante ancho como para que no debas leerlo como un sello de calidad.

- **Proyección de crecimiento:** estimación a cinco años calculada extrapolando **tu rentabilidad anualizada** sobre el valor actual, sin contar aportes futuros. La tasa que se extrapola va en la píldora de la cabecera (p. ej. **+12,4 % anual**), y cada año muestra dos cifras: su **valor proyectado** y, debajo, el **porcentaje acumulado** desde hoy.

  Las dos dicen lo mismo de dos formas, pero solo una es comparable. El importe depende de cuánto tengas hoy en la cuenta —un aporte de mañana lo mueve entero sin que la proyección haya cambiado de opinión—; el porcentaje es la extrapolación en crudo, y es el que puedes comparar con el de cualquier otra cartera o con lo que rinde una alternativa.

  **No es una previsión de mercado**, y por eso la aplicación se abstiene de mostrarla cuando tienes menos de seis meses de historial —en ese caso te dice cuántos días llevas y cuántos faltan— o cuando el ritmo es tan extremo que proyectarlo daría cifras absurdas.

### 12.2 Descargas en Excel (XLSX)

En la parte inferior de la página encontrarás las tarjetas de descarga; pulsa **Descargar** en la que necesites:

| Reporte | Contenido |
|---|---|
| **Resumen mensual** | Resumen mes a mes del portafolio |
| **Estado de resultados** | Historial completo de transacciones |
| **Riesgo y volatilidad** | Las métricas de la página, el detalle mensual y la serie de la que salen |

El reporte de **riesgo y volatilidad** trae tres hojas:

| Hoja | Contenido |
|---|---|
| **Métricas de riesgo** | Las mismas cifras del panel *Estadísticas clave*, cada una con su unidad y una columna que explica cómo se calcula. Una métrica que tu historial todavía no sostiene dice «Sin historial suficiente» en vez de dejar la celda vacía, que se leería como un cero |
| **Rentabilidad mensual** | El porcentaje de cada mes, el mismo del calendario, con una columna que dice si el mes está completo. Los que no lo están son el primero de tu historial y el que esté en curso |
| **Historial de crecimiento** | La serie diaria: valor, coste, ganancia, rentabilidad y el **aporte neto** de cada día. Con esa última columna puedes rehacer cualquiera de las métricas por tu cuenta y comprobar que cuadran |

Los archivos se descargan directamente a tu equipo en formato Excel (XLSX) y puedes abrirlos con Excel, LibreOffice o Google Sheets.

---

## 13. Notificaciones

![Preferencias de notificaciones por correo electrónico y alertas en la app](img/manual/11-notificaciones.png)

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

![Página de configuración con el perfil, la apariencia, la seguridad y las sesiones activas](img/manual/12-configuracion.png)

### 14.1 Perfil

- Edita tu **nombre** y datos personales.
- **Sube o cambia tu avatar** (imagen de perfil), que se mostrará en la cabecera y en tu perfil.
- Elige tu **moneda preferida**: es en la que se muestran los totales del Dashboard y de Mis Activos. Solo aparecen las monedas a las que la aplicación sabe convertir.
- Tu correo electrónico identifica tu cuenta.

### 14.2 Apariencia

Ajusta las preferencias visuales de la aplicación (tema/aspecto de la interfaz) a tu gusto. Los cambios se aplican de inmediato y quedan guardados en tus preferencias.

### 14.3 Seguridad

- **Cambiar contraseña:** introduce tu contraseña actual y la nueva. Usa contraseñas largas y únicas.
- **Verificación en dos pasos (2FA)** y **Sesiones activas:** ver sección 15.

### 14.4 Datos de mercado (tu clave de proveedor)

Aquí es donde das de alta la clave con la que Finexia consulta los precios de tus activos. Sin ella tus posiciones se valoran a precio de compra (sección 7.3).

**Cómo obtener una clave.** Ambos proveedores tienen plan gratuito y basta con registrarse:

| Proveedor | Dónde | Plan gratuito |
|---|---|---|
| **Finnhub** | finnhub.io | ~60 consultas por minuto; es el recomendado |
| **Alpha Vantage** | alphavantage.co | Más limitado (unas pocas consultas por minuto), útil como respaldo |

Puedes configurar las dos. Finexia intenta primero Finnhub y recurre a Alpha Vantage si la primera no cubre un activo.

**Cómo configurarla.** Pega la clave en el campo del proveedor y pulsa **Guardar**. Finexia la comprueba contra el proveedor antes de aceptarla, así que si te equivocas al copiarla te lo dice en el momento en vez de dejarte con una clave que no funciona.

Por cada clave guardada verás:

- Los **cuatro últimos caracteres**, para saber cuál tienes puesta.
- Su **estado**: *activa*, *sin cuota* (has agotado el límite del día; mañana vuelve sola) o *no válida* (el proveedor la rechazó, normalmente porque la revocaste).
- **Verificar**, que vuelve a comprobarla, y **Eliminar**, que la borra.

**Sincronizar** pide los precios de tus posiciones en ese momento, sin esperar a la actualización diaria. Con una clave de Alpha Vantage el proceso va despacio a propósito, para no agotar tu cuota: si tienes muchas posiciones puede que no le dé tiempo a todas y las restantes se actualicen en la sincronización de la noche.

> **La clave no se puede volver a leer, ni siquiera tú.** Se guarda cifrada y Finexia nunca la devuelve: por eso el campo aparece siempre vacío y cambiarla significa escribirla entera otra vez. Si la pierdes, genera una nueva en el proveedor. Si prefieres dejar de usarla, **Eliminar** la borra de verdad.

### 14.5 Acceso para asistentes y aplicaciones conectadas

En la parte baja de la página están las credenciales con las que un asistente de IA puede leer tus carteras: **Acceso para asistentes (MCP)**, donde creas, rotas y eliminas tus tokens, y **Aplicaciones conectadas**, donde retiras el acceso a las aplicaciones que hayas autorizado. Las dos se explican enteras en la sección 16.

---

## 15. Seguridad: 2FA y sesiones

### 15.1 Activar la verificación en dos pasos (2FA)

La 2FA añade una segunda barrera al inicio de sesión: además de tu contraseña, necesitarás un código temporal generado por tu aplicación de autenticación.

1. Ve a **Configuración → Verificación en dos pasos (2FA)** y pulsa **Activar**.
2. **Escanea el código QR** con tu aplicación de autenticación (Google Authenticator, Authy, 1Password, etc.). Si no puedes escanear, usa la **"Clave para ingreso manual"** que se muestra junto al QR.
3. Introduce el **código de 6 dígitos** que genera la aplicación para confirmar la activación.
4. **Guarda tus códigos de recuperación.** La aplicación te mostrará una lista de códigos de un solo uso: descárgalos o cópialos y guárdalos en un lugar seguro (gestor de contraseñas, papel en lugar protegido). Son tu única vía de acceso si pierdes el teléfono.

A partir de ese momento, cada inicio de sesión pedirá el código temporal.

### 15.2 Códigos de recuperación y desactivación

- **Usar un código de recuperación:** en el segundo paso del login, introduce uno de tus códigos de recuperación en lugar del código temporal. Cada código sirve **una sola vez**.
- **Regenerar códigos:** desde la sección 2FA puedes generar una nueva lista (la anterior queda invalidada).
- **Desactivar 2FA:** desde la misma sección, confirmando con un código válido. Tu cuenta volverá a protegerse solo con contraseña (no recomendado).

### 15.3 Sesiones activas

En **Configuración → Sesiones activas** verás todos los dispositivos/navegadores con sesión abierta en tu cuenta, con información para identificarlos (dispositivo, ubicación aproximada y última actividad).

- **Revocar una sesión:** cierra la sesión de un dispositivo concreto.
- **Cerrar las demás sesiones:** cierra todas las sesiones excepto la actual. Útil si sospechas que alguien más accedió a tu cuenta (en ese caso, cambia también tu contraseña).

---

## 16. Conectar un asistente de IA (MCP)

Finexia también puede responder a través de un **asistente de IA**: preguntarle a Claude —o a cualquier cliente que hable **MCP** (*Model Context Protocol*)— cuánto tienes en cripto, qué compraste el mes pasado o cómo está repartida tu cartera, y que conteste con **tus** datos.

Lo primero, porque es lo que importa: **el acceso es de solo lectura**. Un asistente conectado puede consultar tus carteras; no puede crear, modificar ni borrar nada, ni entrar al resto de la aplicación. Y llega solo hasta donde tú le dejes: la conexión se corta desde **Configuración** cuando quieras.

Hay dos formas de conectarlo, según lo que sepa hacer tu cliente:

| Forma | Cuándo usarla |
|---|---|
| **Token personal** | El cliente pide una dirección y una cabecera de autorización. Es el caso de Claude Desktop, Claude Code y la mayoría |
| **Autorización (OAuth)** | El cliente sabe conectarse solo: te lleva a una pantalla de Finexia donde apruebas la conexión |

### 16.1 Crear un token

En **Configuración → Acceso para asistentes (MCP)**:

1. Escribe un **nombre** que te diga después cuál es cuál: "Claude Desktop", "Portátil del trabajo".
2. Elige la **caducidad**: 30 días, 90 días (lo recomendado), 1 año o sin caducidad.
3. Pulsa **Crear token**.

El **secreto se muestra una sola vez**, justo después de crearlo, con un botón para copiarlo y un ejemplo de configuración listo para pegar. Finexia guarda solo su huella, así que nadie —tú incluido— puede volver a leerlo: si lo pierdes, rota el token y reconfigura el cliente.

Cada token de la lista muestra su nombre, sus **cuatro últimos caracteres**, su estado (**Activo**, **Sin usar** mientras nadie lo haya usado todavía, o **Caducado**), cuándo se usó por última vez, cuándo caduca y cuándo se creó. Y dos acciones:

- **Rotar** — cambia el secreto por uno nuevo y lo enseña una vez. El anterior deja de funcionar en el acto: es lo que hay que hacer si crees que se filtró.
- **Eliminar** — revoca el token para siempre.

### 16.2 Configurar el cliente

La página de ajustes muestra la **dirección del endpoint** que hay que pegar en tu cliente. Con esa dirección y el token, la configuración de un cliente MCP tiene esta forma:

```json
{
  "mcpServers": {
    "finexia": {
      "type": "http",
      "url": "https://api.finexia.me/mcp",
      "headers": { "Authorization": "Bearer fnx_mcp_…" }
    }
  }
}
```

En Claude Code, la misma conexión se añade con una orden:

```
claude mcp add --transport http finexia https://api.finexia.me/mcp \
  --header "Authorization: Bearer fnx_mcp_…"
```

> **Trata el token como una contraseña.** Da acceso de lectura a todas tus carteras. No lo pegues en un chat ni lo subas a un repositorio y, si tienes dudas, rótalo: no cuesta nada y el anterior queda inservible.

### 16.3 Autorizar una aplicación (OAuth)

Los clientes que saben conectarse solos no necesitan que copies nada: al añadir Finexia te llevan a una pantalla de la propia aplicación donde se te pide permiso.

Esa pantalla dice **qué aplicación** lo pide, **qué está pidiendo** —leer tus carteras, posiciones, transacciones y datos de mercado, sin poder cambiar nada— y **a dónde volverá** al terminar. Antes de pulsar **Autorizar**, comprueba dos cosas:

- Que el nombre sea el de la aplicación desde la que **acabas de pedir** la conexión. Si la pantalla aparece sin que tú hayas hecho nada, cancela.
- Que la dirección de retorno sea la que esperas de ella. El nombre y el logotipo los eligió quien registró el cliente, y registrarse es abierto: son texto de un desconocido. La dirección de retorno es el dato que sí se puede contrastar.

**Cancelar** no conecta nada y no tiene ninguna consecuencia.

### 16.4 Retirar el acceso

En **Configuración → Aplicaciones conectadas** están las aplicaciones que autorizaste, con los permisos que tienen, cuándo las conectaste y cuándo las usaron por última vez. **Desconectar** corta el acceso al instante, y es la única forma de retirárselo: mientras nadie lo corte, el cliente renueva su permiso solo.

Los tokens personales se revocan desde la sección de al lado, con **Eliminar** (apartado 16.1).

### 16.5 Qué puede consultar un asistente

Con la conexión hecha, el asistente puede preguntar por:

- Tus **portafolios**, con su valor, su coste y su ganancia.
- Lo que tienes de cada **activo**, sumado entre portafolios.
- Tu **asignación** por tipo de activo.
- La **evolución** de un portafolio a lo largo del tiempo.
- Tus **transacciones recientes**.
- Tus **plataformas**, el **catálogo de activos** y las **tasas de cambio**.

Dos cosas que conviene saber para leer sus respuestas:

- Los importes van **cada uno con su moneda**, y dos cifras en monedas distintas no se comparan sin convertirlas antes.
- Una posición sin precio de mercado (apartado 7.3) se valora a lo que costó, así que su ganancia es exactamente cero. El asistente recibe de dónde salió cada precio, y ese cero no es un rendimiento.

---

## 17. Preguntas frecuentes (FAQ)

**¿Finexia se conecta a mis brokers o plataformas?**
No. Finexia nunca accede a tus plataformas ni te pide credenciales. Tú registras manualmente dónde tienes tus activos, así que la información siempre está bajo tu control.

**¿Puedo tener varios portafolios?**
Sí, puedes crear tantos como necesites, cada uno con su moneda, tipo, nivel de riesgo y monto objetivo propios.

**¿Cómo se calculan los valores de mis posiciones?**
Cantidad × precio, convertido a la moneda del portafolio. El precio es, por este orden: el último que trajo **tu** clave de datos de mercado; si no tienes clave, el precio manual que haya fijado un administrador para ese activo; y si tampoco lo hay, tu propio precio de compra. Las tasas de cambio siguen la misma regla. Ver secciones 7.3 y 14.4.

**¿Puedo importar mi histórico desde Excel?**
Sí. Usa **Transacciones → Importar**: sube el archivo (.xlsx o .csv, máximo 8 MB), asigna las columnas, revisa la vista previa (incluidas las filas omitidas) y confirma. Nada se guarda hasta que confirmas.

**No encuentro un activo al crear una posición. ¿Qué hago?**
Créalo tú desde el propio buscador: pulsa **Crear "TICKER"**, indica nombre, tipo y moneda, y podrás seguir con la posición sin salir de la pantalla. Solo lo verás tú mientras el equipo de Finexia no lo incorpore al catálogo general.

**¿Qué pasa si pierdo mi teléfono con la app de autenticación?**
Usa uno de tus **códigos de recuperación** para entrar y luego reconfigura la 2FA. Si tampoco tienes los códigos, contacta con el soporte de Finexia.

**¿Puedo usar Finexia en el móvil?**
Sí. La interfaz es adaptable; en pantallas pequeñas el menú lateral se abre desde el botón de la cabecera (ver sección 4.4).

**¿Cómo exporto mis datos?**
Desde **Reportes** puedes descargar en Excel el resumen mensual, el estado de resultados (transacciones) y las métricas de riesgo y volatilidad.

**¿Dónde encuentro este manual?**
En **Guía de usuario**, en el menú lateral. Puedes leerlo dentro de la aplicación, abrirlo en una pestaña nueva o descargar el PDF (ver sección 4.5).

**¿Puedo usar la aplicación solo con el teclado?**
Sí. Los menús, formularios y tablas se recorren con **Tab**, y la gráfica de crecimiento del Dashboard admite las flechas para ir día a día (sección 5.1).

**¿Dónde veo cuánto tengo de un activo en total?**
En **Mis Activos**, en el menú lateral: una fila por activo con la cantidad sumada de todos tus portafolios, lo que vale y cuánto pesa sobre el total, más un buscador para dar con uno concreto (sección 8).

**Compré en una moneda y mi cuenta está en otra. ¿Cómo lo registro?**
Marca **"Mi cuenta liquidó en otra moneda"** en el formulario e indica la moneda de tu cuenta y la tasa de ese día, la de la confirmación del bróker (apartado 7.4). Lo mismo vale al vender y al registrar transacciones sueltas.

**¿Puedo eliminar una posición o una transacción?**
Sí. Una transacción se elimina desde su fila en el historial; la posición entera, desde **Eliminar posición**, al final de la ficha del activo, y se lleva todas sus transacciones con ella (apartado 7.2).

**¿Puedo conectar Finexia con Claude u otro asistente de IA?**
Sí, en modo **solo lectura**: crea un token en **Configuración → Acceso para asistentes (MCP)** y pégalo en tu cliente, o autoriza la conexión desde el propio asistente si sabe hacerlo por sí mismo. Podrá consultar tus carteras, nunca modificarlas (sección 16).

---

## 18. Solución de problemas

| Problema | Causa probable | Solución |
|---|---|---|
| "Correo sin verificar" al iniciar sesión | No abriste el enlace de verificación | Reenvía el correo de verificación desde la propia pantalla y revisa spam |
| "Demasiados intentos" (error 429) | Límite de peticiones por seguridad | Espera unos minutos y vuelve a intentarlo |
| El enlace de invitación o de restablecimiento no funciona | Enlace caducado o ya usado | Solicita un nuevo enlace o el reenvío de la invitación |
| No puedo registrarme | El registro directo está deshabilitado | Únete a la lista de espera para recibir una invitación |
| El código 2FA no es aceptado | Reloj del teléfono desincronizado o código expirado | Sincroniza la hora del dispositivo y usa el código vigente; como alternativa, un código de recuperación |
| Mi sesión se cerró sola | La sesión fue revocada o expiró | Inicia sesión de nuevo; revisa **Sesiones activas** si no fuiste tú |
| Una importación omite filas | Datos incompletos o formatos no interpretables | Revisa el detalle de **Filas omitidas**, corrige el archivo y repite la vista previa |
| Los valores no cuadran con mi broker | No tienes clave de datos de mercado (se valora a coste), faltan transacciones, o el precio aún no se ha refrescado | Configura tu clave en **Configuración → Datos de mercado** y pulsa **Sincronizar**; completa el historial de transacciones |
| Mi ganancia/pérdida sale exactamente 0 | Sin clave de datos de mercado se valora a precio de compra, así que no hay nada que comparar | Configura tu clave (sección 14.4) |
| El estado de mi clave aparece como "sin cuota" | Has agotado el límite diario o por minuto de tu plan gratuito | No hace falta hacer nada: se reintenta en la siguiente sincronización |
| No puedo eliminar una plataforma | Todavía tiene posiciones registradas, incluidas las que ya vendiste | Elimina primero esas posiciones y vuelve a intentarlo (apartado 9.4) |
| El total de **Mis Activos** mezcla monedas | Falta la tasa de cambio de la moneda de algún activo | Es un aviso, no un error: la página dice cuántos activos van sin convertir. Espera a que haya tasa o lee ese importe en su propia moneda |
| Una compra en otra moneda no cuadra con lo que me debitaron | La tasa registrada no es la del día de la operación | Edita la transacción y pon la tasa de la confirmación del bróker (apartado 7.4) |
| Perdí el token de mi asistente | El secreto solo se muestra al crearlo o al rotarlo | Pulsa **Rotar** en ese token y reconfigura el cliente con el nuevo (apartado 16.1) |
| Mi asistente dice que no tiene acceso | El token caducó o se eliminó, o desconectaste la aplicación | Revisa el estado del token en **Configuración**; crea uno nuevo, o vuelve a autorizar la aplicación (sección 16) |

Si el problema persiste, contacta con el equipo de soporte de Finexia.

---

## 19. Glosario

| Término | Definición |
|---|---|
| **Activo** | Instrumento financiero: acción, criptomoneda, ETF, fondo, etc., identificado por su símbolo/ticker |
| **Posición** | Tenencia de un activo dentro de un portafolio (cantidad + coste) |
| **Portafolio** | Conjunto de posiciones agrupadas bajo un objetivo, con moneda y nivel de riesgo propios |
| **Plataforma** | Broker, exchange o entidad donde custodias tus activos |
| **Transacción** | Operación de compra o venta que modifica una posición |
| **Asignación** | Porcentaje que representa un activo o tipo de activo dentro del total |
| **ROI** | *Return on Investment*: retorno sobre la inversión, en porcentaje |
| **Máxima caída** | *Max drawdown*: la mayor bajada porcentual de la rentabilidad acumulada desde un máximo anterior |
| **CAGR** | *Compound Annual Growth Rate*: ritmo de crecimiento anual compuesto |
| **Volatilidad** | Medida de cuánto oscila la rentabilidad; se publica anualizada a partir de 90 días de historial y por tramo mientras tanto |
| **Ratio de Sharpe** | Rentabilidad obtenida por cada unidad de riesgo asumida (con tasa libre de riesgo 0); por encima de 1 se considera bueno, aunque con poco historial la cifra es una estimación de margen ancho |
| **Rentabilidad del periodo** | Lo que rindió el dinero invertido a lo largo de tu historial, descontando aportes y retiros |
| **Rentabilidad real** | La misma cuenta aplicada al tramo que estás viendo (gráfica de crecimiento) o a un portafolio concreto; se mueve con el mercado, no con tus depósitos |
| **Ganancia sobre coste** | Lo que vale tu cartera frente a lo que te costó; sí depende de cuándo entró cada aporte, y por eso no coincide con la rentabilidad real |
| **2FA / TOTP** | Verificación en dos pasos con códigos temporales de 6 dígitos |
| **Códigos de recuperación** | Códigos de un solo uso para acceder si pierdes tu aplicación de autenticación |
| **Lista de espera** | Registro público para solicitar acceso anticipado a la plataforma |
| **Invitación** | Enlace enviado a tu correo para crear una cuenta |
| **Moneda de la operación** | Aquella en la que se ejecutó la compra o la venta: la del mercado donde cotiza el activo |
| **Moneda de la cuenta** | Aquella en la que tu bróker debitó o abonó el importe, y en la que queda el coste de la posición |
| **Tasa de la operación** | Cuántas unidades de la moneda de la cuenta costaba una de la moneda de la operación ese día |
| **Concentración** | Cuánto pesa un activo sobre todo lo que tienes, sin importar en qué portafolio esté |
| **MCP** | *Model Context Protocol*: el estándar por el que un asistente de IA se conecta a una aplicación para consultar sus datos |
| **Token de acceso (MCP)** | Credencial que creas en Configuración para que un cliente MCP lea tus carteras; no sirve para nada más |
| **OAuth** | Forma de autorizar una aplicación externa sin darle tu contraseña: apruebas la conexión en una pantalla de Finexia y la retiras cuando quieras |
| **XLSX** | Formato de archivo de Excel usado en las exportaciones |

---

*Este manual describe la funcionalidad de Finexia a la fecha indicada en la portada. Las pantallas y textos pueden variar ligeramente según la versión desplegada.*

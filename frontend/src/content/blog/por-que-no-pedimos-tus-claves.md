---
title: Por qué no te pedimos las claves de tus cuentas
description: Conectar cuentas es cómodo y por eso casi todos lo hacen. Esto es lo que cuesta, y por qué preferimos que lo escribas tú.
date: 2026-09-17
tags: [seguridad]
draft: true
---

Cada vez que alguien prueba Finexia aparece la misma pregunta, y es una buena pregunta: «¿por qué
tengo que escribirlo yo, si otras apps se conectan solas?».

La respuesta corta es que sí, se conectan solas, y que para hacerlo necesitan algo que nosotros
preferimos no tener.

## Qué significa «conectar tu cuenta»

Cuando una aplicación lee el saldo de tu banco o las posiciones de tu broker, por debajo está
pasando una de dos cosas.

La primera es que le das tu usuario y tu contraseña, y un servicio intermedio entra a tu cuenta
como si fueras tú, cada día, para copiar lo que ve. Suena mal porque lo es: esas credenciales tienen
que quedar guardadas en algún sitio en una forma que permita volver a usarlas.

La segunda es más limpia: la plataforma emite un **token** de solo lectura, tú autorizas, y la app
usa ese token. Nadie guarda tu contraseña. Pero el token sigue siendo una llave a tus datos
financieros, sigue viviendo en un servidor que no es el tuyo, y sigue estando ahí el día que ese
servidor tenga un mal día.

Las dos opciones funcionan. Las dos crean algo que antes no existía: **un secreto tuyo, fuera de tu
alcance, que a alguien le sirve.**

## La cuenta que hicimos

Un agregador con cien mil usuarios conectados guarda cien mil llaves. Eso lo convierte en un
objetivo, y el valor de reventarlo no baja nunca. No hace falta suponer mala fe de nadie: basta con
una dependencia comprometida, un backup mal configurado o un empleado con más permisos de los que
necesitaba.

Nosotros somos un producto nuevo, con un equipo pequeño. Podríamos construir esa bóveda y
defenderla, y probablemente nos iría bien durante bastante tiempo. Pero la forma más barata de no
filtrar un secreto sigue siendo **no tenerlo**.

Así que no lo tenemos. Finexia no se conecta a tu broker, a tu exchange ni a tu banco. No hay
credenciales de terceros en nuestra base de datos, porque nunca te las pedimos. No hay nada que
filtrar de esa gaveta, porque la gaveta no existe.

## Lo que sí cuesta

Sería deshonesto vender esto como si fuera gratis. Tiene un precio y lo pagas tú:

- El primer día hay que sentarse a registrar lo que tienes.
- Cuando compras o vendes, el movimiento lo apuntas tú.
- Si te olvidas de apuntar algo, el panel se queda desactualizado y no hay nadie que lo note por ti.

Ese es el trato. Más trabajo tuyo, a cambio de que no exista una llave a tus cuentas guardada en un
servidor ajeno.

## Cómo lo hacemos menos pesado

Que la decisión sea deliberada no significa que el trabajo tenga que doler.

- **Importación de extractos.** Descargas el CSV o el Excel de tu plataforma, lo subes, y Finexia te
  enseña una vista previa de lo que va a crear antes de confirmar nada. La carga inicial deja de ser
  una tarde y pasa a ser un rato.
- **Precios al día.** Las posiciones se valoran con datos de mercado, así que lo que tú apuntas es
  _qué_ tienes y _cuánto_ pagaste, no cuánto vale hoy.
- **Todo sale contigo.** Tu resumen, tus movimientos y tu informe de riesgo se descargan en XLSX
  cuando quieras. No hay que pedir permiso ni escribir a soporte.

## Y lo que sí guardamos

Guardamos lo que hace falta para que la aplicación funcione: tu correo, tu nombre y los datos de
portafolios y activos que tú registras. Nada más. No vendemos datos personales y no los usamos para
otra cosa.

Está escrito, con lo que la ley colombiana te garantiza como titular, en la
[Política de Tratamiento de Datos Personales](/privacidad).

Sobre esa base —no tener tus claves— está construido lo demás: verificación en dos pasos, control de
las sesiones abiertas, y un botón para tapar todos los importes cuando abres el panel donde te puede
ver alguien.

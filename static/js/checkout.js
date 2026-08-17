/* ============================================================
   CHECKOUT.HTML - Lógica del formulario de pago y solicitud OTP
   ============================================================ */

console.log("[Checkout] checkout.js se cargó correctamente.");

document.addEventListener("DOMContentLoaded", function () {
    inicializarCheckout();
});

let temporizadorReenvioOtp = null;

function inicializarCheckout() {
    const formulario = document.getElementById("formulario-pago-otp");
    const tabsMetodo = document.getElementById("tabs-tipo-cuenta");
    const overlayConfirmacion = document.getElementById("checkout-overlay-confirmacion");
    const botonCancelarModal = document.getElementById("checkout-boton-cancelar-modal");
    const botonConfirmarModal = document.getElementById("checkout-boton-confirmar-modal");

    if (!formulario || !tabsMetodo) {
        console.error("[Checkout] Faltan elementos del formulario en el DOM.");
        return;
    }

    // Alternar entre "Teléfono" (CELE) y "Cuenta" (CNTA)
    tabsMetodo.addEventListener("click", function (evento) {
        const pestañaClickeada = evento.target.closest(".checkout-tab-btn");
        if (!pestañaClickeada) return;

        document.querySelectorAll(".checkout-tab-btn").forEach(function (pestaña) {
            pestaña.classList.remove("activo");
        });
        pestañaClickeada.classList.add("activo");

        const tipoSeleccionado = pestañaClickeada.dataset.tipo;
        document.getElementById("campo-tipo-cuenta").value = tipoSeleccionado;

        const etiquetaNumero = document.getElementById("etiqueta-numero-cuenta");
        const campoNumero = document.getElementById("campo-numero-cuenta");

        if (tipoSeleccionado === "CELE") {
            etiquetaNumero.textContent = "Teléfono";
            campoNumero.placeholder = "04141234567";
        } else {
            etiquetaNumero.textContent = "Número de Cuenta";
            campoNumero.placeholder = "20 dígitos, inicia con el código del banco";
        }
    });

    // Al enviar el formulario: validar y abrir el modal de confirmación (todavía no llamamos al backend)
    formulario.addEventListener("submit", function (evento) {
        evento.preventDefault();

        const errorValidacion = validarFormularioPago();
        if (errorValidacion) {
            mostrarErrorFormulario(errorValidacion);
            return;
        }

        mostrarErrorFormulario("");
        abrirModalConfirmacion();
    });

    botonCancelarModal.addEventListener("click", cerrarModalConfirmacion);

    overlayConfirmacion.addEventListener("click", function (evento) {
        if (evento.target === overlayConfirmacion) {
            cerrarModalConfirmacion();
        }
    });

    botonConfirmarModal.addEventListener("click", function () {
        cerrarModalConfirmacion();
        solicitarCodigoOtp();
    });

    // Las casillas del OTP avanzan automáticamente al siguiente input
    const casillasOtp = document.querySelectorAll("#checkout-otp-casillas input");
    casillasOtp.forEach(function (casilla, indice) {
        casilla.addEventListener("input", function () {
            casilla.value = casilla.value.replace(/[^0-9]/g, "");
            if (casilla.value && indice < casillasOtp.length - 1) {
                casillasOtp[indice + 1].focus();
            }
        });

        casilla.addEventListener("keydown", function (evento) {
            if (evento.key === "Backspace" && !casilla.value && indice > 0) {
                casillasOtp[indice - 1].focus();
            }
        });
    });

    const botonConfirmarOtp = document.getElementById("checkout-boton-confirmar-otp");
    if (botonConfirmarOtp) {
        botonConfirmarOtp.addEventListener("click", confirmarCodigoOtp);
    }

    const botonReenviarOtp = document.getElementById("checkout-boton-reenviar-otp");
    if (botonReenviarOtp) {
        botonReenviarOtp.addEventListener("click", reenviarCodigoOtp);
    }
}

/* ------------------------------------------------------------
   VALIDACIÓN BÁSICA DEL FORMULARIO (del lado del cliente,
   el backend igual debe validar todo de nuevo)
------------------------------------------------------------- */
function validarFormularioPago() {
    const banco = document.getElementById("campo-banco").value;
    const tipoCuenta = document.getElementById("campo-tipo-cuenta").value;
    const numeroCuenta = document.getElementById("campo-numero-cuenta").value.trim();
    const tipoDocumento = document.getElementById("campo-tipo-documento").value;
    const numeroDocumento = document.getElementById("campo-numero-documento").value.trim();

    if (!banco) {
        return "Selecciona tu banco.";
    }

    if (!numeroDocumento || numeroDocumento.length < 5) {
        return "Ingresa un número de documento válido.";
    }

    if (tipoCuenta === "CELE") {
        const prefijosValidos = ["0424", "0426", "0412", "0416", "0414", "0422"];
        const prefijoIngresado = numeroCuenta.substring(0, 4);
        if (numeroCuenta.length !== 11 || !prefijosValidos.includes(prefijoIngresado)) {
            return "Ingresa un número de teléfono venezolano válido (ej. 04141234567).";
        }
    } else {
        if (numeroCuenta.length !== 20 || !/^\d+$/.test(numeroCuenta)) {
            return "El número de cuenta debe tener exactamente 20 dígitos.";
        }
        if (numeroCuenta.substring(0, 4) !== banco) {
            return "Los primeros 4 dígitos de la cuenta deben coincidir con el código del banco seleccionado.";
        }
    }

    return null; // sin errores
}

function mostrarErrorFormulario(mensaje) {
    const bloqueOtpVisible = document.getElementById("checkout-bloque-otp").style.display === "block";
    const idElementoError = bloqueOtpVisible ? "checkout-otp-mensaje-error" : "checkout-mensaje-error";
    const elementoError = document.getElementById(idElementoError);
    if (elementoError) {
        elementoError.textContent = mensaje;
    }
}

/* ------------------------------------------------------------
   MODAL DE CONFIRMACIÓN
------------------------------------------------------------- */
function abrirModalConfirmacion() {
    const totalTexto = document.getElementById("checkout-total-monto").textContent;
    document.getElementById("checkout-modal-texto-monto").textContent =
        "¿Confirmas la compra por " + totalTexto + "?";
    document.getElementById("checkout-overlay-confirmacion").style.display = "flex";
}

function cerrarModalConfirmacion() {
    document.getElementById("checkout-overlay-confirmacion").style.display = "none";
}

/* ------------------------------------------------------------
   SOLICITAR CÓDIGO OTP AL BACKEND
------------------------------------------------------------- */
async function solicitarCodigoOtp() {
    const idTransaccion = document.getElementById("id-transaccion").value;
    const botonPagar = document.getElementById("checkout-boton-pagar");

    const datosFormulario = {
        tipo_cuenta: document.getElementById("campo-tipo-cuenta").value,
        codigo_banco: document.getElementById("campo-banco").value,
        numero_cuenta: document.getElementById("campo-numero-cuenta").value.trim(),
        tipo_documento: document.getElementById("campo-tipo-documento").value,
        numero_documento: document.getElementById("campo-numero-documento").value.trim()
    };

    if (botonPagar) {
        botonPagar.disabled = true;
        botonPagar.textContent = "Solicitando código...";
    }

    try {
        const respuesta = await fetch("/api/checkout/" + idTransaccion + "/otp", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(datosFormulario)
        });

        if (!respuesta.ok) {
            const detalleError = await respuesta.json().catch(function () { return {}; });
            throw new Error(detalleError.error || "No se pudo solicitar el código OTP.");
        }

        console.log("[Checkout] OTP solicitado correctamente.");
        mostrarBloqueOtp();
    } catch (error) {
        console.error("[Checkout] Error al solicitar OTP:", error);
        mostrarErrorFormulario(error.message);

        if (botonPagar) {
            botonPagar.disabled = false;
            botonPagar.textContent = "Pagar";
        }
    }
}

/* ------------------------------------------------------------
   MOSTRAR EL BLOQUE DE CASILLAS OTP + CONTADOR DE REENVÍO
------------------------------------------------------------- */
function mostrarBloqueOtp() {
    document.getElementById("formulario-pago-otp").style.display = "none";
    document.getElementById("checkout-bloque-otp").style.display = "block";

    const primeraCasilla = document.querySelector("#checkout-otp-casillas input");
    if (primeraCasilla) {
        primeraCasilla.focus();
    }

    iniciarContadorReenvio(300); // 300 segundos, igual que el "expiration" que Sypago maneja
}

/* ------------------------------------------------------------
   CONFIRMAR EL CÓDIGO OTP QUE EL USUARIO ESCRIBIÓ
   (esto solo ENVÍA la solicitud - el resultado real se sabe
   con el polling de más abajo)
------------------------------------------------------------- */
let intervaloPolling = null;
let intentosPolling = 0;
const MAXIMO_INTENTOS_POLLING = 60; // 60 x 2.5s = 150 segundos como techo

async function confirmarCodigoOtp() {
    const idTransaccion = document.getElementById("id-transaccion").value;
    const botonConfirmar = document.getElementById("checkout-boton-confirmar-otp");
    const elementoErrorOtp = document.getElementById("checkout-otp-mensaje-error");

    const casillasOtp = document.querySelectorAll("#checkout-otp-casillas input");
    const codigoOtp = Array.from(casillasOtp).map(function (casilla) {
        return casilla.value;
    }).join("");

    if (codigoOtp.length !== casillasOtp.length) {
        elementoErrorOtp.textContent = "Completa las " + casillasOtp.length + " casillas del código.";
        return;
    }

    elementoErrorOtp.textContent = "";
    botonConfirmar.disabled = true;
    botonConfirmar.textContent = "Enviando...";

    try {
        const respuesta = await fetch("/api/checkout/" + idTransaccion + "/confirmar-otp", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ otp: codigoOtp })
        });

        if (!respuesta.ok) {
            const detalleError = await respuesta.json().catch(function () { return {}; });
            throw new Error(detalleError.error || "No se pudo enviar la solicitud de pago.");
        }

        // OJO: esto NO es éxito todavía. Solo significa que Sypago aceptó
        // procesar la solicitud. El resultado real llega por polling.
        console.log("[Checkout] Solicitud de pago enviada, iniciando verificación de estado...");
        mostrarBloqueProcesando();
        intentosPolling = 0;
        intervaloPolling = setInterval(function () {
            consultarEstadoTransaccion(idTransaccion);
        }, 2500);
    } catch (error) {
        console.error("[Checkout] Error al enviar la solicitud de pago:", error);
        elementoErrorOtp.textContent = error.message;
        botonConfirmar.disabled = false;
        botonConfirmar.textContent = "Confirmar";
    }
}

function mostrarBloqueProcesando() {
    document.getElementById("checkout-bloque-otp").style.display = "none";
    document.getElementById("checkout-bloque-procesando").style.display = "block";
}

/* ------------------------------------------------------------
   POLLING: consulta el estado real cada 2.5 segundos hasta
   obtener un estado definitivo (ACCP, RJCT o CANC).
------------------------------------------------------------- */
async function consultarEstadoTransaccion(idTransaccion) {
    intentosPolling++;

    if (intentosPolling > MAXIMO_INTENTOS_POLLING) {
        detenerPolling();
        mostrarResultadoFinal("rechazo", "Esto está tardando más de lo esperado", "No pudimos confirmar tu pago a tiempo. Si el banco te descontó el monto, contáctanos con tu número de referencia.");
        return;
    }

    try {
        const respuesta = await fetch("/api/checkout/" + idTransaccion + "/estado");

        if (!respuesta.ok) {
            const detalleError = await respuesta.json().catch(function () { return {}; });
            console.error("[Checkout] Error al consultar estado:", detalleError.error);
            return; // seguimos intentando en el próximo tick, no cortamos el polling por un error puntual
        }

        const datosEstado = await respuesta.json();
        console.log("[Checkout] Estado actual:", datosEstado.estado, datosEstado);

        if (datosEstado.estado === "ACCP") {
            detenerPolling();
            sessionStorage.removeItem("carritoSypago"); // limpiamos el carrito, la compra fue exitosa
            mostrarResultadoFinal("exito", "¡Pago confirmado!", "Tu compra fue procesada correctamente. Referencia: " + datosEstado.transaction_id);
        } else if (datosEstado.estado === "RJCT" || datosEstado.estado === "CANC") {
            detenerPolling();
            const motivo = datosEstado.descripcion || "El banco rechazó la operación.";
            console.error("[Checkout] Pago rechazado. Código:", datosEstado.codigo_rechazo, "-", motivo);
            mostrarResultadoFinal("rechazo", "No se pudo procesar tu pago", motivo);
        }
        // Si sigue en PEND o PROC, no hacemos nada: seguimos esperando el próximo tick.
    } catch (error) {
        console.error("[Checkout] Error de red al consultar estado:", error);
        // no detenemos el polling por un error de red puntual
    }
}

function detenerPolling() {
    if (intervaloPolling) {
        clearInterval(intervaloPolling);
        intervaloPolling = null;
    }
}

function mostrarResultadoFinal(tipo, titulo, mensaje) {
    document.getElementById("checkout-bloque-procesando").style.display = "none";

    const bloqueResultado = document.getElementById("checkout-bloque-resultado");
    bloqueResultado.className = "checkout-bloque-resultado " + tipo;
    bloqueResultado.innerHTML = "<h2>" + titulo + "</h2><p>" + mensaje + "</p>";
    bloqueResultado.style.display = "block";

    // Si la compra fue exitosa, "Regresar" ya no debe llevar de vuelta a
    // productos.html (ahí no hay nada más que hacer con esta compra),
    // sino al inicio para arrancar un nuevo proceso desde cero.
    if (tipo === "exito") {
        const botonRegresar = document.getElementById("checkout-boton-regresar");
        if (botonRegresar) {
            botonRegresar.href = "/app";
            botonRegresar.innerHTML = '<i class="fa-solid fa-arrow-left"></i> Volver al inicio';
        }
    }
}

function iniciarContadorReenvio(segundosIniciales) {
    let segundosRestantes = segundosIniciales;
    const elementoContador = document.getElementById("checkout-otp-contador");
    const botonReenviarOtp = document.getElementById("checkout-boton-reenviar-otp");

    if (temporizadorReenvioOtp) {
        clearInterval(temporizadorReenvioOtp);
    }

    botonReenviarOtp.style.display = "none";

    function actualizarTexto() {
        elementoContador.textContent = "Espere antes de solicitar " + segundosRestantes + " seg";
    }

    actualizarTexto();

    temporizadorReenvioOtp = setInterval(function () {
        segundosRestantes--;
        if (segundosRestantes <= 0) {
            clearInterval(temporizadorReenvioOtp);
            elementoContador.textContent = "";
            botonReenviarOtp.style.display = "block";
            return;
        }
        actualizarTexto();
    }, 1000);
}

/* ------------------------------------------------------------
   REENVIAR CÓDIGO: reutiliza los mismos datos del formulario
   (siguen en el DOM aunque el formulario esté oculto) y vuelve
   a pedir un OTP nuevo al backend.
------------------------------------------------------------- */
async function reenviarCodigoOtp() {
    const botonReenviarOtp = document.getElementById("checkout-boton-reenviar-otp");
    const elementoErrorOtp = document.getElementById("checkout-otp-mensaje-error");

    botonReenviarOtp.disabled = true;
    botonReenviarOtp.textContent = "Solicitando...";
    elementoErrorOtp.textContent = "";

    // Limpiamos las casillas para que el usuario no confunda el código viejo con el nuevo
    document.querySelectorAll("#checkout-otp-casillas input").forEach(function (casilla) {
        casilla.value = "";
    });

    try {
        await solicitarCodigoOtp(); // ya arma el request y, si sale bien, reinicia el contador
    } finally {
        botonReenviarOtp.disabled = false;
        botonReenviarOtp.textContent = "Solicitar nuevo código";
    }
}
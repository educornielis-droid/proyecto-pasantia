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
    const elementoError = document.getElementById("checkout-mensaje-error");
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

function iniciarContadorReenvio(segundosIniciales) {
    let segundosRestantes = segundosIniciales;
    const elementoContador = document.getElementById("checkout-otp-contador");

    if (temporizadorReenvioOtp) {
        clearInterval(temporizadorReenvioOtp);
    }

    function actualizarTexto() {
        elementoContador.textContent = "Espere antes de solicitar " + segundosRestantes + " seg";
    }

    actualizarTexto();

    temporizadorReenvioOtp = setInterval(function () {
        segundosRestantes--;
        if (segundosRestantes <= 0) {
            clearInterval(temporizadorReenvioOtp);
            elementoContador.textContent = "Ya puedes solicitar un nuevo código.";
            return;
        }
        actualizarTexto();
    }, 1000);
}
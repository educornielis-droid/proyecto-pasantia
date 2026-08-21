/* ============================================================
   TRANSACCIONES.HTML - Lógica de reembolso
   ============================================================ */

console.log("[Transacciones] transacciones.js se cargó correctamente.");

const MAXIMO_INTENTOS_POLLING_REEMBOLSO = 60; // 60 x 2.5s = 150 segundos como techo

document.addEventListener("DOMContentLoaded", function () {
    document.querySelectorAll(".boton-reembolso").forEach(function (boton) {
        boton.addEventListener("click", function () {
            confirmarYReembolsar(boton);
        });
    });

    // Si la página se recarga mientras un reembolso seguía en PEND/PROC,
    // retomamos la verificación automáticamente en vez de dejarlo colgado.
    document.querySelectorAll(".boton-reembolso-procesando").forEach(function (elemento) {
        const idTransaccion = elemento.dataset.transaccionId;
        if (idTransaccion) {
            pollingEstadoReembolso(idTransaccion, elemento, 0);
        }
    });
});

function confirmarYReembolsar(boton) {
    const idTransaccion = boton.dataset.transaccionId;

    if (typeof Swal === "undefined") {
        if (confirm("¿Confirmas el reembolso de esta transacción? El dinero se devuelve y el stock se repone. No se puede deshacer.")) {
            iniciarReembolso(idTransaccion, boton);
        }
        return;
    }

    Swal.fire({
        title: "¿Confirmar reembolso?",
        text: "Esta acción devuelve el dinero al cliente y repone el stock del producto. No se puede deshacer.",
        icon: "warning",
        showCancelButton: true,
        confirmButtonText: "Sí, reembolsar",
        cancelButtonText: "Cancelar",
        confirmButtonColor: "rgb(0, 101, 187)",
        cancelButtonColor: "#747B8F"
    }).then(function (resultado) {
        if (resultado.isConfirmed) {
            iniciarReembolso(idTransaccion, boton);
        }
    });
}

async function iniciarReembolso(idTransaccion, boton) {
    ponerBotonEnProcesando(boton);

    try {
        const respuesta = await fetch("/admin/reembolso/" + idTransaccion, {
            method: "POST"
        });

        if (!respuesta.ok) {
            const detalleError = await respuesta.json().catch(function () { return {}; });
            throw new Error(detalleError.error || "No se pudo iniciar el reembolso.");
        }

        console.log("[Transacciones] Reembolso enviado a Sypago, verificando estado...");
        pollingEstadoReembolso(idTransaccion, boton, 0);
    } catch (error) {
        console.error("[Transacciones] Error al iniciar reembolso:", error);
        mostrarErrorReembolso(boton, error.message);
    }
}

function pollingEstadoReembolso(idTransaccion, boton, intentos) {
    if (intentos > MAXIMO_INTENTOS_POLLING_REEMBOLSO) {
        mostrarErrorReembolso(boton, "El reembolso está tardando más de lo esperado. Revísalo de nuevo en unos minutos.");
        return;
    }

    fetch("/api/checkout/" + idTransaccion + "/estado?contexto=reembolso")
        .then(function (respuesta) {
            return respuesta.json().then(function (datos) {
                return { ok: respuesta.ok, datos: datos };
            });
        })
        .then(function (resultado) {
            if (!resultado.ok) {
                setTimeout(function () {
                    pollingEstadoReembolso(idTransaccion, boton, intentos + 1);
                }, 2500);
                return;
            }

            const estado = resultado.datos.estado;
            console.log("[Transacciones] Estado del reembolso:", estado, resultado.datos);

            if (estado === "ACCP") {
                marcarComoReembolsado(boton);
                mostrarExitoReembolso();
            } else if (estado === "RJCT" || estado === "CANC") {
                const motivo = resultado.datos.descripcion || "El banco rechazó el reembolso.";
                mostrarErrorReembolso(boton, motivo);
            } else {
                setTimeout(function () {
                    pollingEstadoReembolso(idTransaccion, boton, intentos + 1);
                }, 2500);
            }
        })
        .catch(function (error) {
            console.error("[Transacciones] Error de red al consultar estado del reembolso:", error);
            setTimeout(function () {
                pollingEstadoReembolso(idTransaccion, boton, intentos + 1);
            }, 2500);
        });
}

function ponerBotonEnProcesando(boton) {
    boton.outerHTML = '<span class="boton-reembolso boton-reembolso-procesando" data-transaccion-id="' +
        boton.dataset.transaccionId + '"><i class="fa-solid fa-spinner"></i> Procesando...</span>';
}

function marcarComoReembolsado(elementoActual) {
    const idTransaccion = elementoActual.dataset.transaccionId;
    const elementoVigente = document.querySelector('[data-transaccion-id="' + idTransaccion + '"]');
    const objetivo = elementoVigente || elementoActual;

    objetivo.outerHTML = '<span class="boton-reembolso boton-reembolsado"><i class="fa-solid fa-check"></i> Reembolsado</span>';
}

function mostrarErrorReembolso(elementoActual, mensaje) {
    const idTransaccion = elementoActual.dataset.transaccionId;
    const elementoVigente = document.querySelector('[data-transaccion-id="' + idTransaccion + '"]');
    const objetivo = elementoVigente || elementoActual;

    const nuevoBoton = document.createElement("button");
    nuevoBoton.type = "button";
    nuevoBoton.className = "boton-reembolso";
    nuevoBoton.dataset.transaccionId = idTransaccion;
    nuevoBoton.innerHTML = '<i class="fa-solid fa-rotate-left"></i> Reintentar reembolso';
    nuevoBoton.addEventListener("click", function () {
        confirmarYReembolsar(nuevoBoton);
    });

    objetivo.replaceWith(nuevoBoton);

    if (typeof Swal !== "undefined") {
        Swal.fire({
            icon: "error",
            title: "No se pudo reembolsar",
            text: mensaje
        });
    } else {
        alert("No se pudo reembolsar: " + mensaje);
    }
}

function mostrarExitoReembolso() {
    if (typeof Swal !== "undefined") {
        Swal.fire({
            icon: "success",
            title: "Reembolso exitoso",
            text: "El dinero fue devuelto al cliente y el stock del producto se repuso correctamente."
        });
    } else {
        alert("Reembolso exitoso: el dinero fue devuelto y el stock se repuso.");
    }
}
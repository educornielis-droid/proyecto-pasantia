// document.getElementById('formAdmin').addEventListener('submit', function (e) {
//     e.preventDefault();
    
//     const correo = document.getElementById('correoAdmin').value;
//     const password = document.getElementById('contrasenaAdmin').value;
//     const errorMensaje = document.getElementById('errorMensaje');

//     if (correo == 'admin_test_2026@midominio.local' && password == 'X9#pL0_kZ81') {
//         window.location.href = '/app/admin/ordenes';
//     } else {
//         errorMensaje.style.display = 'block';
//     }
// });

document.getElementById('formAdmin').addEventListener('submit', async (e) => {
    e.preventDefault();

    const correo = document.getElementById('correoAdmin').value;
    const contrasena = document.getElementById('contrasenaAdmin').value;

    try {

        const response = await fetch('/app/admin/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ correo, contrasena })
        });

        const data = await response.json();

        if (!response.ok) {
            Swal.fire({
                icon: 'error',
                title: 'Acceso Denegado',
                text: data.message,
                confirmButtonColor: '#d33'
            });
            return;
        }

        window.location.href = '/app/admin/ordenes';

    } catch (error) {
        Swal.fire({
            icon: 'error',
            title: 'Error de red',
            text: 'Ocurrió un fallo al comunicarse con el servidor'
        });
    };

});
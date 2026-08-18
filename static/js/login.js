const container = document.getElementById('containerLogin');
const registerBtn = document.getElementById('register');
const loginBtn = document.getElementById('login');

// Cambiar entre páneles
registerBtn.addEventListener('click', () => container.classList.add("active"));
loginBtn.addEventListener('click', () => container.classList.remove("active"));

// --- MOSTRAR / OCULTAR CONTRASEÑA ---
document.querySelectorAll('.toggle-password').forEach(icon => {
    icon.addEventListener('click', () => {
        const targetId = icon.getAttribute('data-target');
        const input = document.getElementById(targetId);
        
        if (input.type === 'password') {
            input.type = 'text';
            icon.classList.remove('fa-eye');
            icon.classList.add('fa-eye-slash');
        } else {
            input.type = 'password';
            icon.classList.remove('fa-eye-slash');
            icon.classList.add('fa-eye');
        }
    });
});


// --- FUNCIONES DE VALIDACIÓN ---
const showError = (input, message) => {
    input.classList.add('invalid');
    // const inputPlaceholder = input.nextElementSibling.classList.remove('placeholder')
    const errorSpan = input.nextElementSibling.classList.contains('error-msg') 
        ? input.nextElementSibling 
        : input.parentElement.querySelector('.error-msg');
    
    if (errorSpan) errorSpan.textContent = message;
};

const clearError = (input) => {
    input.classList.remove('invalid');
    const errorSpan = input.nextElementSibling.classList.contains('error-msg') 
        ? input.nextElementSibling 
        : input.parentElement.querySelector('.error-msg');
    
    if (errorSpan) errorSpan.textContent = '';
};

// Expresiones Regulares
const isOnlyLetters = (str) => /^[a-zA-ZáéíóúÁÉÍÓÚñÑ\s]+$/.test(str.trim());
const isValidEmail = (email) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim());

// Contraseña: 4-12 caracteres, >=1 mayúscula, >=1 minúscula, >=1 número, >=1 carácter especial
const isValidPassword = (pass) => {
    const passRegex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]).{4,12}$/;
    return passRegex.test(pass);
};

// Limpiar errores al escribir
document.querySelectorAll('.containerLogin input').forEach(input => {
    input.addEventListener('input', () => clearError(input));
});

// --- FORMULARIO DE REGISTRO ---
const signupForm = document.getElementById('signup-form');
signupForm.addEventListener('submit', (e) => {
    e.preventDefault();
    let isValid = true;

    const name = document.getElementById('signup-name');
    const lastname = document.getElementById('signup-lastname');
    const email = document.getElementById('signup-email');
    const password = document.getElementById('signup-password');

    // Validar Nombre
    if (name.value.trim() === '') {
        showError(name, 'El nombre es obligatorio.');
        isValid = false;
    } else if (!isOnlyLetters(name.value)) {
        showError(name, 'Solo se permiten letras.');
        isValid = false;
    } else { clearError(name); }

    // Validar Apellido
    if (lastname.value.trim() === '') {
        showError(lastname, 'El apellido es obligatorio.');
        isValid = false;
    } else if (!isOnlyLetters(lastname.value)) {
        showError(lastname, 'Solo se permiten letras.');
        isValid = false;
    } else { clearError(lastname); }

    // Validar Correo
    if (!isValidEmail(email.value)) {
        showError(email, 'Ingresa un correo válido.');
        isValid = false;
    } else { clearError(email); }

    // Validar Contraseña
    if (!isValidPassword(password.value)) {
        showError(password, '4-12 carac, incl. Mayús, Minús, Num y Especial.');
        isValid = false;
    } else { clearError(password); }

    // Guardado y redirección
    if (isValid) {
        const userData = {
            nombre: name.value.trim(),
            apellido: lastname.value.trim(),
            email: email.value.trim(),
            password: password.value // Nota: En producción esto debe ir encriptado/hash en BD
        };

        // Guardar sesión y perfil
        sessionStorage.setItem('userProfile', JSON.stringify(userData));
        sessionStorage.setItem('isLoggedIn', 'true');

        window.location.href = '/app/productos';
    }
});

// --- FORMULARIO DE INICIAR SESIÓN ---
const loginForm = document.getElementById('login-form');
loginForm.addEventListener('submit', (e) => {
    e.preventDefault();
    let isValid = true;

    const email = document.getElementById('login-email');
    const password = document.getElementById('login-password');

    if (!isValidEmail(email.value)) {
        showError(email, 'Ingresa un correo válido.');
        isValid = false;
    } else { clearError(email); }

    if (password.value.trim() === '') {
        showError(password, 'Ingresa tu contraseña.');
        isValid = false;
    } else { clearError(password); }

    if (isValid) {
        const storedUser = JSON.parse(sessionStorage.getItem('userProfile'));

        if (storedUser && storedUser.email === email.value.trim() && storedUser.password === password.value) {
            sessionStorage.setItem('isLoggedIn', 'true');
            window.location.href = '/app/productos';
        } else if (storedUser) {
            showError(password, 'Correo o contraseña incorrectos.');
        } else {
            // Si no hay registro en memoria, guardamos las credenciales actuales e ingresamos
            const userData = {
                nombre: 'Usuario',
                apellido: '',
                email: email.value.trim()
            };
            sessionStorage.setItem('userProfile', JSON.stringify(userData));
            sessionStorage.setItem('isLoggedIn', 'true');
            window.location.href = '/app/productos';
        }
    }
});
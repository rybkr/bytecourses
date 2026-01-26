import { $ } from "../core/dom.js";
import { showError, hideError } from "../core/utils.js";
import api from "../core/api.js";

/**
 * Generic auth form handler that reduces duplication across login, register,
 * forgot password, and reset password forms.
 */
function createAuthFormHandler(config) {
    const form = $(config.formSelector);
    if (!form) return null;

    const errorDiv = $(config.errorContainer || "#error-message");
    const successDiv = config.successContainer ? $(config.successContainer) : null;

    form.addEventListener("submit", async (e) => {
        e.preventDefault();
        hideError(errorDiv);

        if (successDiv) {
            successDiv.classList.add("hidden");
        }

        const submitBtn = form.querySelector('button[type="submit"]');
        if (submitBtn) submitBtn.disabled = true;

        try {
            // Get form data
            const data = config.getFormData(form);

            // Validate if needed
            if (config.validate) {
                const validationError = config.validate(data, form);
                if (validationError) {
                    showError(validationError, errorDiv);
                    return;
                }
            }

            const response = await api.post(config.endpoint, data);

            // Check for success based on config
            const isSuccess = config.isSuccess
                ? config.isSuccess(response)
                : response && response.ok;

            if (isSuccess) {
                if (config.onSuccess) {
                    config.onSuccess(response, form, successDiv);
                } else if (config.successRedirect) {
                    window.location.href = config.successRedirect;
                }
            } else {
                showError(config.defaultError || "An error occurred. Please try again.", errorDiv);
            }
        } catch (error) {
            showError(error.message || config.defaultError || "An error occurred. Please try again.", errorDiv);
        } finally {
            if (submitBtn) submitBtn.disabled = false;
        }
    });

    return form;
}

function validateNextUrl(next) {
    if (!next || typeof next !== "string") return "/";
    const s = next.trim();
    if (s === "" || !s.startsWith("/") || s.startsWith("//")) return "/";
    if (s.toLowerCase().startsWith("javascript:")) return "/";
    if (s === "/login" || s === "/register") return "/";
    if (s.startsWith("/login?") || s.startsWith("/register?")) return "/";
    return s;
}

export function initLoginForm() {
    createAuthFormHandler({
        formSelector: "#loginForm",
        endpoint: "/api/login",
        errorContainer: "#error-message",
        defaultError: "Invalid credentials",
        getFormData: (form) => ({
            email: form.email.value,
            password: form.password.value,
        }),
        onSuccess: () => {
            const params = new URLSearchParams(window.location.search);
            const nextUrl = validateNextUrl(params.get("next"));
            window.location.href = nextUrl;
        },
    });
}

export function initRegisterForm() {
    createAuthFormHandler({
        formSelector: "#registerForm",
        endpoint: "/api/register",
        errorContainer: "#error-message",
        defaultError: "Registration failed",
        getFormData: (form) => ({
            name: form.name.value,
            email: form.email.value,
            password: form.password.value,
        }),
        successRedirect: "/login",
    });
}

export function initForgotPasswordForm() {
    createAuthFormHandler({
        formSelector: "#forgotPasswordForm",
        endpoint: "/api/password-reset/request",
        errorContainer: "#error-message",
        successContainer: "#success-message",
        defaultError: "An error occurred. Please try again.",
        getFormData: (form) => ({
            email: form.email.value,
        }),
        isSuccess: (response) => response && response.status === 202,
        onSuccess: (response, form, successDiv) => {
            form.style.display = "none";
            if (successDiv) {
                successDiv.textContent =
                    "If an account exists with that email, you will receive password reset instructions shortly.";
                successDiv.classList.remove("hidden");
            }
        },
    });
}

export function initResetPasswordForm() {
    createAuthFormHandler({
        formSelector: "#resetPasswordForm",
        errorContainer: "#error-message",
        successContainer: "#success-message",
        defaultError: "Invalid or expired token. Please request a new password reset.",
        getFormData: (form) => {
            const params = new URLSearchParams(window.location.search);
            const token = params.get("token");
            return {
                new_password: form.password.value,
                _token: token,
                _confirmPassword: form.confirmPassword.value,
            };
        },
        get endpoint() {
            const params = new URLSearchParams(window.location.search);
            const token = params.get("token");
            return `/api/password-reset/confirm?token=${encodeURIComponent(token || "")}`;
        },
        validate: (data) => {
            if (!data._token) {
                return "Invalid reset link. Please request a new password reset.";
            }
            if (data.new_password !== data._confirmPassword) {
                return "Passwords do not match.";
            }
            return null;
        },
        isSuccess: (response) => response && response.status === 204,
        onSuccess: (response, form, successDiv) => {
            form.style.display = "none";
            if (successDiv) {
                successDiv.textContent =
                    "Password reset successfully! Redirecting to login...";
                successDiv.classList.remove("hidden");
            }
            setTimeout(() => {
                window.location.href = "/login";
            }, 2000);
        },
    });
}

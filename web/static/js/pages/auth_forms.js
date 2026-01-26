import FormSubmitHandler from "../components/FormSubmitHandler.js";
import { $ } from "../core/dom.js";
import { showError, hideError } from "../core/utils.js";

function initAuthForm(config) {
    const form = $(config.formSelector);
    if (!form) return;

    const errorDiv = $(config.errorContainer || "#error-message");
    const successDiv = config.successContainer
        ? $(config.successContainer)
        : null;

    const handler = new FormSubmitHandler(config.formSelector, {
        endpoint: config.endpoint,
        method: "POST",
        successRedirect: config.successRedirect,
        errorContainer: config.errorContainer || "#error-message",
        transformData: config.transformData || ((data) => data),
        onSuccess: async (response) => {
            if (config.onSuccess) {
                await config.onSuccess(response);
            } else if (config.successRedirect) {
                window.location.href = config.successRedirect;
            }
        },
        onError: (error) => {
            if (config.onError) {
                config.onError(error);
            }
        },
    });

    if (config.handleSuccessMessage) {
        const originalHandleSubmit = handler.handleSubmit.bind(handler);
        handler.handleSubmit = async function () {
            hideError(errorDiv);
            if (successDiv) {
                successDiv.classList.add("hidden");
            }

            try {
                const data = handler.getFormData();
                const response = await fetch(handler.options.endpoint, {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/json",
                    },
                    body: JSON.stringify(data),
                });

                if (response.status === 202 || response.ok) {
                    if (config.hideFormOnSuccess && form) {
                        form.style.display = "none";
                    }
                    if (successDiv && config.successMessage) {
                        successDiv.textContent = config.successMessage;
                        successDiv.classList.remove("hidden");
                    }
                    if (config.successRedirect) {
                        setTimeout(() => {
                            window.location.href = config.successRedirect;
                        }, config.redirectDelay || 2000);
                    }
                } else {
                    const errorText = await response.text();
                    showError(
                        errorText ||
                            config.defaultError ||
                            "An error occurred. Please try again.",
                        errorDiv,
                    );
                }
            } catch (error) {
                showError(
                    error.message ||
                        config.defaultError ||
                        "An error occurred. Please try again.",
                    errorDiv,
                );
            } finally {
                if (handler.submitButton) {
                    handler.submitButton.disabled = false;
                }
            }
        };
    }
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
    const form = $("#loginForm");
    if (!form) return;

    const errorDiv = $("#error-message");

    function getNextUrl() {
        const params = new URLSearchParams(window.location.search);
        return params.get("next");
    }

    form.addEventListener("submit", async (e) => {
        e.preventDefault();
        hideError(errorDiv);

        const email = form.email.value;
        const password = form.password.value;

        const submitBtn = form.querySelector('button[type="submit"]');
        if (submitBtn) {
            submitBtn.disabled = true;
        }

        try {
            const response = await fetch("/api/login", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({ email, password }),
            });

            if (response.ok) {
                const nextUrl = validateNextUrl(getNextUrl());
                window.location.href = nextUrl;
            } else {
                const errorText = await response.text();
                showError(errorText || "Invalid credentials", errorDiv);
            }
        } catch (error) {
            showError("An error occurred. Please try again.", errorDiv);
        } finally {
            if (submitBtn) {
                submitBtn.disabled = false;
            }
        }
    });
}

export function initRegisterForm() {
    const form = $("#registerForm");
    if (!form) return;

    const errorDiv = $("#error-message");

    form.addEventListener("submit", async (e) => {
        e.preventDefault();
        hideError(errorDiv);

        const name = form.name.value;
        const email = form.email.value;
        const password = form.password.value;

        const submitBtn = form.querySelector('button[type="submit"]');
        if (submitBtn) {
            submitBtn.disabled = true;
        }

        try {
            const response = await fetch("/api/register", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({ name, email, password }),
            });

            if (response.ok) {
                window.location.href = "/login";
            } else {
                const errorText = await response.text();
                showError(errorText || "Registration failed", errorDiv);
            }
        } catch (error) {
            showError("An error occurred. Please try again.", errorDiv);
        } finally {
            if (submitBtn) {
                submitBtn.disabled = false;
            }
        }
    });
}

export function initForgotPasswordForm() {
    const form = $("#forgotPasswordForm");
    if (!form) return;

    const errorDiv = $("#error-message");
    const successDiv = $("#success-message");

    form.addEventListener("submit", async (e) => {
        e.preventDefault();
        hideError(errorDiv);
        if (successDiv) {
            successDiv.classList.add("hidden");
        }

        const email = form.email.value;

        const submitBtn = form.querySelector('button[type="submit"]');
        if (submitBtn) {
            submitBtn.disabled = true;
        }

        try {
            const response = await fetch("/api/password-reset/request", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({ email }),
            });

            if (response.status === 202) {
                form.style.display = "none";
                if (successDiv) {
                    successDiv.textContent =
                        "If an account exists with that email, you will receive password reset instructions shortly.";
                    successDiv.classList.remove("hidden");
                }
            } else {
                const errorText = await response.text();
                showError(
                    errorText || "An error occurred. Please try again.",
                    errorDiv,
                );
            }
        } catch (error) {
            showError("An error occurred. Please try again.", errorDiv);
        } finally {
            if (submitBtn) {
                submitBtn.disabled = false;
            }
        }
    });
}

export function initResetPasswordForm() {
    const form = $("#resetPasswordForm");
    if (!form) return;

    const errorDiv = $("#error-message");
    const successDiv = $("#success-message");

    function getTokenFromURL() {
        const params = new URLSearchParams(window.location.search);
        return params.get("token");
    }

    form.addEventListener("submit", async (e) => {
        e.preventDefault();

        hideError(errorDiv);
        if (successDiv) {
            successDiv.style.display = "none";
        }

        const password = form.password.value;
        const confirmPassword = form.confirmPassword.value;

        const token = getTokenFromURL();
        if (!token) {
            showError(
                "Invalid reset link. Please request a new password reset.",
                errorDiv,
            );
            return;
        }

        if (password !== confirmPassword) {
            showError("Passwords do not match.", errorDiv);
            return;
        }

        const submitBtn = form.querySelector('button[type="submit"]');
        if (submitBtn) {
            submitBtn.disabled = true;
        }

        try {
            const response = await fetch(
                `/api/password-reset/confirm?token=${encodeURIComponent(token)}`,
                {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/json",
                    },
                    body: JSON.stringify({ new_password: password }),
                },
            );

            if (response.status === 204) {
                form.style.display = "none";
                if (successDiv) {
                    successDiv.textContent =
                        "Password reset successfully! Redirecting to login...";
                    successDiv.classList.remove("hidden");
                }
                setTimeout(() => {
                    window.location.href = "/login";
                }, 2000);
            } else {
                const errorText = await response.text();
                showError(
                    errorText ||
                        "Invalid or expired token. Please request a new password reset.",
                    errorDiv,
                );
            }
        } catch (error) {
            showError("An error occurred. Please try again.", errorDiv);
        } finally {
            if (submitBtn) {
                submitBtn.disabled = false;
            }
        }
    });
}

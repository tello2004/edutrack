import { apiClient, setToken, removeToken } from './client';
import type { LoginRequest, LoginResponse, LicenseLoginRequest, LicenseLoginResponse } from './types';
import { useAuthStore } from '../stores/authStore';

/**
 * Login with email and password
 */
export async function login(email: string, password: string): Promise<LoginResponse> {
  const payload: LoginRequest = { email, password };

  const response = await apiClient.post<LoginResponse>('/auth/login', payload);
  const { token, role, user } = response.data;

  // Store token
  setToken(token);

  // Update auth store
  useAuthStore.getState().setAuth(token, role, user);

  return response.data;
}

/**
 * Validate a license key and get tenant information
 */
export async function validateLicense(licenseKey: string): Promise<LicenseLoginResponse> {
  const payload: LicenseLoginRequest = { license_key: licenseKey };

  const response = await apiClient.post<LicenseLoginResponse>('/auth/license', payload);

  return response.data;
}

/**
 * Logout the current user
 */
export function logout(): void {
  removeToken();
  useAuthStore.getState().logout();
}

/**
 * Check if user is currently authenticated (has valid token in storage)
 */
export function isLoggedIn(): boolean {
  return useAuthStore.getState().isAuthenticated;
}

/**
 * Get current user information
 */
export function getCurrentUser() {
  return useAuthStore.getState().user;
}

/**
 * Get current user's role
 */
export function getCurrentRole() {
  return useAuthStore.getState().role;
}

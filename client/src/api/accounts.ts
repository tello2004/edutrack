import { apiClient } from './client';
import type {
  Account,
  CreateAccountRequest,
  UpdateAccountRequest,
} from './types';

const BASE_PATH = '/accounts';

/**
 * Get all accounts
 */
export async function getAccounts(): Promise<Account[]> {
  const response = await apiClient.get<Account[]>(BASE_PATH);
  return response.data;
}

/**
 * Get an account by ID
 */
export async function getAccount(id: number): Promise<Account> {
  const response = await apiClient.get<Account>(`${BASE_PATH}/${id}`);
  return response.data;
}

/**
 * Create a new account
 */
export async function createAccount(data: CreateAccountRequest): Promise<Account> {
  const response = await apiClient.post<Account>(BASE_PATH, data);
  return response.data;
}

/**
 * Update an existing account
 */
export async function updateAccount(id: number, data: UpdateAccountRequest): Promise<Account> {
  const response = await apiClient.put<Account>(`${BASE_PATH}/${id}`, data);
  return response.data;
}

/**
 * Delete an account
 */
export async function deleteAccount(id: number): Promise<void> {
  await apiClient.delete(`${BASE_PATH}/${id}`);
}

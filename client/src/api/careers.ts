import { apiClient } from './client';
import type {
  Career,
  CreateCareerRequest,
  UpdateCareerRequest,
} from './types';

const BASE_PATH = '/careers';

/**
 * Get all careers
 */
export async function getCareers(): Promise<Career[]> {
  const response = await apiClient.get<Career[]>(BASE_PATH);
  return response.data;
}

/**
 * Get a career by ID
 */
export async function getCareer(id: number): Promise<Career> {
  const response = await apiClient.get<Career>(`${BASE_PATH}/${id}`);
  return response.data;
}

/**
 * Create a new career
 */
export async function createCareer(data: CreateCareerRequest): Promise<Career> {
  const response = await apiClient.post<Career>(BASE_PATH, data);
  return response.data;
}

/**
 * Update an existing career
 */
export async function updateCareer(id: number, data: UpdateCareerRequest): Promise<Career> {
  const response = await apiClient.put<Career>(`${BASE_PATH}/${id}`, data);
  return response.data;
}

/**
 * Delete a career
 */
export async function deleteCareer(id: number): Promise<void> {
  await apiClient.delete(`${BASE_PATH}/${id}`);
}

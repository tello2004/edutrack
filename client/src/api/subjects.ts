import { apiClient } from './client';
import type {
  Subject,
  CreateSubjectRequest,
  UpdateSubjectRequest,
} from './types';

const BASE_PATH = '/subjects';

/**
 * Get all subjects
 */
export async function getSubjects(): Promise<Subject[]> {
  const response = await apiClient.get<Subject[]>(BASE_PATH);
  return response.data;
}

/**
 * Get a subject by ID
 */
export async function getSubject(id: number): Promise<Subject> {
  const response = await apiClient.get<Subject>(`${BASE_PATH}/${id}`);
  return response.data;
}

/**
 * Create a new subject
 */
export async function createSubject(data: CreateSubjectRequest): Promise<Subject> {
  const response = await apiClient.post<Subject>(BASE_PATH, data);
  return response.data;
}

/**
 * Update an existing subject
 */
export async function updateSubject(id: number, data: UpdateSubjectRequest): Promise<Subject> {
  const response = await apiClient.put<Subject>(`${BASE_PATH}/${id}`, data);
  return response.data;
}

/**
 * Delete a subject
 */
export async function deleteSubject(id: number): Promise<void> {
  await apiClient.delete(`${BASE_PATH}/${id}`);
}

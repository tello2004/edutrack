import { apiClient } from './client';
import type {
  Grade,
  CreateGradeRequest,
  UpdateGradeRequest,
} from './types';

const BASE_PATH = '/grades';

export interface ListGradesParams {
  student_id?: number;
  subject_id?: number;
  teacher_id?: number;
}

/**
 * Get all grades with optional filters
 */
export async function getGrades(params?: ListGradesParams): Promise<Grade[]> {
  const response = await apiClient.get<Grade[]>(BASE_PATH, { params });
  return response.data;
}

/**
 * Get a grade by ID
 */
export async function getGrade(id: number): Promise<Grade> {
  const response = await apiClient.get<Grade>(`${BASE_PATH}/${id}`);
  return response.data;
}

/**
 * Create a new grade
 */
export async function createGrade(data: CreateGradeRequest): Promise<Grade> {
  const response = await apiClient.post<Grade>(BASE_PATH, data);
  return response.data;
}

/**
 * Update an existing grade
 */
export async function updateGrade(id: number, data: UpdateGradeRequest): Promise<Grade> {
  const response = await apiClient.put<Grade>(`${BASE_PATH}/${id}`, data);
  return response.data;
}

/**
 * Delete a grade
 */
export async function deleteGrade(id: number): Promise<void> {
  await apiClient.delete(`${BASE_PATH}/${id}`);
}

import { apiClient } from './client';
import type {
  Teacher,
  CreateTeacherRequest,
  UpdateTeacherRequest,
} from './types';

const BASE_PATH = '/teachers';

/**
 * Get all teachers
 */
export async function getTeachers(): Promise<Teacher[]> {
  const response = await apiClient.get<Teacher[]>(BASE_PATH);
  return response.data;
}

/**
 * Get a teacher by ID
 */
export async function getTeacher(id: number): Promise<Teacher> {
  const response = await apiClient.get<Teacher>(`${BASE_PATH}/${id}`);
  return response.data;
}

/**
 * Create a new teacher
 */
export async function createTeacher(data: CreateTeacherRequest): Promise<Teacher> {
  const response = await apiClient.post<Teacher>(BASE_PATH, data);
  return response.data;
}

/**
 * Update an existing teacher
 */
export async function updateTeacher(id: number, data: UpdateTeacherRequest): Promise<Teacher> {
  const response = await apiClient.put<Teacher>(`${BASE_PATH}/${id}`, data);
  return response.data;
}

/**
 * Delete a teacher
 */
export async function deleteTeacher(id: number): Promise<void> {
  await apiClient.delete(`${BASE_PATH}/${id}`);
}

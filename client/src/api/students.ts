import { apiClient } from './client';
import type {
  Student,
  CreateStudentRequest,
  UpdateStudentRequest,
} from './types';

export interface ListStudentsParams {
  career_id?: number;
  student_id?: string;
  name?: string;
}

export async function getStudents(params?: ListStudentsParams): Promise<Student[]> {
  const response = await apiClient.get<Student[]>('/students', { params });
  return response.data;
}

export async function getStudent(id: number): Promise<Student> {
  const response = await apiClient.get<Student>(`/students/${id}`);
  return response.data;
}

export async function createStudent(data: CreateStudentRequest): Promise<Student> {
  const response = await apiClient.post<Student>('/students', data);
  return response.data;
}

export async function updateStudent(id: number, data: UpdateStudentRequest): Promise<Student> {
  const response = await apiClient.put<Student>(`/students/${id}`, data);
  return response.data;
}

export async function deleteStudent(id: number): Promise<void> {
  await apiClient.delete(`/students/${id}`);
}

import { apiClient } from './client';
import type {
  Attendance,
  CreateAttendanceRequest,
  UpdateAttendanceRequest,
} from './types';

const BASE_PATH = '/attendances';

export interface ListAttendancesParams {
  student_id?: number;
  subject_id?: number;
  date?: string;
  status?: string;
}

/**
 * Get all attendances with optional filters
 */
export async function getAttendances(params?: ListAttendancesParams): Promise<Attendance[]> {
  const response = await apiClient.get<Attendance[]>(BASE_PATH, { params });
  return response.data;
}

/**
 * Get a single attendance record by ID
 */
export async function getAttendance(id: number): Promise<Attendance> {
  const response = await apiClient.get<Attendance>(`${BASE_PATH}/${id}`);
  return response.data;
}

/**
 * Create a new attendance record
 */
export async function createAttendance(data: CreateAttendanceRequest): Promise<Attendance> {
  const response = await apiClient.post<Attendance>(BASE_PATH, data);
  return response.data;
}

/**
 * Update an existing attendance record
 */
export async function updateAttendance(id: number, data: UpdateAttendanceRequest): Promise<Attendance> {
  const response = await apiClient.put<Attendance>(`${BASE_PATH}/${id}`, data);
  return response.data;
}

/**
 * Delete an attendance record
 */
export async function deleteAttendance(id: number): Promise<void> {
  await apiClient.delete(`${BASE_PATH}/${id}`);
}

/**
 * Create multiple attendance records at once (batch)
 */
export async function createAttendanceBatch(records: CreateAttendanceRequest[]): Promise<Attendance[]> {
  const results = await Promise.all(
    records.map((record) => createAttendance(record))
  );
  return results;
}

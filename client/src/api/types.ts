// ============================================
// Base types
// ============================================

export interface BaseModel {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt?: string | null;
}

// ============================================
// Auth types
// ============================================

export type Role = 'secretary' | 'teacher';

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  role: Role;
  user: {
    id: number;
    name: string;
    email: string;
  };
}

export interface LicenseLoginRequest {
  license_key: string;
}

export interface LicenseLoginResponse {
  tenant_id: string;
  tenant_name: string;
  message: string;
}

// ============================================
// Account types
// ============================================

export interface Account extends BaseModel {
  Name: string;
  Email: string;
  Role: Role;
  Active: boolean;
  TenantID: string;
  Tenant?: Tenant;
}

export interface CreateAccountRequest {
  name: string;
  email: string;
  password: string;
  role: Role;
}

export interface UpdateAccountRequest {
  name?: string;
  email?: string;
  password?: string;
  role?: Role;
  active?: boolean;
}

// ============================================
// Tenant types
// ============================================

export interface Tenant {
  ID: string;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt?: string | null;
  Name: string;
  LicenseID: number;
  License?: License;
}

export interface License extends BaseModel {
  Key: string;
  ExpiresAt: string;
  Active: boolean;
  MaxUsers: number;
}

// ============================================
// Career types
// ============================================

export interface Career extends BaseModel {
  Name: string;
  Code: string;
  Description: string;
  Duration: number;
  Active: boolean;
  TenantID: string;
  Subjects?: Subject[];
  Students?: Student[];
}

export interface CreateCareerRequest {
  name: string;
  code: string;
  description?: string;
  duration?: number;
}

export interface UpdateCareerRequest {
  name?: string;
  code?: string;
  description?: string;
  duration?: number;
  active?: boolean;
}

// ============================================
// Student types
// ============================================

export interface Student extends BaseModel {
  StudentID: string;
  TenantID: string;
  AccountID: number;
  Account?: Account;
  CareerID: number;
  Career?: Career;
}

export interface CreateStudentRequest {
  student_id: string;
  account_id: number;
  career_id?: number;
}

export interface UpdateStudentRequest {
  student_id?: string;
  career_id?: number;
}

// ============================================
// Teacher types
// ============================================

export interface Teacher extends BaseModel {
  TenantID: string;
  AccountID: number;
  Account?: Account;
  Subjects?: Subject[];
}

export interface CreateTeacherRequest {
  account_id: number;
}

export interface UpdateTeacherRequest {
  account_id?: number;
}

// ============================================
// Subject types
// ============================================

export interface Subject extends BaseModel {
  Name: string;
  Code: string;
  Description: string;
  Credits: number;
  TenantID: string;
  TeacherID?: number | null;
  Teacher?: Teacher | null;
  Careers?: Career[];
}

export interface CreateSubjectRequest {
  name: string;
  code: string;
  description?: string;
  credits?: number;
  teacher_id?: number;
  career_ids?: number[];
}

export interface UpdateSubjectRequest {
  name?: string;
  code?: string;
  description?: string;
  credits?: number;
  teacher_id?: number | null;
  career_ids?: number[];
}

// ============================================
// Attendance types
// ============================================

export type AttendanceStatus = 'present' | 'absent' | 'late' | 'excused';

export interface Attendance extends BaseModel {
  Date: string;
  Status: AttendanceStatus;
  Notes: string;
  StudentID: number;
  Student?: Student;
  SubjectID: number;
  Subject?: Subject;
  TenantID: string;
}

export interface CreateAttendanceRequest {
  date: string;
  status: AttendanceStatus;
  notes?: string;
  student_id: number;
  subject_id: number;
}

export interface UpdateAttendanceRequest {
  date?: string;
  status?: AttendanceStatus;
  notes?: string;
}

// ============================================
// Grade types
// ============================================

export interface Grade extends BaseModel {
  Value: number;
  Notes: string;
  StudentID: number;
  Student?: Student;
  SubjectID: number;
  Subject?: Subject;
  TeacherID: number;
  Teacher?: Teacher;
  TenantID: string;
}

export interface CreateGradeRequest {
  value: number;
  notes?: string;
  student_id: number;
  subject_id: number;
}

export interface UpdateGradeRequest {
  value?: number;
  notes?: string;
}

// ============================================
// API Error type
// ============================================

export interface ApiError {
  message: string;
}

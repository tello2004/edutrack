import {
  getStudents,
  getStudent,
  createStudent,
  updateStudent,
  deleteStudent,
  createAccount,
  type Student,
  type ListStudentsParams,
} from "../api";

// Legacy interface for backward compatibility with existing components
export interface Alumno {
  id: string;
  nombre: string;
  matricula: string;
  grupo: string;
  email?: string;
  careerId?: number;
  accountId?: number;
}

// Transform backend Student to frontend Alumno format
function studentToAlumno(student: Student): Alumno {
  return {
    id: student.ID.toString(),
    nombre: student.Account?.Name || "",
    matricula: student.StudentID,
    grupo: student.Career?.Name || "",
    email: student.Account?.Email,
    careerId: student.CareerID,
    accountId: student.AccountID,
  };
}

/**
 * Get all students (alumnos)
 */
export async function getAlumnos(
  params?: ListStudentsParams,
): Promise<Alumno[]> {
  const students = await getStudents(params);
  return students.map(studentToAlumno);
}

/**
 * Get a single student by ID
 */
export async function getAlumno(id: string): Promise<Alumno> {
  const student = await getStudent(parseInt(id, 10));
  return studentToAlumno(student);
}

/**
 * Add a new student
 * This creates both an account and a student record
 */
export async function addAlumno(alumno: Omit<Alumno, "id">): Promise<Alumno> {
  // First, create the account for the student
  const account = await createAccount({
    name: alumno.nombre,
    email: alumno.email || `${alumno.matricula}@estudiante.edutrack.com`,
    password: alumno.matricula, // Default password is the student ID
    role: "teacher", // Students use teacher role in the current model
  });

  // Then create the student record
  const student = await createStudent({
    student_id: alumno.matricula,
    account_id: account.ID,
    career_id: alumno.careerId,
  });

  return {
    id: student.ID.toString(),
    nombre: account.Name,
    matricula: student.StudentID,
    grupo: alumno.grupo,
    email: account.Email,
    careerId: student.CareerID,
    accountId: student.AccountID,
  };
}

/**
 * Update an existing student
 */
export async function updateAlumno(alumno: Alumno): Promise<Alumno> {
  const student = await updateStudent(parseInt(alumno.id, 10), {
    student_id: alumno.matricula,
    career_id: alumno.careerId,
  });

  return studentToAlumno(student);
}

/**
 * Delete a student by ID
 */
export async function deleteAlumno(id: string): Promise<void> {
  await deleteStudent(parseInt(id, 10));
}

/**
 * Search students by name or matricula
 */
export async function searchAlumnos(query: string): Promise<Alumno[]> {
  // Try searching by name first
  const byName = await getStudents({ name: query });
  if (byName.length > 0) {
    return byName.map(studentToAlumno);
  }

  // Try searching by student ID (matricula)
  const byMatricula = await getStudents({ student_id: query });
  return byMatricula.map(studentToAlumno);
}

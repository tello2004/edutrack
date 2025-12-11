import {
  getGrades,
  getGrade,
  createGrade,
  updateGrade,
  deleteGrade,
  type Grade,
  type ListGradesParams,
} from "../api";

// Legacy interface for backward compatibility with existing components
export interface Calificacion {
  id: string;
  alumno: string;
  matricula: string;
  evaluacion: string;
  calificacion: number;
  studentId?: number;
  subjectId?: number;
  teacherId?: number;
  notas?: string;
}

// Transform backend Grade to frontend Calificacion format
function gradeToCalificacion(grade: Grade): Calificacion {
  return {
    id: grade.ID.toString(),
    alumno: grade.Student?.Account?.Name || "",
    matricula: grade.Student?.StudentID || "",
    evaluacion: grade.Subject?.Name || "",
    calificacion: grade.Value,
    studentId: grade.StudentID,
    subjectId: grade.SubjectID,
    teacherId: grade.TeacherID,
    notas: grade.Notes,
  };
}

/**
 * Get all grades (calificaciones)
 */
export async function getCalificaciones(
  params?: ListGradesParams,
): Promise<Calificacion[]> {
  const grades = await getGrades(params);
  return grades.map(gradeToCalificacion);
}

/**
 * Get a single grade by ID
 */
export async function getCalificacion(id: string): Promise<Calificacion> {
  const grade = await getGrade(parseInt(id, 10));
  return gradeToCalificacion(grade);
}

/**
 * Get grades for a specific student
 */
export async function getCalificacionesByAlumno(
  studentId: number,
): Promise<Calificacion[]> {
  const grades = await getGrades({ student_id: studentId });
  return grades.map(gradeToCalificacion);
}

/**
 * Get grades for a specific subject
 */
export async function getCalificacionesByMateria(
  subjectId: number,
): Promise<Calificacion[]> {
  const grades = await getGrades({ subject_id: subjectId });
  return grades.map(gradeToCalificacion);
}

/**
 * Add a new grade
 */
export async function addCalificacion(
  calificacion: Omit<
    Calificacion,
    "id" | "alumno" | "matricula" | "evaluacion"
  >,
): Promise<Calificacion> {
  if (!calificacion.studentId || !calificacion.subjectId) {
    throw new Error("Student ID and Subject ID are required");
  }

  const grade = await createGrade({
    value: calificacion.calificacion,
    notes: calificacion.notas,
    student_id: calificacion.studentId,
    subject_id: calificacion.subjectId,
  });

  return gradeToCalificacion(grade);
}

/**
 * Update an existing grade
 */
export async function updateCalificacion(
  calificacion: Calificacion,
): Promise<Calificacion> {
  const grade = await updateGrade(parseInt(calificacion.id, 10), {
    value: calificacion.calificacion,
    notes: calificacion.notas,
  });

  return gradeToCalificacion(grade);
}

/**
 * Delete a grade by ID
 */
export async function deleteCalificacion(id: string): Promise<void> {
  await deleteGrade(parseInt(id, 10));
}

/**
 * Calculate average grade for a student
 */
export async function calcularPromedioAlumno(
  studentId: number,
): Promise<number> {
  const grades = await getGrades({ student_id: studentId });
  if (grades.length === 0) return 0;

  const sum = grades.reduce((acc, grade) => acc + grade.Value, 0);
  return sum / grades.length;
}

/**
 * Calculate average grade for a subject
 */
export async function calcularPromedioMateria(
  subjectId: number,
): Promise<number> {
  const grades = await getGrades({ subject_id: subjectId });
  if (grades.length === 0) return 0;

  const sum = grades.reduce((acc, grade) => acc + grade.Value, 0);
  return sum / grades.length;
}

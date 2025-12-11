import {
  getSubjects,
  getSubject,
  createSubject,
  updateSubject,
  deleteSubject,
  type Subject,
} from "../api";

// Legacy interface for backward compatibility with existing components
export interface Evaluacion {
  id: string;
  nombre: string;
  descripcion?: string;
  fecha: string;
  peso: number;
  codigo?: string;
  creditos?: number;
  teacherId?: number;
}

// Transform backend Subject to frontend Evaluacion format
function subjectToEvaluacion(subject: Subject): Evaluacion {
  return {
    id: subject.ID.toString(),
    nombre: subject.Name,
    descripcion: subject.Description,
    fecha: subject.CreatedAt.split("T")[0], // Use creation date as default
    peso: subject.Credits * 10, // Convert credits to weight percentage
    codigo: subject.Code,
    creditos: subject.Credits,
    teacherId: subject.TeacherID || undefined,
  };
}

/**
 * Get all evaluations (subjects)
 */
export async function getEvaluaciones(): Promise<Evaluacion[]> {
  const subjects = await getSubjects();
  return subjects.map(subjectToEvaluacion);
}

/**
 * Get a single evaluation by ID
 */
export async function getEvaluacion(id: string): Promise<Evaluacion> {
  const subject = await getSubject(parseInt(id, 10));
  return subjectToEvaluacion(subject);
}

/**
 * Add a new evaluation
 */
export async function addEvaluacion(
  evaluacion: Omit<Evaluacion, "id">,
): Promise<Evaluacion> {
  const subject = await createSubject({
    name: evaluacion.nombre,
    code:
      evaluacion.codigo || evaluacion.nombre.toUpperCase().replace(/\s+/g, "-"),
    description: evaluacion.descripcion,
    credits: evaluacion.creditos || Math.round(evaluacion.peso / 10),
    teacher_id: evaluacion.teacherId,
  });

  return subjectToEvaluacion(subject);
}

/**
 * Update an existing evaluation
 */
export async function updateEvaluacion(
  evaluacion: Evaluacion,
): Promise<Evaluacion> {
  const subject = await updateSubject(parseInt(evaluacion.id, 10), {
    name: evaluacion.nombre,
    code: evaluacion.codigo,
    description: evaluacion.descripcion,
    credits: evaluacion.creditos || Math.round(evaluacion.peso / 10),
    teacher_id: evaluacion.teacherId,
  });

  return subjectToEvaluacion(subject);
}

/**
 * Delete an evaluation by ID
 */
export async function deleteEvaluacion(id: string): Promise<void> {
  await deleteSubject(parseInt(id, 10));
}

/**
 * Get evaluations for a specific teacher
 */
export async function getEvaluacionesByTeacher(
  teacherId: number,
): Promise<Evaluacion[]> {
  const subjects = await getSubjects();
  return subjects
    .filter((subject) => subject.TeacherID === teacherId)
    .map(subjectToEvaluacion);
}

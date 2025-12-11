import {
  getAttendances,
  getAttendance,
  createAttendance,
  updateAttendance,
  deleteAttendance,
  type Attendance,
  type AttendanceStatus,
  type ListAttendancesParams,
} from "../api";

// Legacy interface for backward compatibility with existing components
export interface Asistencia {
  id: string;
  alumno: string;
  matricula: string;
  grupo: string;
  fecha: string;
  presente: boolean;
  status?: AttendanceStatus;
  notas?: string;
  studentId?: number;
  subjectId?: number;
}

// Transform backend Attendance to frontend Asistencia format
function attendanceToAsistencia(attendance: Attendance): Asistencia {
  return {
    id: attendance.ID.toString(),
    alumno: attendance.Student?.Account?.Name || "",
    matricula: attendance.Student?.StudentID || "",
    grupo: attendance.Subject?.Name || "",
    fecha: attendance.Date.split("T")[0], // Extract date part
    presente: attendance.Status === "present",
    status: attendance.Status,
    notas: attendance.Notes,
    studentId: attendance.StudentID,
    subjectId: attendance.SubjectID,
  };
}

// Transform frontend status to backend format
function parseStatus(presente: boolean): AttendanceStatus {
  return presente ? "present" : "absent";
}

/**
 * Get all attendances
 */
export async function getAsistencias(
  params?: ListAttendancesParams,
): Promise<Asistencia[]> {
  const attendances = await getAttendances(params);
  return attendances.map(attendanceToAsistencia);
}

/**
 * Get a single attendance by ID
 */
export async function getAsistencia(id: string): Promise<Asistencia> {
  const attendance = await getAttendance(parseInt(id, 10));
  return attendanceToAsistencia(attendance);
}

/**
 * Get attendances filtered by student
 */
export async function getAsistenciasByAlumno(
  studentId: number,
): Promise<Asistencia[]> {
  const attendances = await getAttendances({ student_id: studentId });
  return attendances.map(attendanceToAsistencia);
}

/**
 * Get attendances filtered by date
 */
export async function getAsistenciasByFecha(
  fecha: string,
): Promise<Asistencia[]> {
  const attendances = await getAttendances({ date: fecha });
  return attendances.map(attendanceToAsistencia);
}

/**
 * Get attendances filtered by subject
 */
export async function getAsistenciasByMateria(
  subjectId: number,
): Promise<Asistencia[]> {
  const attendances = await getAttendances({ subject_id: subjectId });
  return attendances.map(attendanceToAsistencia);
}

/**
 * Add a new attendance record
 */
export async function addAsistencia(
  asistencia: Omit<Asistencia, "id" | "alumno" | "matricula" | "grupo">,
): Promise<Asistencia> {
  if (!asistencia.studentId || !asistencia.subjectId) {
    throw new Error("studentId and subjectId are required");
  }

  const attendance = await createAttendance({
    date: asistencia.fecha,
    status: asistencia.status || parseStatus(asistencia.presente),
    notes: asistencia.notas,
    student_id: asistencia.studentId,
    subject_id: asistencia.subjectId,
  });

  return attendanceToAsistencia(attendance);
}

/**
 * Update an existing attendance record
 */
export async function updateAsistencia(
  asistencia: Asistencia,
): Promise<Asistencia> {
  const attendance = await updateAttendance(parseInt(asistencia.id, 10), {
    date: asistencia.fecha,
    status: asistencia.status || parseStatus(asistencia.presente),
    notes: asistencia.notas,
  });

  return attendanceToAsistencia(attendance);
}

/**
 * Delete an attendance record
 */
export async function deleteAsistencia(id: string): Promise<void> {
  await deleteAttendance(parseInt(id, 10));
}

/**
 * Register attendance for multiple students at once
 */
export async function registrarAsistenciaGrupal(
  subjectId: number,
  fecha: string,
  registros: { studentId: number; presente: boolean; notas?: string }[],
): Promise<Asistencia[]> {
  const results = await Promise.all(
    registros.map((registro) =>
      createAttendance({
        date: fecha,
        status: parseStatus(registro.presente),
        notes: registro.notas,
        student_id: registro.studentId,
        subject_id: subjectId,
      }),
    ),
  );

  return results.map(attendanceToAsistencia);
}

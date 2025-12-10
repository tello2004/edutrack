export interface Asistencia {
  id: string;
  alumno: string;
  matricula: string;
  grupo: string;
  fecha: string;
  presente: boolean;
}

export async function getAsistencias(): Promise<Asistencia[]> {
  return [
    { id: '1', alumno: 'Juan Pérez', matricula: 'A001', grupo: '1A', fecha: '2025-12-10', presente: true },
    { id: '2', alumno: 'María López', matricula: 'A002', grupo: '1B', fecha: '2025-12-10', presente: false },
    { id: '2', alumno: 'Diego Nuñez', matricula: 'A007', grupo: '1A', fecha: '2025-12-09', presente: false },
    { id: '2', alumno: 'Fernando María', matricula: 'A006', grupo: '1B', fecha: '2025-12-03', presente: true },
    { id: '2', alumno: 'Carlos Alberto', matricula: 'A004', grupo: '1B', fecha: '2025-12-12', presente: true },
  ];
}

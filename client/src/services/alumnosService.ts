export interface Alumno {
  id: string;
  nombre: string;
  matricula: string;
  grupo: string;
  email?: string;
}

export async function getAlumnos(): Promise<Alumno[]> {
  return [
    { id: '1', nombre: 'Juan Pérez', matricula: 'A001', grupo: '1A', email: 'juan@mail.com' },
    { id: '2', nombre: 'María López', matricula: 'A002', grupo: '1B', email: 'maria@mail.com' },
    { id: '3', nombre: 'Carlos Ruiz', matricula: 'A003', grupo: '1A' },
  ];
}

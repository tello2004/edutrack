export interface Grupo {
  id: string;
  nombre: string;
  descripcion?: string;
}

export async function getGrupos(): Promise<Grupo[]> {
  return [
    { id: "1", nombre: "1A", descripcion: "Primer grupo A" },
    { id: "2", nombre: "1B", descripcion: "Primer grupo B" },
    { id: "3", nombre: "2A", descripcion: "Segundo grupo A" },
  ];
}

export async function addGrupo(grupo: Grupo): Promise<Grupo> {
  return grupo;
}

export async function updateGrupo(grupo: Grupo): Promise<Grupo> {
  return grupo;
}

/*export async function deleteGrupo(id: string): Promise<void> {
  return;
}*/
